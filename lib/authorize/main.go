package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"os"
)

type Message struct {
	Token string `json:"token"`
}

func getToken(event events.APIGatewayWebsocketProxyRequest) (*jwt.Token, error) {
	var message Message
	err := json.Unmarshal([]byte(event.Body), &message)
	if err != nil {
		return nil, err
	}

	jsonWebKeys, err := getJsonWebKeys(os.Getenv("JWKS_URL"))
	if err != nil {
		return nil, err
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return getPublicKeyForToken(jsonWebKeys, token)
	}
	token, err := jwt.Parse(message.Token, keyFunc)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func HandleRequest(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (*events.APIGatewayProxyResponse, error) {
	token, err := getToken(event)
	if err != nil {
		return nil, err
	}

	connectionId := event.RequestContext.ConnectionID

	if token.Valid {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}

		dynamoDBClient := dynamodb.NewFromConfig(cfg)
		updateItemInput := dynamodb.UpdateItemInput{
			TableName: aws.String("connections"),
			Key: map[string]types.AttributeValue{
				"connection_id": &types.AttributeValueMemberS{
					Value: connectionId,
				},
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":authenticated": &types.AttributeValueMemberBOOL{
					Value: true,
				},
			},
			UpdateExpression: aws.String("set authenticated = :authenticated"),
		}
		_, err = dynamoDBClient.UpdateItem(ctx, &updateItemInput)
		if err != nil {
			return nil, err
		}

		log.Printf("Client %s authenticated.", connectionId)
	} else {
		log.Printf("Client %s supplied an invalid token.", connectionId)
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: 202,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
