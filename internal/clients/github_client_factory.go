package clients

import (
	"context"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
	"strings"
)

const (
	githubClientTypeEnterpriseServer = "GitHub Enterprise Server"
)

func GetGitHubClient(hostType string, baseUrl, authenticationType string, authenticationSecret string) (*github.Client, error) {
	context := context.Background()

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authenticationSecret},
	)
	tokenClient := oauth2.NewClient(context, tokenSource)

	if strings.EqualFold(githubClientTypeEnterpriseServer, hostType) {
		client, err := github.NewEnterpriseClient(baseUrl, baseUrl, tokenClient)
		return client, err
	}

	client := github.NewClient(tokenClient)
	return client, nil
}
