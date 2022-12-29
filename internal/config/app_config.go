package config

import (
	"os"
	"strconv"
)

type AppConfig struct {
	AwsRegion                string
	HostTableName            string
	RepositoryOwnerTableName string
	DefaultTTLMinutes        int
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		AwsRegion:                os.Getenv("aws_region"),
		HostTableName:            os.Getenv("codeowners_host_table"),
		RepositoryOwnerTableName: os.Getenv("codeowners_repositoryowner_table"),
		DefaultTTLMinutes:        getIntegerConfigValue("codeowners_ttl_minutes", 60),
	}
}

func getIntegerConfigValue(name string, defaultValue int) int {
	rawValue := os.Getenv(name)
	if rawValue == "" {
		return defaultValue
	}

	parsedValue, err := strconv.Atoi(rawValue)
	if err != nil {
		return defaultValue
	}

	return parsedValue
}
