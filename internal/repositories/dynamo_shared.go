package repositories

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"strconv"
	"time"
)

func getStringValue(item *dynamodb.AttributeValue) string {
	if item == nil || item.S == nil {
		return ""
	}
	return aws.StringValue(item.S)
}

func toDynamoString(value string) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{S: aws.String(value)}
}

func toDynamoTime(value time.Time) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(value.Unix(), 10))}
}

func getArrayValue(item *dynamodb.AttributeValue) []string {
	result := make([]string, 0)
	if item == nil || item.SS == nil {
		return result
	}

	for _, arrayValue := range item.SS {
		value := aws.StringValue(arrayValue)
		result = append(result, value)
	}

	return result
}

func toDynamoArray(value []string) *dynamodb.AttributeValue {
	result := &dynamodb.AttributeValue{}

	arrayValues := make([]*string, 0)
	for _, item := range value {
		itemValue := aws.String(item)
		arrayValues = append(arrayValues, itemValue)
	}

	result.SetSS(arrayValues)
	return result
}
