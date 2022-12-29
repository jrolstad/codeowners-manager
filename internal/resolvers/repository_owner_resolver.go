package resolvers

import (
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/models"
)

type RepositoryOwnerResolver interface {
	ProcessRepositoryOwners(host *models.Host, organization string, processor func([]*models.RepositoryOwner)) error
	ResolveRepositoryOwners(host *models.Host, organization string, repository string) ([]*models.RepositoryOwner, error)
}

func NewRepositoryOwnerResolver(secretClient clients.SecretClient) RepositoryOwnerResolver {
	instance := &SfdcRepositoryOwnerResolver{secretClient: secretClient}
	return instance
}
