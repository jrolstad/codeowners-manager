package orchestration

import (
	"github.com/jrolstad/codeowners-manager/internal/config"
	"github.com/jrolstad/codeowners-manager/internal/core"
	"github.com/jrolstad/codeowners-manager/internal/logging"
	"github.com/jrolstad/codeowners-manager/internal/mappings"
	"github.com/jrolstad/codeowners-manager/internal/models"
	"github.com/jrolstad/codeowners-manager/internal/repositories"
	"github.com/jrolstad/codeowners-manager/internal/resolvers"
	"github.com/pkg/errors"
	"time"
)

func LoadRepositoryOwners(host string,
	organization string,
	appConfig *config.AppConfig,
	hostRepository repositories.HostRepository,
	repositoryOwnerRepository repositories.RepositoryOwnerRepository,
	repositoryOwnerResolver resolvers.RepositoryOwnerResolver) error {

	hosts, err := resolveHosts(host, hostRepository)
	if err != nil {
		return err
	}

	processor := func(data []*models.RepositoryOwner) {
		if data == nil || len(data) == 0 {
			return
		}

		logging.LogInfo("Processing RepositoryOwner data",
			"length", len(data))
		expiryTime := getRepositoryOwnerExpiryTime(time.Now().UTC(), appConfig)
		mappedData := mappings.MapRepositoryOwners(data)
		saveError := repositoryOwnerRepository.Save(mappedData, expiryTime)
		if saveError != nil {
			loggedError := errors.Wrap(saveError, "error when saving repository owners")
			logging.LogError(loggedError, "data", mappedData)
		} else {
			logging.LogInfo("Saved RepositoryOwner data",
				"length", len(data),
				"expiry", expiryTime.String())
		}
	}

	processingErrors := make([]error, 0)
	for _, host := range hosts {
		err := repositoryOwnerResolver.ProcessRepositoryOwners(host, organization, processor)
		if err != nil {
			processingErrors = append(processingErrors, err)
		}
	}

	return core.ConsolidateErrors(processingErrors)
}

func resolveHosts(host string, hostRepository repositories.HostRepository) ([]*models.Host, error) {
	if host != "" {
		hostData, err := hostRepository.Get(host)
		return []*models.Host{hostData}, err
	}

	return hostRepository.GetAll()

}
