package lambdaclient

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

type LambdaMethods interface {
	InvokeAsyncLambda(functionARN string) error
}

type LambdaClient struct {
	client *lambda.Lambda
}

func NewLambdaClient(session *session.Session) LambdaMethods {
	if os.Getenv("AWS_SAM_LOCAL") == "true" {
		return &LambdaClient{
			client: lambda.New(
				session,
				aws.NewConfig().WithEndpoint("http://host.docker.internal:3001"), // Windows Docker Setup
			),
		}
	} else {
		return &LambdaClient{
			client: lambda.New(session),
		}
	}
}

func (lc *LambdaClient) InvokeAsyncLambda(functionARN string) error {
	if os.Getenv("AWS_SAM_LOCAL") == "true" {
		_, err := lc.client.Invoke(&lambda.InvokeInput{
			FunctionName:   aws.String(functionARN),
			InvocationType: aws.String("RequestResponse"), // Event InvocationType is not supported for SAM local
		})
		if err != nil {
			return err
		}
		return nil
	} else {
		_, err := lc.client.Invoke(&lambda.InvokeInput{
			FunctionName:   aws.String(functionARN),
			InvocationType: aws.String("Event"),
		})
		if err != nil {
			return err
		}
		return nil
	}
}
