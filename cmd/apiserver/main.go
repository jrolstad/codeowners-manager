package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/config"
	"github.com/jrolstad/codeowners-manager/internal/logging"
	"github.com/jrolstad/codeowners-manager/internal/models"
	"github.com/jrolstad/codeowners-manager/internal/orchestration"
	"github.com/jrolstad/codeowners-manager/internal/repositories"
	"github.com/jrolstad/codeowners-manager/internal/resolvers"
	"net/http"
	"time"
)

func main() {
	r := gin.Default()

	appConfig := config.NewAppConfig()

	secretClient := clients.NewSecretClient(appConfig)
	hostRepository := repositories.NewHostRepository(appConfig, secretClient)
	repositoryOwnerRepository := repositories.NewRepositoryOwnerRepository(appConfig, secretClient)
	ownerResolver := resolvers.NewRepositoryOwnerResolver(secretClient)

	r.GET("/repository/owner", func(c *gin.Context) {
		host, organization, repository := parseArgumentsFromRequest(c)

		result, err := orchestration.GetRepositoryOwners(host, organization, repository, appConfig, hostRepository, repositoryOwnerRepository, ownerResolver)
		if err != nil {
			logging.LogError(err)
		}

		mapDataToResponse(c, result, err)
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"now": time.Now()})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}

func parseArgumentsFromRequest(context *gin.Context) (string, string, string) {
	host := context.Query("host")
	organization := context.Query("organization")
	repository := context.Query("repository")

	return host, organization, repository
}

func mapDataToResponse(context *gin.Context, data []*models.RepositoryOwner, err error) {
	if err != nil {
		context.JSON(http.StatusInternalServerError, data)
	} else {
		context.JSON(http.StatusOK, data)
	}
}
