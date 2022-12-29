package main

import (
	"errors"
	"flag"
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/config"
	"github.com/jrolstad/codeowners-manager/internal/logging"
	"github.com/jrolstad/codeowners-manager/internal/orchestration"
	"github.com/jrolstad/codeowners-manager/internal/repositories"
	"github.com/jrolstad/codeowners-manager/internal/resolvers"
	"strings"
)

var (
	actionArgument       = flag.String("action", "get", "Action to perform")
	hostArgument         = flag.String("host", "", "Host to search")
	organizationArgument = flag.String("organization", "", "Organization name")
	repositoryArgument   = flag.String("repository", "", "Repository name")
)

func main() {
	flag.Parse()
	appConfig := config.NewAppConfig()

	secretClient := clients.NewSecretClient(appConfig)
	hostRepository := repositories.NewHostRepository(appConfig, secretClient)
	repositoryOwnerRepository := repositories.NewRepositoryOwnerRepository(appConfig, secretClient)
	ownerResolver := resolvers.NewRepositoryOwnerResolver(secretClient)

	if strings.EqualFold(*actionArgument, "get") {
		result, err := orchestration.GetRepositoryOwners(*hostArgument, *organizationArgument, *repositoryArgument, appConfig, hostRepository, repositoryOwnerRepository, ownerResolver)
		if err != nil {
			logging.LogPanic(err)
		}

		logging.LogInfo("Result obtained", "result", result)
	} else if strings.EqualFold(*actionArgument, "load") {
		err := orchestration.LoadRepositoryOwners(*hostArgument, *organizationArgument, appConfig, hostRepository, repositoryOwnerRepository, ownerResolver)
		if err != nil {
			logging.LogPanic(err)
		}

		logging.LogInfo("Owners loaded")
	} else {
		logging.LogPanic(errors.New("unknown action"), "action", *actionArgument)
	}

}
