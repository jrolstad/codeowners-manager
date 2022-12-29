package orchestration

import (
	"errors"
	"github.com/jrolstad/codeowners-manager/internal/config"
	"github.com/jrolstad/codeowners-manager/internal/logging"
	"github.com/jrolstad/codeowners-manager/internal/mappings"
	"github.com/jrolstad/codeowners-manager/internal/models"
	"github.com/jrolstad/codeowners-manager/internal/repositories"
	"github.com/jrolstad/codeowners-manager/internal/resolvers"
	"time"
)

func GetRepositoryOwners(host string,
	organization string,
	repository string,
	appConfig *config.AppConfig,
	hostRepository repositories.HostRepository,
	repositoryOwnerRepository repositories.RepositoryOwnerRepository,
	repositoryOwnerResolver resolvers.RepositoryOwnerResolver) ([]*models.RepositoryOwner, error) {

	now := time.Now().UTC()

	logging.LogInfo("GetRepositoryOwners",
		"host", host,
		"organization", organization,
		"repository", repository)
	defaultResult := make([]*models.RepositoryOwner, 0)

	if host == "" || organization == "" || repository == "" {
		return defaultResult, errors.New("input parameters are not specified")
	}

	hostData, err := hostRepository.Get(host)
	if err != nil {
		return defaultResult, err
	}
	logging.LogInfo("Host details obtained", "id", hostData.Id)

	repositoryOwners, err := repositoryOwnerRepository.Get(hostData.Name, organization, repository, now)
	if err != nil {
		return defaultResult, err
	}
	logging.LogInfo("Existing repository owners obtained", "count", len(repositoryOwners))

	if len(repositoryOwners) > 0 {
		mappedValues := mappings.MapRepositoryOwnersData(repositoryOwners)
		return mappedValues, nil
	}

	resolvedOwners, err := repositoryOwnerResolver.ResolveRepositoryOwners(hostData, organization, repository)
	if err != nil {
		return defaultResult, err
	}
	logging.LogInfo("New repository owners resolved", "count", len(resolvedOwners))

	if len(resolvedOwners) == 0 {
		return defaultResult, nil
	}

	resolvedOwnerData := mappings.MapRepositoryOwners(resolvedOwners)

	expiryTime := getRepositoryOwnerExpiryTime(now, appConfig)
	logging.LogInfo("Saving repository owners", "expiry", expiryTime.String())
	err = repositoryOwnerRepository.Save(resolvedOwnerData, expiryTime)
	if err != nil {
		return defaultResult, err
	}
	logging.LogInfo("Resolved repository owners saved", "count", len(resolvedOwners))

	return resolvedOwners, nil
}

func getRepositoryOwnerExpiryTime(now time.Time, appConfig *config.AppConfig) time.Time {
	expiryTime := now.Add(time.Minute * time.Duration(appConfig.DefaultTTLMinutes))
	return expiryTime
}
