package processor

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v55/github"
	"log"
	"os"
	"renovate-controller/internal/service"
	"renovate-controller/internal/store"
	"text/template"
)

type GenerateTaskFunc interface {
	GenerateConfig(repos []string, installationToken string, endpoint string)
}

type GenerateTask interface {
	GenerateConfig() error
}

type GenerateCommand struct {
	CommandOptions GenerateCommandOptions
	GitHubClient   *github.Client
}

type GenerateCommandOptions struct {
	InstallationID   int64
	TargetRepository string
	S3Bucket         string
	S3ConfigKey      string
	Output           string
}

type GenerateFuncCallback struct {
	Command GenerateCommand
}

func (g GenerateCommand) GenerateConfig() error {
	var generateTask GenerateTaskFunc
	generateTask = &GenerateFuncCallback{
		Command: g,
	}

	svc := service.NewRenovateGitHubApplicationService(g.GitHubClient)
	err := svc.ProcessInstallationRepository(g.CommandOptions.InstallationID, generateTask.GenerateConfig)
	if err != nil {
		return fmt.Errorf("error while processing repositoriest: %v", err)
	}

	return nil
}

func Generate(githubConfig *GitHubConfig, options GenerateCommandOptions) error {
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

	var renovateTask GenerateTask
	renovateTask = &GenerateCommand{
		CommandOptions: options,
		GitHubClient:   client,
	}

	err = renovateTask.GenerateConfig()
	if err != nil {
		return fmt.Errorf("error creating renovate tasks: %v", err)
	}

	return nil
}

type TemplateData struct {
	InstallationToken string
	Endpoint          string
	Repository        string
	Repositories      []string
}

func (g GenerateFuncCallback) GenerateConfig(repos []string, installationToken string, endpoint string) {
	config, err := store.GetS3Object(g.Command.CommandOptions.S3Bucket, g.Command.CommandOptions.S3ConfigKey)
	if err != nil {
		log.Printf("error getting SSM parameter: %v", err)
		return
	}
	_ = config

	tmpl, err := template.New("config").Parse(config)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		return
	}

	data := TemplateData{
		InstallationToken: installationToken,
		Endpoint:          endpoint,
		Repositories:      repos,
		Repository:        g.Command.CommandOptions.TargetRepository,
	}

	log.Printf("Template 'Endpoint' = '%s'", data.Endpoint)
	log.Printf("Template 'Repository' = '%s'", data.Repository)

	file, err := os.Create(g.Command.CommandOptions.Output)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		return
	}

	log.Printf("Template successfully created at '%s'", g.Command.CommandOptions.Output)
}
