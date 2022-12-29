package repositories

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/jrolstad/codeowners-manager/internal/clients"
	"github.com/jrolstad/codeowners-manager/internal/models"
)

type DynamoDbHostRepository struct {
	awsRegion string
	tableName string
	client    *dynamodb.DynamoDB
}

func (r *DynamoDbHostRepository) init(awsRegion string, tableName string) {
	r.awsRegion = awsRegion
	r.tableName = tableName

	session := clients.GetAwsSession(r.awsRegion)
	r.client = dynamodb.New(session)
}

func (r *DynamoDbHostRepository) GetAll() ([]*models.Host, error) {
	result := make([]*models.Host, 0)
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}
	queryResult, err := r.client.Scan(scanInput)
	if err != nil {
		return result, err
	}

	for _, item := range queryResult.Items {
		hostData := r.mapItemToHost(item)
		result = append(result, hostData)
	}

	return result, nil
}

func (r *DynamoDbHostRepository) Get(identifier string) (*models.Host, error) {
	itemInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(identifier),
			},
		},
		TableName: aws.String(r.tableName),
	}
	queryResult, err := r.client.GetItem(itemInput)
	if err != nil {
		return nil, err
	}

	result := r.mapItemToHost(queryResult.Item)
	return result, nil
}

func (r *DynamoDbHostRepository) mapItemToHost(item map[string]*dynamodb.AttributeValue) *models.Host {
	return &models.Host{
		Id:                     getStringValue(item["Id"]),
		Name:                   getStringValue(item["Name"]),
		BaseUrl:                getStringValue(item["BaseUrl"]),
		Type:                   getStringValue(item["Type"]),
		SubType:                getStringValue(item["SubType"]),
		AuthenticationType:     getStringValue(item["AuthenticationType"]),
		ClientSecretName:       getStringValue(item["ClientSecretName"]),
		ParentOwnerLinePattern: getStringValue(item["ParentOwnerLinePattern"]),
	}
}
