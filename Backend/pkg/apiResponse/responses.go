package apiresponse

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

func APISuccessResponse(body any, resp *events.APIGatewayProxyResponse) {
	resp.StatusCode = 200
	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Printf("Failed to marshal response body: %v", err)
		APIErrorResponse(500, "Internal Server Error", resp)
		return
	}
	resp.Body = string(jsonData)
}

func APIErrorResponse(status int, message string, resp *events.APIGatewayProxyResponse) {
	resp.StatusCode = status
	resp.Body = fmt.Sprintf(`{"message": "%s"}`, message)
}
