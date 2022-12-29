package resolvers

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/core"
	"github.com/jrolstad/codeowners-manager/internal/logging"
	"github.com/jrolstad/codeowners-manager/internal/mappings"
	"github.com/jrolstad/codeowners-manager/internal/models"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

type SfdcRepositoryOwnerResolver struct {
	secretClient clients.SecretClient
}

const (
	githubClientTypeEnterpriseServer = "GitHub Enterprise Server"
)

func (r *SfdcRepositoryOwnerResolver) ProcessRepositoryOwners(host *models.Host,
	organization string,
	processor func([]*models.RepositoryOwner)) error {
	hostSecret, err := r.secretClient.GetSecret(host.ClientSecretName)
	if err != nil {
		return err
	}

	client, err := clients.GetGitHubClient(host.SubType, host.BaseUrl, host.AuthenticationType, hostSecret)
	if err != nil {
		return err
	}

	if organization != "" {
		organizationData, _, err := client.Organizations.Get(context.Background(), organization)
		if err != nil {
			return err
		}
		return r.processOwnersInOrganization(host, client, organizationData, processor)
	}

	return r.processOrganizationsOnHost(host, client, processor)
}

func (r *SfdcRepositoryOwnerResolver) processOrganizationsOnHost(host *models.Host,
	client *github.Client,
	processor func([]*models.RepositoryOwner)) error {
	if strings.EqualFold(githubClientTypeEnterpriseServer, host.SubType) {
		return r.processAllOrganizationsOnHost(host, client, processor)
	}

	return r.processMembersOrganizationsOnHost(host, client, processor)
}

func (r *SfdcRepositoryOwnerResolver) processAllOrganizationsOnHost(host *models.Host,
	client *github.Client,
	processor func([]*models.RepositoryOwner)) error {
	listOptions := &github.OrganizationsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	processingErrors := make([]error, 0)
	for {
		organizations, response, err := client.Organizations.ListAll(context.Background(), listOptions)
		if err != nil {
			processingErrors = append(processingErrors, err)
		}

		for _, item := range organizations {
			err := r.processOwnersInOrganization(host, client, item, processor)
			if err != nil {
				processingErrors = append(processingErrors, err)
			}
		}

		if response == nil || response.NextPage == 0 || len(organizations) == 0 {
			break
		}

		listOptions.Since = getLastOrganization(organizations)
		listOptions.Page = response.NextPage
	}

	return core.ConsolidateErrors(processingErrors)
}

func getLastOrganization(data []*github.Organization) int64 {
	lastOrganizationPosition := len(data) - 1
	return data[lastOrganizationPosition].GetID()
}

func (r *SfdcRepositoryOwnerResolver) processMembersOrganizationsOnHost(host *models.Host,
	client *github.Client,
	processor func([]*models.RepositoryOwner)) error {
	processingErrors := make([]error, 0)
	listOptions := &github.ListOrgMembershipsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		memberOrganizations, response, err := client.Organizations.ListOrgMemberships(context.Background(), listOptions)
		if err != nil {
			processingErrors = append(processingErrors, err)
		}

		for _, item := range memberOrganizations {
			err = r.processOwnersInOrganization(host, client, item.GetOrganization(), processor)
			if err != nil {
				processingErrors = append(processingErrors, err)
			}
		}

		if response == nil || response.NextPage == 0 || len(memberOrganizations) == 0 {
			break
		}

		listOptions.Page = response.NextPage
	}

	return core.ConsolidateErrors(processingErrors)
}

func (r *SfdcRepositoryOwnerResolver) processOwnersInOrganization(host *models.Host,
	client *github.Client,
	organization *github.Organization,
	processor func([]*models.RepositoryOwner)) error {
	logging.LogInfo("Processing Organization Owners", "organization", organization.GetLogin(), "url", organization.GetHTMLURL())

	processingErrors := make([]error, 0)
	codeOwners, err := r.getCodeOwnersForOrganization(client, organization.GetLogin(), "")
	if err != nil {
		processingErrors = append(processingErrors, errors.Wrapf(err, "Unable to find CODEOWNERS for %s", organization.GetURL()))
	}

	opt := &github.RepositoryListByOrgOptions{
		Sort:        "full_name",
		Direction:   "asc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repositories, response, err := client.Repositories.ListByOrg(context.Background(), organization.GetLogin(), opt)
		if err != nil {
			return err
		}

		for _, item := range repositories {
			logging.LogInfo("Processing Repository Owners", "organization", item.GetOrganization().GetLogin(),
				"repository", item.GetName(),
				"url", item.GetHTMLURL())

			ownerData, err := r.resolveRepositoryCodeOwners(host, organization.GetLogin(), item.GetName(), codeOwners)
			if err != nil {
				processingErrors = append(processingErrors, errors.Wrapf(err, "error when processing %s", item.GetURL()))
				continue
			}

			processor(ownerData)
		}

		if response.NextPage == 0 {
			break
		}

		opt.Page = response.NextPage
	}

	return core.ConsolidateErrors(processingErrors)
}

