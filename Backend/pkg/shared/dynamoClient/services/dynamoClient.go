package services

import (
	"fmt"
	"log"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/dynamoClient/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoMethods interface {
	// GetUser(userID string) (models.UserDocument, error)
	GetJob(entryID string) (*models.JobDocument, error)
	NewJob(newJobData *models.JobDocument) error
	UpdateSummary(entryID string) error
	UpdateSubtitles(entryID string) error // Implicitly done in AddVideo.
	AddVideo(entryID string, videoType string) error
}

type DynamoClient struct {
	client *dynamodb.DynamoDB
}

func NewDynamoClient(awsSession *session.Session) DynamoMethods {
	return &DynamoClient{
		client: dynamodb.New(awsSession),
	}
}

// func (dc *DynamoClient) GetUser(userID string) (models.UserDocument, error){

// }

func (dc *DynamoClient) GetJob(entryID string) (*models.JobDocument, error) {
	result, err := dc.client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]*dynamodb.AttributeValue{
			"entryID": {
				S: aws.String(entryID),
			},
		},
	})
	if err != nil || result.Item == nil {
		log.Printf("Failed to fetch item with entryID (%s) from DynamoDB", entryID)
		return nil, fmt.Errorf("failed to fetch item from dynamodb")
	}
	retVal := &models.JobDocument{}
	if err = dynamodbattribute.UnmarshalMap(result.Item, retVal); err != nil {
		log.Printf("Failed to unmarshal DynamoDB Job document: %v", err)
		return nil, err
	}
	return retVal, nil
}

func (dc *DynamoClient) NewJob(newJobData *models.JobDocument) error {
	av, err := dynamodbattribute.MarshalMap(newJobData)
	if err != nil {
		log.Printf("Failed to marshal data into Attribute Values: %v", err)
		return err
	}

	_, err = dc.client.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String("Jobs"),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(entryID)"),
	})
	if err != nil {
		log.Printf("Failed to update job item: %v", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) UpdateSummary(entryID string) error {
	updateExpression := "SET summaryGenerated = :summaryGenerated"
	_, err := dc.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]*dynamodb.AttributeValue{
			"entryID": {
				S: aws.String(entryID),
			},
		},
		UpdateExpression: aws.String(updateExpression),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":summaryGenerated": {BOOL: aws.Bool(true)},
		},
		ConditionExpression: aws.String("attribute_exists(entryID)"),
	})
	if err != nil {
		log.Printf("Failed to update summary in DynamoDB: %v", err)
		return err
	}

	return nil
}

func (dc *DynamoClient) UpdateSubtitles(entryID string) error {
	updateExpression := "SET subtitlesGenerated = :subtitlesGenerated"
	_, err := dc.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]*dynamodb.AttributeValue{
			"entryID": {
				S: aws.String(entryID),
			},
		},
		UpdateExpression: aws.String(updateExpression),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":subtitlesGenerated": {BOOL: aws.Bool(true)},
		},
		ConditionExpression: aws.String("attribute_exists(entryID)"),
	})
	if err != nil {
		log.Printf("Failed to update summary in DynamoDB: %v", err)
		return err
	}

	return nil
}

func (dc *DynamoClient) AddVideo(entryID string, videoType string) error {
	updateExpression := "ADD videosAvailable :newVideo"
	_, err := dc.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]*dynamodb.AttributeValue{
			"entryID": {
				S: aws.String(entryID),
			},
		},
		UpdateExpression: aws.String(updateExpression),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":newVideo": {SS: aws.StringSlice([]string{videoType})},
		},
		ConditionExpression: aws.String("attribute_exists(entryID)"),
	})
	if err != nil {
		log.Printf("Failed to update summary in DynamoDB: %v", err)
		return err
	}

	return nil
}
