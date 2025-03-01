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
	// User modification methods
	CreateUserIfNotExists(userID string, name string) error
	AddScheduledJobToUser(userID string, entryID string) error
	DeregisterJobFromUser(userID string, entryID string) error

	// Job modification methods
	CreateJobIfNotExists(entryID string, generatedBy string) error
	DeleteJobByUser(entryID string, userID string) error
	UpdateJobStatus(entryID string, videoID string) error
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

func (dc *DynamoClient) AddScheduledJobToUser(userID string, entryID string) error {
	_, err := dc.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"userID": {
				S: aws.String(userID),
			},
		},
		UpdateExpression:    aws.String("ADD scheduledJobs :entryID"),
		ConditionExpression: aws.String("(attribute_not_exists(scheduledJobs) AND permittedGenerations > :zero) OR size(scheduledJobs) < permittedGenerations"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":entryID": {
				SS: aws.StringSlice([]string{entryID}),
			},
			":zero": {
				// Prescedence rules in condition expressions ensures that OR does not take precedence
				N: aws.String("0"),
			},
		},
	})
	if err != nil {
		log.Println("Error updating user data: ", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) DeregisterJobFromUser(userID string, entryID string) error {
	_, err := dc.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Users"),
		Key: map[string]*dynamodb.AttributeValue{
			"userID": {
				S: aws.String(userID),
			},
		},
		UpdateExpression:    aws.String("DELETE scheduledJobs :entryID"),
		ConditionExpression: aws.String("attribute_exists(scheduledJobs)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":entryID": {
				SS: aws.StringSlice([]string{entryID}),
			},
		},
	})
	if err != nil {
		log.Println("Error updating user data: ", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) CreateJobIfNotExists(entryID string, generatedBy string) error {
	jobData, err := dynamodbattribute.MarshalMap(
		JobDocument{
			EntryID:            entryID,
			GeneratedOn:        time.Now().Format("2006-01-02 15:04:05"),
			GeneratedBy:        generatedBy,
			SubtitlesGenerated: false,
			VideosAvailable:    []string{},
		},
	)
	if err != nil {
		log.Println("Error marshalling job data: ", err)
		return err
	}
	_, err = dc.client.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String("Jobs"),
		Item:                jobData,
		ConditionExpression: aws.String("attribute_not_exists(entryID)"),
	})
	if err != nil {
		log.Println("Error putting job data: ", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) DeleteJobByUser(entryID string, userID string) error {
	_, err := dc.client.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]*dynamodb.AttributeValue{
			"entryID": {
				S: aws.String(entryID),
			},
		},
		ConditionExpression: aws.String("generatedBy = :userID"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userID": {
				S: aws.String(userID),
			},
		},
	})
	if err != nil {
		log.Println("Error deleting job data", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) UpdateJobStatus(entryID string, videoID string) error {
	if videoID == "" {
		return nil
	}
	_, err := dc.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]*dynamodb.AttributeValue{
			"entryID": {
				S: aws.String(entryID),
			},
		},
		UpdateExpression:    aws.String("ADD videosAvailable :videoID SET subtitlesGenerated = :true"),
		ConditionExpression: aws.String("attribute_exists(entryID)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":videoID": {
				SS: aws.StringSlice([]string{videoID}),
			},
			":true": {
				BOOL: aws.Bool(true),
			},
		},
	})
	if err != nil {
		log.Printf("Error updating job data: %v", err)
		return err
	}
	return nil
}