func (r *SfdcRepositoryOwnerResolver) ResolveRepositoryOwners(host *models.Host,
	organization string,
	repository string) ([]*models.RepositoryOwner, error) {
	defaultResult := make([]*models.RepositoryOwner, 0)

	hostSecret, err := r.secretClient.GetSecret(host.ClientSecretName)
	if err != nil {
		return defaultResult, err
	}

	client, err := clients.GetGitHubClient(host.SubType, host.BaseUrl, host.AuthenticationType, hostSecret)
	if err != nil {
		return defaultResult, err
	}

	_, response, err := client.Repositories.Get(context.Background(), organization, repository)
	if response != nil {
		if response.StatusCode == http.StatusNotFound {
			return defaultResult, nil
		}
	}
	if err != nil {
		return defaultResult, err
	}

	codeOwners, err := r.getCodeOwnersForOrganization(client, organization, repository)
	if err != nil {
		return defaultResult, err
	}

	return r.resolveRepositoryCodeOwners(host, organization, repository, codeOwners)

}

func (r *SfdcRepositoryOwnerResolver) resolveRepositoryCodeOwners(host *models.Host,
	organization string,
	repository string,
	codeOwners map[string]map[string]*codeOwnerData) ([]*models.RepositoryOwner, error) {
	repositoryCodeOwner := r.coalesceCodeOwners(codeOwners[strings.ToLower(repository)]["CODEOWNERS"],
		codeOwners[strings.ToLower(repository)]["docs/CODEOWNERS"],
		codeOwners[strings.ToLower(repository)][".github/CODEOWNERS"])
	organizationCodeOwner := r.coalesceCodeOwners(
		codeOwners["sfdc-codeowners"][fmt.Sprintf("%s/CODEOWNERS", strings.ToLower(repository))],
		codeOwners["sfdc-codeowners"]["sfdc-codeowners-uo/CODEOWNERS"])

	repositoryCodeOwners := make([]*models.RepositoryOwner, 0)
	if repositoryCodeOwner != nil {
		data := r.parseCodeOwners(host, organization, repository, repositoryCodeOwner.Contents)
		repositoryCodeOwners = append(repositoryCodeOwners, data...)
	}

	organizationCodeOwners := make([]*models.RepositoryOwner, 0)
	if organizationCodeOwner != nil {
		data := r.parseCodeOwners(host, organization, repository, organizationCodeOwner.Contents)
		organizationCodeOwners = append(organizationCodeOwners, data...)
	}
	r.applyOrganizationDefaults(repositoryCodeOwners, organizationCodeOwners)

	if len(repositoryCodeOwners) > 0 {
		return repositoryCodeOwners, nil
	}

	return organizationCodeOwners, nil
}

func (r *SfdcRepositoryOwnerResolver) getCodeOwnersForOrganization(client *github.Client,
	organization string,
	repository string) (map[string]map[string]*codeOwnerData, error) {
	searchOptions := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	results := make(map[string]map[string]*codeOwnerData, 0)
	query := r.buildCodeOwnersSearchQuery(client, organization, repository)

	logging.LogInfo("Searching host for CODEOWNERS", "query", query)
	for {
		result, response, err := client.Search.Code(context.Background(), query, searchOptions)
		if err != nil {
			return results, err
		}
		if result.GetIncompleteResults() {
			logging.LogError(errors.New("incomplete github search results"), "query", query, "page", searchOptions.Page)
		}

		for _, item := range result.CodeResults {
			repositoryName := item.GetRepository().GetName()
			path := item.GetPath()

			if results[strings.ToLower(repositoryName)] == nil {
				results[strings.ToLower(repositoryName)] = make(map[string]*codeOwnerData, 0)
			}

			data := &codeOwnerData{
				Organization: strings.ToLower(organization),
				Repository:   strings.ToLower(repositoryName),
				Path:         path,
			}
			results[strings.ToLower(repositoryName)][data.Path] = data
		}

		if response.NextPage == 0 {
			break
		}
		searchOptions.Page = response.NextPage
	}

	r.getCodeOwnersContent(client, results)
	return results, nil
}

