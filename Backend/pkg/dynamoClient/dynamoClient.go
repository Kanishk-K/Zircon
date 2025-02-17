package dynamo

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoMethods interface {
	CreateUserIfNotExists(userID string, name string) error
}

type DynamoClient struct {
	client *dynamodb.DynamoDB
}

func NewDynamoClient(awsSession *session.Session) DynamoMethods {
	return &DynamoClient{
		client: dynamodb.New(awsSession),
	}
}

func (dc *DynamoClient) CreateUserIfNotExists(userID string, name string) error {
	userData, err := dynamodbattribute.MarshalMap(
		UserDocument{
			UserID:               userID,
			Name:                 name,
			CreatedOn:            time.Now().Format("2006-01-02 15:04:05"),
			PermittedGenerations: 5,
			ScheduledJobs:        []string{},
		},
	)
	if err != nil {
		log.Println("Error marshalling user data: ", err)
		return err
	}
	_, err = dc.client.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String("Users"),
		Item:                userData,
		ConditionExpression: aws.String("attribute_not_exists(userID)"),
	})
	if err != nil {
		log.Println("Error putting user data: ", err)
		return err
	}
	return nil
}
