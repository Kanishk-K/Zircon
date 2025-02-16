package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	apiresponse "github.com/Kanishk-K/UniteDownloader/Backend/pkg/api-response"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/authutil"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type CallbackService struct {
	googleOauthConfig *oauth2.Config
	authClient        authutil.AuthClientMethods
}

func (cs *CallbackService) GetUserDataFromGoogle(codeValue string) (*authutil.ProfileData, error) {
	token, err := cs.googleOauthConfig.Exchange(context.Background(), codeValue)
	if err != nil {
		return nil, fmt.Errorf("code exchange went wrong: %s", err.Error())
	}
	client := cs.googleOauthConfig.Client(context.Background(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed getting info for user: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from API: %s", err.Error())
	}
	responseData := authutil.ProfileData{}
	if err = json.Unmarshal(contents, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse response from response: %s", err.Error())
	}
	return &responseData, nil
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

	JWTClient := authutil.NewAuthClient([]byte(os.Getenv("JWT_PRIVATE")))

	cbs := CallbackService{
		googleOauthConfig: GoogleOauthConfig,
		authClient:        JWTClient,
	}

	profile, err := cbs.GetUserDataFromGoogle(request.QueryStringParameters["code"])
	if err != nil {
		apiresponse.APIErrorResponse(500, "Failed to process callback, please try again.", &resp)
		return resp, err
	}
	if profile.OrganizationDomain != "umn.edu" {
		apiresponse.APIErrorResponse(401, "You're not registered with a valid \"umn.edu\" email.", &resp)
		return resp, nil
	}
	tokenDetails, err := cbs.authClient.SignJWT(profile)
	if err != nil {
		apiresponse.APIErrorResponse(500, "Failed to sign JWT, please try again.", &resp)
		return resp, err
	}
	tokenDetailsJSON, err := json.Marshal(tokenDetails)
	if err != nil {
		apiresponse.APIErrorResponse(500, "Failed to encode JWT, please try again.", &resp)
		return resp, err
	}
	resp.Body = string(tokenDetailsJSON)
	resp.StatusCode = 200
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
