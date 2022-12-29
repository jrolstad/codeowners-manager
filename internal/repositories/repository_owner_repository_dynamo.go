package repositories

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/core"
	"github.com/jrolstad/codeowners-manager/internal/models"
	"time"
)

type DynamoDbRepositoryOwnerRepository struct {
	awsRegion string
	tableName string
	client    *dynamodb.DynamoDB
}

func (r *DynamoDbRepositoryOwnerRepository) init(awsRegion string, tableName string) {
	r.awsRegion = awsRegion
	r.tableName = tableName

	session := clients.GetAwsSession(r.awsRegion)
	r.client = dynamodb.New(session)
}

func (r *DynamoDbRepositoryOwnerRepository) Get(host string, organization string, repository string, expiry time.Time) ([]*models.RepositoryOwnerData, error) {
	result := make([]*models.RepositoryOwnerData, 0)

	filterExpression, err := r.buildGetFilterExpression(host, organization, repository, expiry)
	if err != nil {
		return result, err
	}
	scanInput := &dynamodb.ScanInput{
		TableName:                 aws.String(r.tableName),
		FilterExpression:          filterExpression.Filter(),
		ExpressionAttributeNames:  filterExpression.Names(),
		ExpressionAttributeValues: filterExpression.Values(),
	}
	queryResult, err := r.client.Scan(scanInput)
	if err != nil {
		return result, err
	}

	for _, item := range queryResult.Items {
		ownerData := r.mapAttributesToRepositoryOwner(item)
		result = append(result, ownerData)
	}

	return result, nil
}

func (r *DynamoDbRepositoryOwnerRepository) buildGetFilterExpression(host string, organization string, repository string, expiry time.Time) (expression.Expression, error) {
	filter := expression.Name("Host").Equal(expression.Value(host)).
		And(expression.Name("Organization").Equal(expression.Value(organization))).
		And(expression.Name("Repository").Equal(expression.Value(repository))).
		And(expression.Name("ExpiresAt").GreaterThan(expression.Value(expiry.Unix())))

	return expression.NewBuilder().
		WithFilter(filter).Build()
}

func (r *DynamoDbRepositoryOwnerRepository) mapAttributesToRepositoryOwner(item map[string]*dynamodb.AttributeValue) *models.RepositoryOwnerData {
	return &models.RepositoryOwnerData{
		Id:           getStringValue(item["Id"]),
		Host:         getStringValue(item["Host"]),
		Organization: getStringValue(item["Organization"]),
		Repository:   getStringValue(item["Repository"]),
		Parent:       getStringValue(item["Parent"]),
		Pattern:      getStringValue(item["Pattern"]),
		Owners:       getArrayValue(item["Owners"]),
	}
}

func (r *DynamoDbRepositoryOwnerRepository) mapRepositoryOwnerToAttributes(data *models.RepositoryOwnerData, expiresAt time.Time) map[string]*dynamodb.AttributeValue {
	resolvedOwners := make([]string, 0)
	if len(data.Owners) == 0 {
		resolvedOwners = append(resolvedOwners, "")
	} else {
		resolvedOwners = data.Owners
	}
	return map[string]*dynamodb.AttributeValue{
		"Id":           toDynamoString(r.resolveRepositoryOwnerId(data)),
		"Host":         toDynamoString(data.Host),
		"Organization": toDynamoString(data.Organization),
		"Repository":   toDynamoString(data.Repository),
		"Parent":       toDynamoString(data.Parent),
		"Pattern":      toDynamoString(data.Pattern),
		"Owners":       toDynamoArray(resolvedOwners),
		"ExpiresAt":    toDynamoTime(expiresAt),
	}
}

func (r *DynamoDbRepositoryOwnerRepository) resolveRepositoryOwnerId(data *models.RepositoryOwnerData) string {
	if data.Id == "" {
		data.Id = core.MapUniqueIdentifier(data.Host, data.Organization, data.Repository, data.Pattern, data.Parent, core.MergeValues(data.Owners))
	}

	return data.Id
}

func (r *DynamoDbRepositoryOwnerRepository) Save(data []*models.RepositoryOwnerData, expiry time.Time) error {
	writeInput := &dynamodb.BatchWriteItemInput{
		RequestItems: r.mapRepositoryOwnersToWriteRequests(data, expiry),
	}

	_, err := r.client.BatchWriteItem(writeInput)
	return err
}

func (r *DynamoDbRepositoryOwnerRepository) mapRepositoryOwnersToWriteRequests(data []*models.RepositoryOwnerData, expiresAt time.Time) map[string][]*dynamodb.WriteRequest {
	result := make(map[string][]*dynamodb.WriteRequest)

	writeRequests := make([]*dynamodb.WriteRequest, 0)

	dataRequests := make(map[string]*dynamodb.PutRequest, 0)
	for _, item := range data {
		itemRequest := r.mapRepositoryOwnerToPutRequest(item, expiresAt)

		// Ensure there is only 1 items with the id in this batch
		if dataRequests[*itemRequest.Item["Id"].S] != nil {
			continue
		}
		dataRequests[*itemRequest.Item["Id"].S] = itemRequest

		putRequest := &dynamodb.WriteRequest{
			PutRequest: itemRequest,
		}
		writeRequests = append(writeRequests, putRequest)
	}

	result[r.tableName] = writeRequests

	return result
}

func (r *DynamoDbRepositoryOwnerRepository) mapRepositoryOwnerToPutRequest(data *models.RepositoryOwnerData, expiresAt time.Time) *dynamodb.PutRequest {
	return &dynamodb.PutRequest{
		Item: r.mapRepositoryOwnerToAttributes(data, expiresAt),
	}
}
