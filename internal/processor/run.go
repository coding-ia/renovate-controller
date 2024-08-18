package processor

import (
	"fmt"
	internalservice "github.com/coding-ia/renovate-controller/internal/service"
	"github.com/coding-ia/renovate-controller/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v63/github"
	"log"
	"strconv"
)

type RenovateTaskFunc interface {
	CreateTask(installation *github.Installation, repository *github.Repository)
}

type RenovateTask interface {
	CreateRenovateTasks() error
}

type TaskCommandOptions struct {
	ApplicationID string
	PEMAWSSecret  string
	Endpoint      string
}

type RunCommandOptions struct {
	TaskDefinition string
	ClusterName    string
	ContainerName  string
	AssignPublicIP bool
	TaskOptions    TaskCommandOptions
}

type GitHubConfig struct {
	ApplicationID string
	PrivateKey    []byte
	Endpoint      string
}

type RenovateCommand struct {
	RunOptions   *RunCommandOptions
	GitHubClient *github.Client
}

func (r RenovateCommand) CreateRenovateTasks() error {
	var renovateTask RenovateTaskFunc
	renovateTask = r.RunOptions

	svc := internalservice.NewRenovateGitHubApplicationService(r.GitHubClient)
	err := svc.EnumerateInstallationRepositories(renovateTask.CreateTask)
	if err != nil {
		return fmt.Errorf("error while processing repositoriest: %v", err)
	}

	return nil
}

func Run(githubConfig *GitHubConfig, runConfig *RunCommandOptions) error {
	parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM(githubConfig.PrivateKey)
	if err != nil {
		return err
	}

	tokenString, err := internalservice.GenerateJWT(githubConfig.ApplicationID, parsedKey)
	if err != nil {
		return fmt.Errorf("error generating JWT: %v", err)
	}

	client, err := internalservice.CreateClient(tokenString, githubConfig.Endpoint)
	if err != nil {
		return fmt.Errorf("error creating github client: %v", err)
	}

	var renovateTask RenovateTask
	renovateTask = &RenovateCommand{
		RunOptions:   runConfig,
		GitHubClient: client,
	}

	err = renovateTask.CreateRenovateTasks()
	if err != nil {
		return fmt.Errorf("error creating renovate tasks: %v", err)
	}

	return nil
}

func (r RunCommandOptions) CreateTask(installation *github.Installation, repository *github.Repository) {
	repo := fmt.Sprintf("%s/%s", repository.GetOwner().GetLogin(), repository.GetName())
	installationID := strconv.FormatInt(installation.GetID(), 10)

	log.Printf("Creating renovate task for %s", repo)

	config := service.ECSConfig{
		Cluster:   r.ClusterName,
		Task:      r.TaskDefinition,
		Container: r.ContainerName,
		PublicIP:  r.AssignPublicIP,
	}

	svc := service.NewRenovateTaskService(config)

	taskConfig := service.RunTaskConfig{
		ApplicationID:  r.TaskOptions.ApplicationID,
		Repository:     repo,
		InstallationID: installationID,
	}
	_, err := svc.RunTask(taskConfig)
	if err != nil {
		log.Printf("error running task: %v", err)
		return
	}
}
