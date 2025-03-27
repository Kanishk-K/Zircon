package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
	// Validate that the user is part of the ORGANIZATION email.
	organization := os.Getenv("ORGANIZATION")
	log.Print(request)
	request.Response.AutoConfirmUser = false
	request.Response.AutoVerifyEmail = false
	request.Response.AutoVerifyPhone = false
	if organization == "" {
		log.Fatal("ORGANIZATION environment variable is not set")
		return request, fmt.Errorf("organization environment variable is not set.")
	}
	if request.Request.UserAttributes["email"] == "" {
		return request, fmt.Errorf("email is empty")
	}
	if !strings.HasSuffix(request.Request.UserAttributes["email"], organization) {
		return request, fmt.Errorf("email is not part of the organization")
	}

	request.Response.AutoConfirmUser = true
	request.Response.AutoVerifyEmail = true
	request.Response.AutoVerifyPhone = false
	return request, nil
}

func main() {
	lambda.Start(handler)
}
