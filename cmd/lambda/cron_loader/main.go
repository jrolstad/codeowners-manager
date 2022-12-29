package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/config"
	"github.com/jrolstad/codeowners-manager/internal/logging"
	"github.com/jrolstad/codeowners-manager/internal/orchestration"
	"github.com/jrolstad/codeowners-manager/internal/repositories"
	"github.com/jrolstad/codeowners-manager/internal/resolvers"
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

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	const noHostSpecified = ""
	const noOrganizationSpecified = ""
	err := orchestration.LoadRepositoryOwners(noHostSpecified, noOrganizationSpecified, appConfig, hostRepository, repositoryOwnerRepository, ownerResolver)
	if err != nil {
		logging.LogError(err)
	}

	return err
}