func (r *SfdcRepositoryOwnerResolver) buildCodeOwnersSearchQuery(client *github.Client,
	organization string,
	repository string) string {
	if repository == "" {
		return fmt.Sprintf("filename:CODEOWNERS org:%s", organization)
	}

	organizationCodeOwnersRepositoryData, _, _ := client.Repositories.Get(context.Background(), organization, "sfdc-codeowners")

	repositoryName := fmt.Sprintf("%s/%s", organization, repository)
	organizationCodeOwnersRepositoryName := fmt.Sprintf("%s/sfdc-codeowners", organization)

	if organizationCodeOwnersRepositoryData == nil {
		return fmt.Sprintf("filename:CODEOWNERS repo:%s", repositoryName)
	} else {
		query := fmt.Sprintf("filename:CODEOWNERS repo:%s repo:%s", repositoryName, organizationCodeOwnersRepositoryName)
		return query
	}
}

func (r *SfdcRepositoryOwnerResolver) getCodeOwnersContent(client *github.Client,
	organizationCodeOwners map[string]map[string]*codeOwnerData) {
	options := &github.RepositoryContentGetOptions{}
	for _, repositoryCodeOwners := range organizationCodeOwners {
		for _, file := range repositoryCodeOwners {
			fileContent, _, _, err := client.Repositories.GetContents(context.Background(), file.Organization, file.Repository, file.Path, options)
			if err == nil && fileContent != nil {
				content, contentErr := fileContent.GetContent()
				if contentErr == nil {
					file.Contents = content
				}
			}
		}
	}
}

func (r *SfdcRepositoryOwnerResolver) coalesceCodeOwners(items ...*codeOwnerData) *codeOwnerData {
	for _, value := range items {
		if value != nil {
			return value
		}
	}

	return nil
}

func (r *SfdcRepositoryOwnerResolver) parseCodeOwners(host *models.Host,
	organization string,
	repository string,
	contents string) []*models.RepositoryOwner {
	if strings.TrimSpace(contents) == "" {
		return make([]*models.RepositoryOwner, 0)
	}

	owners := make(map[string][]*models.RepositoryOwner, 0)

	linesInFile := strings.Split(contents, "\n")

	const commentPrefix = "#"

	parentOwner := ""
	for _, line := range linesInFile {
		cleanLine := strings.TrimSpace(line)
		if strings.HasPrefix(cleanLine, host.ParentOwnerLinePattern) {
			delimitedValues := strings.TrimSpace(strings.ReplaceAll(cleanLine, host.ParentOwnerLinePattern, ""))
			splitValues := strings.Split(delimitedValues, ",")

			parentOwner = core.GetValueAt(splitValues, 0)
		}

		if owners[parentOwner] == nil {
			owners[parentOwner] = make([]*models.RepositoryOwner, 0)
		}

		if strings.HasPrefix(cleanLine, commentPrefix) || cleanLine == "" {
			continue
		}

		ownerParts := strings.Fields(cleanLine)
		pattern := core.GetValueAt(ownerParts, 0)

		patternOwners := ownerParts[1:]
		ownerData := mappings.MapRepositoryOwnerValues(host.Name, organization, repository, pattern, patternOwners, parentOwner)
		owners[parentOwner] = append(owners[parentOwner], ownerData)
	}

	ownersWithDefaults := r.applyDefaultOwners(host.Name, organization, repository, owners, parentOwner)

	return r.mapRepositoryOwnersToSlice(ownersWithDefaults)
}

func (r *SfdcRepositoryOwnerResolver) applyDefaultOwners(host string,
	organization string,
	repository string,
	owners map[string][]*models.RepositoryOwner,
	parentOwner string) map[string][]*models.RepositoryOwner {
	for key, value := range owners {
		if len(value) == 0 {
			defaultOwner := mappings.MapRepositoryOwnerValues(host, organization, repository, "*", []string{}, parentOwner)
			owners[key] = []*models.RepositoryOwner{defaultOwner}
		}
	}

	return owners
}

func (r *SfdcRepositoryOwnerResolver) mapRepositoryOwnersToSlice(data map[string][]*models.RepositoryOwner) []*models.RepositoryOwner {
	results := make([]*models.RepositoryOwner, 0)

	for _, value := range data {
		results = append(results, value...)
	}

	return results
}

func (r *SfdcRepositoryOwnerResolver) applyOrganizationDefaults(repositoryCodeOwners []*models.RepositoryOwner,
	organizationCodeOwners []*models.RepositoryOwner) {
	for _, item := range repositoryCodeOwners {
		for _, orgItem := range organizationCodeOwners {
			if item.Parent == "" {
				item.Parent = orgItem.Parent
			}
		}
	}
}

type codeOwnerData struct {
	Organization string
	Repository   string
	Path         string
	Contents     string
}
