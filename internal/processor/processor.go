package processor

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v55/github"
	"log"
	"renovate-controller/internal/aws_helper"
	"renovate-controller/internal/github_helper"
)

type RenovateTask interface {
	CreateTask(repository *github.Repository, installationToken string)
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

type RenovateTaskConfig struct {
	TaskDefinition string
	ClusterName    string
	AssignPublicIP bool
}

func (r *RenovateTaskConfig) CreateTask(repository *github.Repository, installationToken string) {
	repo := fmt.Sprintf("%s/%s", repository.GetOwner().GetLogin(), repository.GetName())

	log.Printf("Creating renovate task for %s", repo)
	_, err := aws_helper.RunTask(r.ClusterName, r.TaskDefinition, installationToken, repo, r.AssignPublicIP)
	if err != nil {
		log.Printf("error running task: %v", err)
		return
	}
}
