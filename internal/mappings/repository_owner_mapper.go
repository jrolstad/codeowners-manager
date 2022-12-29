package mappings

import (
	"github.com/jrolstad/codeowners-manager/internal/models"
)

func MapRepositoryOwners(toMap []*models.RepositoryOwner) []*models.RepositoryOwnerData {
	result := make([]*models.RepositoryOwnerData, 0)

	for _, item := range toMap {
		mappedItem := mapRepositoryOwner(item)
		result = append(result, mappedItem)
	}

	return result
}

func mapRepositoryOwner(toMap *models.RepositoryOwner) *models.RepositoryOwnerData {
	return &models.RepositoryOwnerData{
		Id:           "",
		Host:         toMap.Host,
		Organization: toMap.Organization,
		Repository:   toMap.Repository,
		Pattern:      toMap.Pattern,
		Owners:       toMap.Owners,
		Parent:       toMap.Parent,
	}
}

func MapRepositoryOwnersData(toMap []*models.RepositoryOwnerData) []*models.RepositoryOwner {
	result := make([]*models.RepositoryOwner, 0)

	for _, item := range toMap {
		mappedItem := mapRepositoryOwnerData(item)
		result = append(result, mappedItem)
	}

	return result
}

func mapRepositoryOwnerData(toMap *models.RepositoryOwnerData) *models.RepositoryOwner {
	return &models.RepositoryOwner{
		Host:         toMap.Host,
		Organization: toMap.Organization,
		Repository:   toMap.Repository,
		Pattern:      toMap.Pattern,
		Owners:       toMap.Owners,
		Parent:       toMap.Parent,
	}
}

func MapRepositoryOwnerValues(host string, organization string, repository string, pattern string, owners []string, parentOwner string) *models.RepositoryOwner {
	return &models.RepositoryOwner{
		Host:         host,
		Organization: organization,
		Repository:   repository,
		Pattern:      pattern,
		Owners:       owners,
		Parent:       parentOwner,
	}
}
