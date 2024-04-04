package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
)

func HandleRequest(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var connectionId = event.RequestContext.ConnectionID

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	dynamoDBClient := dynamodb.NewFromConfig(cfg)
	deleteItemInput := dynamodb.DeleteItemInput{
		TableName: aws.String("connections"),
		Key: map[string]types.AttributeValue{
			"connection_id": &types.AttributeValueMemberS{
				Value: connectionId,
			},
		},
	}
	_, err = dynamoDBClient.DeleteItem(ctx, &deleteItemInput)
	if err != nil {
		return nil, err
	}

	log.Printf("Client %s disconnected.", connectionId)

	return &events.APIGatewayProxyResponse{
		StatusCode: 202,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
