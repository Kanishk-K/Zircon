package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	dynamo "github.com/Kanishk-K/UniteDownloader/Backend/pkg/dynamoClient"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type PostSignUpService struct {
	dynamoClient dynamo.DynamoMethods
}

func (psus PostSignUpService) handler(request events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	log.Println(request.CognitoEventUserPoolsHeader.UserName)
	// Add the user to the DynamoDB table
	err := psus.dynamoClient.CreateUserIfNotExists(request.CognitoEventUserPoolsHeader.UserName)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			// The user already exists, this should not be called but it is not a critical error.
			log.Println("User already exists")
			return request, nil
		} else {
			log.Println("Error creating user: ", err)
			return request, err
		}
	}
	return request, nil
}

func main() {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	awsSession, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		fmt.Println("Failed to load AWS configuration:", err)
		return
	}
	dynamoClient := dynamo.NewDynamoClient(awsSession)
	psus := PostSignUpService{dynamoClient}
	lambda.Start(psus.handler)
}
