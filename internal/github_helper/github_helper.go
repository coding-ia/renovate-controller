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

func GenerateInstallationToken(applicationID string, privateKey []byte, endpoint string, owner string, repository string) (string, error) {
	parsedKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return "", err
	}

	tokenString, err := generateJWT(applicationID, parsedKey)
	if err != nil {
		return "", fmt.Errorf("error generating JWT: %v", err)
	}

	client, err := createClient(tokenString, endpoint)
	if err != nil {
		return "", fmt.Errorf("error creating github client: %v", err)
	}

	repo := fmt.Sprintf("%s/%s", owner, repository)
	installation, err := filterInstallation(client, repo)
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

func createClient(token string, endpoint string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)

	var client *github.Client
	if endpoint == "" {
		client = github.NewClient(tc)
	} else {
		endpointUrl, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}

		client = github.NewClient(tc)
		client.BaseURL = endpointUrl
		client.UploadURL = endpointUrl
	}

	return client, nil
}

func generateJWT(applicationID string, privateKey *rsa.PrivateKey) (string, error) {
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

func filterInstallation(client *github.Client, targetRepo string) (*github.Installation, error) {
	opts := &github.ListOptions{PerPage: 10}
	for {
		installations, resp, err := client.Apps.ListInstallations(context.Background(), opts)
		if err != nil {
			return nil, err
		}

		for _, installation := range installations {
			token, _, err := client.Apps.CreateInstallationToken(context.Background(), installation.GetID(), nil)
			if err != nil {
				return nil, err
			}

			installationTS := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token.GetToken()},
			)
			installationClient := github.NewClient(oauth2.NewClient(context.Background(), installationTS))

			repoOpts := &github.ListOptions{PerPage: 10}
			for {
				repos, repoResp, err := installationClient.Apps.ListRepos(context.Background(), repoOpts)
				if err != nil {
					return nil, err
				}

				for _, repo := range repos.Repositories {
					if fmt.Sprintf("%s/%s", repo.GetOwner().GetLogin(), repo.GetName()) == targetRepo {
						return installation, nil
					}
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

	return nil, nil
}
