package clients

import "github.com/jrolstad/codeowners-manager/internal/config"

type SecretClient interface {
	GetSecret(name string) (string, error)
}

func NewSecretClient(appConfig *config.AppConfig) SecretClient {
	client := &SecretManagerClient{awsRegion: appConfig.AwsRegion}
	client.init()

	return client
}
