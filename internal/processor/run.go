package processor

import (
	"context"
	"fmt"
	internalservice "github.com/coding-ia/renovate-controller/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v63/github"
	"log"
	"strconv"
)

type RenovateTaskFunc interface {
	CreateTask(ctx context.Context, installation *github.Installation)
}

type RenovateTask interface {
	CreateRenovateTasks(ctx context.Context) error
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
	Subnets        []string
	SecurityGroups []string
	TaskOptions    TaskCommandOptions

	ctx context.Context
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

func (r RenovateCommand) CreateRenovateTasks(ctx context.Context) error {
	var renovateTask RenovateTaskFunc
	renovateTask = r.RunOptions

	svc := internalservice.NewRenovateGitHubApplicationService(r.GitHubClient)
	err := svc.EnumerateInstallationRepositories(ctx, renovateTask.CreateTask)
	if err != nil {
		return fmt.Errorf("error while processing repositoriest: %v", err)
	}

	return nil
}

func Run(ctx context.Context, githubConfig *GitHubConfig, runConfig *RunCommandOptions) error {
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

	err = renovateTask.CreateRenovateTasks(ctx)
	if err != nil {
		return fmt.Errorf("error creating renovate tasks: %v", err)
	}

	return nil
}

func (r RunCommandOptions) CreateTask(ctx context.Context, installation *github.Installation) {
	installationID := strconv.FormatInt(installation.GetID(), 10)

	log.Printf("Creating renovate task for %s", *installation.Account.Login)

	config := internalservice.ECSConfig{
		Cluster:   r.ClusterName,
		Task:      r.TaskDefinition,
		Container: r.ContainerName,
		AWSVPCConfig: internalservice.ECSVPCConfig{
			Subnets:        r.Subnets,
			SecurityGroups: r.SecurityGroups,
			AssignPublicIP: r.AssignPublicIP,
		},
	}

	svc := internalservice.NewRenovateTaskService(config)

	taskConfig := internalservice.RunTaskConfig{
		ApplicationID:  r.TaskOptions.ApplicationID,
		InstallationID: installationID,
	}
	_, err := svc.RunTask(ctx, taskConfig)
	if err != nil {
		log.Printf("error running task: %v", err)
		return
	}
}
