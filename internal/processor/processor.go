package processor

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v55/github"
	"renovate-controller/internal/aws_helper"
	"renovate-controller/internal/github_helper"
)

type RenovateTask interface {
	CreateTask(client *github.Client, installation *github.Installation, repo *github.Repository)
}

func RunRenovateTasks(applicationID string, privateKey []byte, endpoint string, config *RenovateTaskConfig) error {
	parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return err
	}

	tokenString, err := github_helper.GenerateJWT(applicationID, parsedKey)
	if err != nil {
		return fmt.Errorf("error generating JWT: %v", err)
	}

	client, err := github_helper.CreateClient(tokenString, endpoint)
	if err != nil {
		return fmt.Errorf("error creating github client: %v", err)
	}

	var renovateTask RenovateTask
	renovateTask = config

	err = github_helper.ProcessAllInstallationRepositories(client, renovateTask.CreateTask)
	if err != nil {
		return fmt.Errorf("error while processing repositoriest: %v", err)
	}

	return nil
}

func GenerateInstallationToken(applicationID string, privateKey []byte, endpoint string, owner string, repository string) (string, error) {
	parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return "", err
	}

	tokenString, err := github_helper.GenerateJWT(applicationID, parsedKey)
	if err != nil {
		return "", fmt.Errorf("error generating JWT: %v", err)
	}

	client, err := github_helper.CreateClient(tokenString, endpoint)
	if err != nil {
		return "", fmt.Errorf("error creating github client: %v", err)
	}

	repo := fmt.Sprintf("%s/%s", owner, repository)
	installation, err := github_helper.FilterInstallation(client, repo)
	if err != nil {
		return "", fmt.Errorf("error finding installation: %v", err)
	}

	if installation == nil {
		return "", fmt.Errorf("no installation found for %s", repo)
	}

	token, _, err := client.Apps.CreateInstallationToken(context.Background(), installation.GetID(), &github.InstallationTokenOptions{
		Repositories: []string{repository},
	})
	if err != nil {
		return "", fmt.Errorf("error creating installation token: %v", err)
	}

	return token.GetToken(), nil
}

type RenovateTaskConfig struct {
	TaskDefinition string
	ClusterName    string
	AssignPublicIP bool
}

func (r *RenovateTaskConfig) CreateTask(client *github.Client, installation *github.Installation, repo *github.Repository) {
	token, _, err := client.Apps.CreateInstallationToken(context.Background(), installation.GetID(), &github.InstallationTokenOptions{
		Repositories: []string{repo.GetName()},
	})
	if err != nil {
		fmt.Errorf("error creating installation token: %v", err)
		return
	}

	repository := fmt.Sprintf("%s/%s", repo.GetOwner().GetLogin(), repo.GetName())
	_ = aws_helper.RunTask(r.ClusterName, r.TaskDefinition, token.GetToken(), repository, r.AssignPublicIP)
}
