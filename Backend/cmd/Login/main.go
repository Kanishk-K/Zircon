package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GenerateStateOAuthCookie() string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	return state
}

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

	// Get the PORT from environment variables
	host := os.Getenv("HOST")
	port := ""
	if host == "" {
		port = ":3000"
		host = "http://localhost"
		fmt.Println("No HOST environment variable detected, defaulting to", host)
		fmt.Println("No PORT environment variable detected, defaulting to", port)
	}

	GoogleOauthConfig := &oauth2.Config{
		RedirectURL:  host + port + "/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	oauthState := GenerateStateOAuthCookie()
	authURL := GoogleOauthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	resp.StatusCode = 302
	resp.Headers["Location"] = authURL
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
