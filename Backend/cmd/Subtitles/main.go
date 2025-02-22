package main

import (
	apiresponse "github.com/Kanishk-K/UniteDownloader/Backend/pkg/api-response"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		IsBase64Encoded: false,
	}
	apiresponse.APIErrorResponse(200, "Successful Invocation", &resp)
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
