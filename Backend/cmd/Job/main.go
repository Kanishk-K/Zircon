package main

import (
	"fmt"

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

	/*
		Buisness logic goes here
	*/
	subject := request.RequestContext.Authorizer["lambda"].(map[string]interface{})["sub"]
	resp.Body = fmt.Sprintf("Hello, %s", subject)
	resp.StatusCode = 200
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
