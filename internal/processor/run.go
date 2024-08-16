package processor

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v55/github"
	"log"
	"renovate-controller/internal/service"
)

type RenovateTaskFunc interface {
	CreateTask(repository *github.Repository, installationToken string, endpoint string)
}

type RenovateTask interface {
	CreateRenovateTasks() error
}

type RunConfig struct {
	TaskDefinition string
	ClusterName    string
	ContainerName  string
	AssignPublicIP bool
}

type GitHubConfig struct {
	ApplicationID string
	PrivateKey    []byte
	Endpoint      string
}

type RenovateCommand struct {
	Config       *RunConfig
	GitHubClient *github.Client
}

func (r RenovateCommand) CreateRenovateTasks() error {
	var renovateTask RenovateTaskFunc
	renovateTask = r.Config

	svc := service.NewRenovateGitHubApplicationService(r.GitHubClient)
	err := svc.EnumerateInstallationRepositories(renovateTask.CreateTask)
	if err != nil {
		return fmt.Errorf("error while processing repositoriest: %v", err)
	}

	return nil
}

func Run(githubConfig *GitHubConfig, runConfig *RunConfig) error {
	parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM(githubConfig.PrivateKey)
	if err != nil {
		return err
	}

	tokenString, err := service.GenerateJWT(githubConfig.ApplicationID, parsedKey)
	if err != nil {
		return fmt.Errorf("error generating JWT: %v", err)
	}

	client, err := service.CreateClient(tokenString, githubConfig.Endpoint)
	if err != nil {
		return fmt.Errorf("error creating github client: %v", err)
	}

	var renovateTask RenovateTask
	renovateTask = &RenovateCommand{
		Config:       runConfig,
		GitHubClient: client,
	}

	err = renovateTask.CreateRenovateTasks()
	if err != nil {
		return fmt.Errorf("error creating renovate tasks: %v", err)
	}

	return nil
}

func (r RunConfig) CreateTask(repository *github.Repository, installationToken string, endpoint string) {
	repo := fmt.Sprintf("%s/%s", repository.GetOwner().GetLogin(), repository.GetName())

	log.Printf("Creating renovate task for %s", repo)

	config := service.ECSConfig{
		Cluster:   r.ClusterName,
		Task:      r.TaskDefinition,
		Container: r.ContainerName,
		PublicIP:  r.AssignPublicIP,
	}

	svc := service.NewRenovateTaskService(config)
	_, err := svc.RunTask(installationToken, repo, endpoint)
	if err != nil {
		log.Printf("error running task: %v", err)
		return
	}
}
