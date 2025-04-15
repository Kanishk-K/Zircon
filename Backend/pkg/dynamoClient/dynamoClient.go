package dynamo

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoMethods interface {
	// User modification methods
	CreateUserIfNotExists(userID string) error
	AddScheduledJobToUser(userID string, entryID string) error
	DeregisterJobFromUser(userID string, entryID string) error

	// Job modification methods
	CreateJobIfNotExists(entryID string, title string, generatedBy string) error
	DeleteJobByUser(entryID string, userID string) error
	GenerateSubtitles(entryID string, videoID string) error
	AddVideoToJob(entryID string, videoID string) (*dynamodb.UpdateItemOutput, error)
	RemoveVideoFromJob(entryID string, videoID string) error
	GetJob(entryID string) (*JobDocument, error)

	// Video request methods
	CreateVideoRequest(entryID string, requestedVideo string, requestedBy string) error
	EntityVideoNumber(entryID string) (int, error)
}

type DynamoClient struct {
	client *dynamodb.Client
}

func NewDynamoClient(awsSession aws.Config) DynamoMethods {
	return &DynamoClient{
		client: dynamodb.NewFromConfig(awsSession),
	}
}

func (dc *DynamoClient) CreateUserIfNotExists(userID string) error {
	userData, err := attributevalue.MarshalMap(
		UserDocument{
			UserID:               userID,
			CreatedOn:            time.Now().Format("2006-01-02 15:04:05"),
			PermittedGenerations: 50,
		},
	)
	if err != nil {
		log.Println("Error marshalling user data: ", err)
		return err
	}
	_, err = dc.client.PutItem(context.Background(), &dynamodb.PutItemInput{
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
	_, err := dc.client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Users"),
		Key: map[string]types.AttributeValue{
			"userID": &types.AttributeValueMemberS{
				Value: userID,
			},
		},
		UpdateExpression:    aws.String("ADD scheduledJobs :entryID"),
		ConditionExpression: aws.String("(attribute_not_exists(scheduledJobs) AND permittedGenerations > :zero) OR size(scheduledJobs) < permittedGenerations"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":entryID": &types.AttributeValueMemberSS{
				Value: []string{entryID},
			},
			":zero": &types.AttributeValueMemberN{
				Value: "0",
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
	_, err := dc.client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Users"),
		Key: map[string]types.AttributeValue{
			"userID": &types.AttributeValueMemberS{
				Value: userID,
			},
		},
		UpdateExpression:    aws.String("DELETE scheduledJobs :entryID"),
		ConditionExpression: aws.String("attribute_exists(scheduledJobs)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":entryID": &types.AttributeValueMemberSS{
				Value: []string{entryID},
			},
		},
	})
	if err != nil {
		log.Println("Error updating user data: ", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) CreateJobIfNotExists(entryID string, title string, generatedBy string) error {
	jobData, err := attributevalue.MarshalMap(
		JobDocument{
			EntryID:            entryID,
			Title:              title,
			GeneratedOn:        time.Now().Format("2006-01-02 15:04:05"),
			GeneratedBy:        generatedBy,
			SubtitlesGenerated: false,
		},
	)
	if err != nil {
		log.Println("Error marshalling job data: ", err)
		return err
	}
	_, err = dc.client.PutItem(context.Background(), &dynamodb.PutItemInput{
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
	_, err := dc.client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]types.AttributeValue{
			"entryID": &types.AttributeValueMemberS{
				Value: entryID,
			},
		},
		ConditionExpression: aws.String("generatedBy = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{
				Value: userID,
			},
		},
	})
	if err != nil {
		log.Println("Error deleting job data", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) GenerateSubtitles(entryID string, videoID string) error {
	_, err := dc.client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]types.AttributeValue{
			"entryID": &types.AttributeValueMemberS{
				Value: entryID,
			},
		},
		UpdateExpression:    aws.String("SET subtitlesGenerated = :true"),
		ConditionExpression: aws.String("attribute_exists(entryID)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":true": &types.AttributeValueMemberBOOL{
				Value: true,
			},
		},
	})
	if err != nil {
		log.Printf("Error updating job data: %v", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) AddVideoToJob(entryID string, videoID string) (*dynamodb.UpdateItemOutput, error) {
	update, err := dc.client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]types.AttributeValue{
			"entryID": &types.AttributeValueMemberS{
				Value: entryID,
			},
		},
		UpdateExpression:    aws.String("ADD videosAvailable :videoID"),
		ConditionExpression: aws.String("attribute_exists(entryID)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":videoID": &types.AttributeValueMemberSS{
				Value: []string{videoID},
			},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		log.Printf("Error updating job data: %v", err)
		return nil, err
	}
	return update, nil
}

func (dc *DynamoClient) RemoveVideoFromJob(entryID string, videoID string) error {
	_, err := dc.client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]types.AttributeValue{
			"entryID": &types.AttributeValueMemberS{
				Value: entryID,
			},
		},
		UpdateExpression:    aws.String("DELETE videosAvailable :videoID"),
		ConditionExpression: aws.String("attribute_exists(entryID)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":videoID": &types.AttributeValueMemberSS{
				Value: []string{videoID},
			},
		},
	})
	if err != nil {
		log.Printf("Error updating job data: %v", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) GetJob(entryID string) (*JobDocument, error) {
	result, err := dc.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("Jobs"),
		Key: map[string]types.AttributeValue{
			"entryID": &types.AttributeValueMemberS{
				Value: entryID,
			},
		},
	})
	if err != nil {
		log.Println("Error getting job data: ", err)
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var job JobDocument
	err = attributevalue.UnmarshalMap(result.Item, &job)
	if err != nil {
		log.Println("Error unmarshalling job data: ", err)
		return nil, err
	}
	return &job, nil
}

func (dc *DynamoClient) CreateVideoRequest(entryID string, requestedVideo string, requestedBy string) error {
	videoRequestData, err := attributevalue.MarshalMap(
		VideoRequestDocument{
			EntryID:        entryID,
			RequestedVideo: requestedVideo,
			RequestedOn:    time.Now().Format("2006-01-02 15:04:05"),
			RequestedBy:    requestedBy,
			VideoExpiry:    int(time.Now().Add(time.Hour * 24 * 30).Unix()),
		},
	)
	if err != nil {
		log.Println("Error marshalling video request data: ", err)
		return err
	}
	_, err = dc.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName:           aws.String("VideoRequests"),
		Item:                videoRequestData,
		ConditionExpression: aws.String("attribute_not_exists(entryID) AND attribute_not_exists(requestedVideo)"),
	})
	if err != nil {
		log.Println("Error putting video request data: ", err)
		return err
	}
	return nil
}

func (dc *DynamoClient) EntityVideoNumber(entryID string) (int, error) {
	result, err := dc.client.Query(context.Background(), &dynamodb.QueryInput{
		TableName: aws.String("VideoRequests"),
		KeyConditions: map[string]types.Condition{
			"entryID": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{
						Value: entryID,
					},
				},
			},
		},
	})
	if err != nil {
		log.Println("Error scanning video requests: ", err)
		return 0, err
	}
	return len(result.Items), nil
}
