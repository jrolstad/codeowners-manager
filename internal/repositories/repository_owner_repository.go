package repositories

import (
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/config"
	"github.com/jrolstad/codeowners-manager/internal/models"
	"time"
)

type RepositoryOwnerRepository interface {
	Get(host string, organization string, repository string, expiry time.Time) ([]*models.RepositoryOwnerData, error)
	Save(data []*models.RepositoryOwnerData, expiry time.Time) error
}

func NewRepositoryOwnerRepository(appConfig *config.AppConfig, secretClient clients.SecretClient) RepositoryOwnerRepository {
	repository := &DynamoDbRepositoryOwnerRepository{}
	repository.init(appConfig.AwsRegion, appConfig.RepositoryOwnerTableName)

	return repository
}
