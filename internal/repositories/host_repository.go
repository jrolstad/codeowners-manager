package repositories

import (
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/config"
	"github.com/jrolstad/codeowners-manager/internal/models"
)

type HostRepository interface {
	GetAll() ([]*models.Host, error)
	Get(identifier string) (*models.Host, error)
}

func NewHostRepository(appConfig *config.AppConfig, secretClient clients.SecretClient) HostRepository {
	repository := &DynamoDbHostRepository{}
	repository.init(appConfig.AwsRegion, appConfig.HostTableName)

	return repository
}
