package github_helper

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
	"net/url"
	"time"
)

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

type processFunc func(*github.Repository, string, string)

func ProcessAllInstallationRepositories(client *github.Client, processor processFunc) error {
	opts := &github.ListOptions{PerPage: 10}
	for {
		installations, resp, err := client.Apps.ListInstallations(context.Background(), opts)
		if err != nil {
			return err
		}

		for _, installation := range installations {
			token, _, err := client.Apps.CreateInstallationToken(context.Background(), installation.GetID(), nil)
			if err != nil {
				return err
			}

			installationToken := token.GetToken()
			installationClient, _ := CreateClient(installationToken, client.BaseURL.Host)

			repoOpts := &github.ListOptions{PerPage: 10}
			for {
				repos, repoResp, err := installationClient.Apps.ListRepos(context.Background(), repoOpts)
				if err != nil {
					return err
				}

				for _, repo := range repos.Repositories {
					endpoint := fmt.Sprintf("%s://%s%s", client.BaseURL.Scheme, client.BaseURL.Host, client.BaseURL.Path)
					processor(repo, installationToken, endpoint)
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
