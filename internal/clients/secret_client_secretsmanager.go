package clients

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type SecretManagerClient struct {
	awsRegion string
	client    *secretsmanager.SecretsManager
}

func (c *SecretManagerClient) init() {
	session := GetAwsSession(c.awsRegion)
	c.client = secretsmanager.New(session)
}

func (c *SecretManagerClient) GetSecret(name string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	}

	secretValue, err := c.client.GetSecretValue(input)
	if err != nil {
		return "", err
	}

	return *secretValue.SecretString, nil
}
