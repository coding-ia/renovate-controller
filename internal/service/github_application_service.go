package service

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
	"log"
	"net/url"
	"time"
)

type enumerateFunc func(*github.Installation, *github.Repository)
type processFunc func([]string, string, string)

type RenovateGitHubApplicationService interface {
	EnumerateInstallationRepositories(processor enumerateFunc)
	ProcessInstallationRepository(installationId int64, processor processFunc) error
}

type ApplicationService struct {
	ApplicationID string
	Client        *github.Client
}

func NewRenovateGitHubApplicationService(client *github.Client) *ApplicationService {
	return &ApplicationService{
		Client: client,
	}
}

func (a *ApplicationService) EnumerateInstallationRepositories(processor enumerateFunc) error {
	opts := &github.ListOptions{PerPage: 10}
	for {
		installations, resp, err := a.Client.Apps.ListInstallations(context.Background(), opts)

		if err != nil {
			return err
		}

		for _, installation := range installations {
			token, _, err := a.Client.Apps.CreateInstallationToken(context.Background(), installation.GetID(), nil)
			if err != nil {
				return err
			}

			log.Printf("Processing repositories for installation %d", installation.GetID())

			installationToken := token.GetToken()
			installationClient, _ := CreateClient(installationToken, a.Client.BaseURL.Host)

			repoOpts := &github.ListOptions{PerPage: 10}
			for {
				repos, repoResp, err := installationClient.Apps.ListRepos(context.Background(), repoOpts)
				if err != nil {
					return err
				}

				for _, repo := range repos.Repositories {
					processor(installation, repo)
				}

				if repoResp.NextPage == 0 {
					break
				}
				repoOpts.Page = repoResp.NextPage
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

func (a *ApplicationService) ProcessInstallationRepository(installationId int64, processor processFunc) error {
	installation, _, err := a.Client.Apps.GetInstallation(context.Background(), installationId)

	if err != nil {
		return err
	}

	token, _, err := a.Client.Apps.CreateInstallationToken(context.Background(), installation.GetID(), nil)
	if err != nil {
		return err
	}

	installationToken := token.GetToken()
	installationClient, _ := CreateClient(installationToken, a.Client.BaseURL.Host)

	var repoList []string
	repoOpts := &github.ListOptions{PerPage: 10}
	for {
		repos, repoResp, err := installationClient.Apps.ListRepos(context.Background(), repoOpts)
		if err != nil {
			return err
		}

		for _, repo := range repos.Repositories {
			repoList = append(repoList, repo.GetFullName())
		}

		if repoResp.NextPage == 0 {
			break
		}
		repoOpts.Page = repoResp.NextPage
	}

	endpoint := fmt.Sprintf("%s://%s%s", a.Client.BaseURL.Scheme, a.Client.BaseURL.Host, a.Client.BaseURL.Path)
	processor(repoList, installationToken, endpoint)

	return nil
}

func CreateClient(token string, endpoint string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	var client *github.Client
	if endpoint == "" || endpoint == "api.github.com" {
		client = github.NewClient(tc)
	} else {
		endpointUrl, err := url.Parse(fmt.Sprintf("https://%s/api/v3/", endpoint))
		if err != nil {
			return nil, err
		}

		client = github.NewClient(tc)
		client.BaseURL = endpointUrl
		client.UploadURL = endpointUrl
	}

	return client, nil
}

func GenerateJWT(applicationID string, privateKey *rsa.PrivateKey) (string, error) {
	// Create the claims
	claims := jwt.MapClaims{
		"iat": time.Now().Unix(),                      // Issued at time
		"exp": time.Now().Add(time.Minute * 5).Unix(), // Expiration time (10 minutes from now)
		"iss": applicationID,                          // GitHub App ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
