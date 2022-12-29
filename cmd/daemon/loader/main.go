package main

import (
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/config"
	"github.com/jrolstad/codeowners-manager/internal/logging"
	"github.com/jrolstad/codeowners-manager/internal/orchestration"
	"github.com/jrolstad/codeowners-manager/internal/repositories"
	"github.com/jrolstad/codeowners-manager/internal/resolvers"
)

func main() {
	appConfig := config.NewAppConfig()

	secretClient := clients.NewSecretClient(appConfig)
	hostRepository := repositories.NewHostRepository(appConfig, secretClient)
	repositoryOwnerRepository := repositories.NewRepositoryOwnerRepository(appConfig, secretClient)
	ownerResolver := resolvers.NewRepositoryOwnerResolver(secretClient)

	const noHostSpecified = ""
	const noOrganizationSpecified = ""

	err := orchestration.LoadRepositoryOwners(noHostSpecified, noOrganizationSpecified, appConfig, hostRepository, repositoryOwnerRepository, ownerResolver)
	if err != nil {
		logging.LogPanic(err)
	}
}
