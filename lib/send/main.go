package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"os"
)

type Request struct {
	TaskId  string `json:"taskId"`
	Message string `json:"message"`
}

type Connection struct {
	ConnectionId  string `dynamodbav:"connection_id"`
	TaskId        string `dynamodbav:"task_id"`
	Authenticated bool   `dynamodbav:"authenticated"`
}

func HandleRequest(ctx context.Context, event *Request) (*events.LambdaFunctionURLResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	dynamoDBClient := dynamodb.NewFromConfig(cfg)
	queryInput := dynamodb.QueryInput{
		TableName:              aws.String("connections"),
		IndexName:              aws.String("task_id-index"),
		KeyConditionExpression: aws.String("task_id = :taskId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":taskId": &types.AttributeValueMemberS{
				Value: event.TaskId,
			},
		},
	}
	queryOutput, err := dynamoDBClient.Query(ctx, &queryInput)
	if err != nil {
		return nil, err
	}

	var connections []Connection
	err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, &connections)
	if err != nil {
		return nil, err
	}

	connectionsURL := os.Getenv("CONNECTIONS_URL")
	apiGatewayManagementApiClient := apigatewaymanagementapi.NewFromConfig(cfg, func(options *apigatewaymanagementapi.Options) {
		options.BaseEndpoint = &connectionsURL
	})

	for _, connection := range connections {
		if connection.Authenticated {
			postToConnectionInput := apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: &connection.ConnectionId,
				Data:         []byte(event.Message),
			}
			_, err := apiGatewayManagementApiClient.PostToConnection(ctx, &postToConnectionInput)
			if err != nil {
				return nil, err
			}
		}
	}

	log.Printf("Sent message '%s' to %d clients", event.Message, len(connections))

	return &events.LambdaFunctionURLResponse{
		StatusCode: 202,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
