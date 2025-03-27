package cognitoclient

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type CognitoMethods interface {
	GetEmailFromUsername(username string) (string, error)
}

type CognitoClient struct {
	client *cognitoidentityprovider.Client
}

func NewCognitoClient(awsSession aws.Config) CognitoMethods {
	return &CognitoClient{
		client: cognitoidentityprovider.NewFromConfig(awsSession),
	}
}

func (cc *CognitoClient) GetEmailFromUsername(username string) (string, error) {
	if os.Getenv("COGNITO_POOL") == "" {
		return "", fmt.Errorf("environment variable (COGNITO_POOL) not set")
	}
	userInfo, err := cc.client.AdminGetUser(context.Background(), &cognitoidentityprovider.AdminGetUserInput{
		Username:   aws.String(username),
		UserPoolId: aws.String(os.Getenv("COGNITO_POOL")),
	})
	if err != nil {
		return "", err
	}
	for _, attr := range userInfo.UserAttributes {
		if *attr.Name == "email" {
			return *attr.Value, nil
		}
	}
	return "", fmt.Errorf("email not found")
}
