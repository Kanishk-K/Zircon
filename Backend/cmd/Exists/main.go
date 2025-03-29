package main

import (
	"context"
	"fmt"
	"log"
	"os"

	apiresponse "github.com/Kanishk-K/UniteDownloader/Backend/pkg/apiResponse"
	dynamo "github.com/Kanishk-K/UniteDownloader/Backend/pkg/dynamoClient"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
)

type ExistsService struct {
	dynamoClient dynamo.DynamoMethods
}

type ExistsRequest struct {
	EntryID string `json:"entryID"`
}

func NewExistsService(dynamoClient dynamo.DynamoMethods) *ExistsService {
	return &ExistsService{dynamoClient: dynamoClient}
}

func (es ExistsService) handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		IsBase64Encoded: false,
	}
	entryID := request.QueryStringParameters["entryID"]
	if entryID == "" {
		apiresponse.APIErrorResponse(400, "No EntryID provided", &resp)
		return resp, nil
	}
	jobInfo, err := es.dynamoClient.GetJob(entryID)
	if err != nil {
		log.Println("Error getting job info: ", err)
		apiresponse.APIErrorResponse(500, "Error getting job info", &resp)
	}
	respBody := make(map[string]any)
	if jobInfo != nil {
		if jobInfo.VideosAvailable != nil {
			respBody["videosAvailable"] = jobInfo.VideosAvailable
		}
	} else {
		apiresponse.APIErrorResponse(404, "Job not found", &resp)
		return resp, nil
	}
	apiresponse.APISuccessResponse(respBody, &resp)
	return resp, nil
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
	es := NewExistsService(dynamoClient)
	lambda.Start(es.handler)
}
