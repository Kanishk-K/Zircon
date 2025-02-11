package main

import (
	"encoding/json"

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
	retVal, err := json.Marshal(map[string]interface{}{"message": "Hello World"})
	if err != nil {
		resp.StatusCode = 500
		return resp, err
	}
	resp.StatusCode = 200
	resp.Body = string(retVal)
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
