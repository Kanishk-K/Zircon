package apiresponse

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func APIErrorResponse(status int, message string, resp *events.APIGatewayProxyResponse) {
	resp.StatusCode = status
	resp.Body = fmt.Sprintf(`{"message": "%s"}`, message)
}
