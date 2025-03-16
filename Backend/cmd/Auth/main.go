package main

import (
	"os"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/authutil"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type AuthService struct {
	authClient authutil.AuthClientMethods
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	// Buisness logic goes here
	resp := events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: false,
	}

	JWTClient := authutil.NewAuthClient([]byte(os.Getenv("JWT_PRIVATE")))

	as := AuthService{
		authClient: JWTClient,
	}

	claims, err := as.authClient.SecureRoute(request)
	if err != nil {
		return resp, nil
	}

	resp.IsAuthorized = true
	resp.Context = claims

	return resp, nil
}

func main() {
	lambda.Start(handler)
}
