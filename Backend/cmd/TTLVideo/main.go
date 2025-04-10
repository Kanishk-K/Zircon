package main

import (
	"context"
	"fmt"
	"os"

	dynamo "github.com/Kanishk-K/UniteDownloader/Backend/pkg/dynamoClient"
	s3client "github.com/Kanishk-K/UniteDownloader/Backend/pkg/s3Client"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
)

/*
This path should be protected by the following dynamodb filter:

	{
	  "eventName": ["REMOVE"],
	}
*/
const BUCKET = "lecture-processor"

type TTLVideoService struct {
	dynamoClient dynamo.DynamoMethods
	s3Client     s3client.S3Methods
}

func (tvs *TTLVideoService) handler(request events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
	resp := events.DynamoDBEventResponse{}
	entryID := request.Records[0].Change.OldImage["entryID"].String()
	requestedVideo := request.Records[0].Change.OldImage["requestedVideo"].String()
	if entryID == "" || requestedVideo == "" {
		fmt.Println("EntryID or requested video is empty")
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{
			{
				ItemIdentifier: request.Records[0].EventID,
			},
		}
		return resp, fmt.Errorf("entryID or requested video is empty")
	}
	fmt.Printf("Removing %s video for entryID: %s\n", requestedVideo, entryID)
	err := tvs.dynamoClient.RemoveVideoFromJob(entryID, requestedVideo)
	if err != nil {
		fmt.Printf("Could not remove video from job: %s\n", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{
			{
				ItemIdentifier: request.Records[0].EventID,
			},
		}
		return resp, err
	}
	err = tvs.s3Client.DeleteFile(BUCKET, fmt.Sprintf("assets/%s/%s.mp4", entryID, requestedVideo))
	if err != nil {
		fmt.Printf("Could not delete the video: %s\n", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{
			{
				ItemIdentifier: request.Records[0].EventID,
			},
		}
		return resp, err
	}
	fmt.Printf("Successfully deleted %s video for entryID: %s\n", requestedVideo, entryID)
	return resp, nil
}

func main() {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	awsSession, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		fmt.Println("Failed to load AWS configuration:", err)
		return
	}
	dynamoClient := dynamo.NewDynamoClient(awsSession)
	s3Client := s3client.NewS3Client(awsSession)
	ttlVideoService := TTLVideoService{
		dynamoClient: dynamoClient,
		s3Client:     s3Client,
	}
	lambda.Start(ttlVideoService.handler)
}
