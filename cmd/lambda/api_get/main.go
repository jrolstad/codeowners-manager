package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/config"
	"github.com/jrolstad/codeowners-manager/internal/core"
	"github.com/jrolstad/codeowners-manager/internal/logging"
	"github.com/jrolstad/codeowners-manager/internal/models"
	"github.com/jrolstad/codeowners-manager/internal/orchestration"
	"github.com/jrolstad/codeowners-manager/internal/repositories"
	"github.com/jrolstad/codeowners-manager/internal/resolvers"
	"net/http"
)

var (
	appConfig                 *config.AppConfig
	secretClient              clients.SecretClient
	hostRepository            repositories.HostRepository
	repositoryOwnerRepository repositories.RepositoryOwnerRepository
	ownerResolver             resolvers.RepositoryOwnerResolver
)

func init() {
	appConfig = config.NewAppConfig()

	secretClient = clients.NewSecretClient(appConfig)
	hostRepository = repositories.NewHostRepository(appConfig, secretClient)
	repositoryOwnerRepository = repositories.NewRepositoryOwnerRepository(appConfig, secretClient)
	ownerResolver = resolvers.NewRepositoryOwnerResolver(secretClient)
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	host, organization, repository := parseArgumentsFromRequeset(event)

	result, err := orchestration.GetRepositoryOwners(host, organization, repository, appConfig, hostRepository, repositoryOwnerRepository, ownerResolver)
	if err != nil {
		logging.LogError(err)
	}

	return mapDataToResponse(result, err), err
}

func parseArgumentsFromRequeset(event events.APIGatewayProxyRequest) (string, string, string) {
	host := event.QueryStringParameters["host"]
	organization := event.QueryStringParameters["organization"]
	repository := event.QueryStringParameters["repository"]

	return host, organization, repository
}

func mapDataToResponse(data []*models.RepositoryOwner, err error) events.APIGatewayProxyResponse {
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Body: core.MapToJson(data)}
}
