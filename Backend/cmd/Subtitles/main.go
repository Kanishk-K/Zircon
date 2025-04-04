package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	s3client "github.com/Kanishk-K/UniteDownloader/Backend/pkg/s3Client"
	subtitleclient "github.com/Kanishk-K/UniteDownloader/Backend/pkg/subtitleClient"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
)

const BUCKET = "lecture-processor"

type SubtitleGenerationService struct {
	s3Client  s3client.S3Methods
	TTSClient subtitleclient.SubtitleGenerationMethods
}

/*
This path should be protected by the following dynamodb filter:
{
  "eventName": ["MODIFY"],
  "dynamodb": {
    "OldImage": {
      "subtitlesGenerated": {
        "BOOL": [false]
      }
    },
    "NewImage": {
      "subtitlesGenerated": {
        "BOOL": [true]
      }
    }
  }
}
*/

func (sgs SubtitleGenerationService) handler(request events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
	resp := events.DynamoDBEventResponse{}
	// Print the request for debugging
	entryID := request.Records[0].Change.NewImage["entryID"].String()
	log.Printf("Processing request for entryID: %s\n", entryID)

	// Read the summary from S3
	summary, err := sgs.s3Client.ReadFile(BUCKET, fmt.Sprintf("assets/%s/Summary.txt", entryID))
	if err != nil {
		log.Printf("Failed to read summary from S3: %v", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{{
			ItemIdentifier: request.Records[0].EventID,
		}}
		return resp, err
	}
	defer summary.Close()
	summaryBytes, err := io.ReadAll(summary)
	if err != nil {
		log.Printf("Failed to read summary from S3: %v", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{{
			ItemIdentifier: request.Records[0].EventID,
		}}
		return resp, err
	}

	// Generate TTS
	ttsResponse, err := sgs.TTSClient.GenerateTTS(string(summaryBytes))
	if err != nil {
		log.Printf("Failed to generate TTS: %v", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{{
			ItemIdentifier: request.Records[0].EventID,
		}}
		return resp, err
	}
	ttsResponseBytes, err := json.Marshal(ttsResponse)
	if err != nil {
		log.Printf("Failed to marshal TTS response: %v", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{{
			ItemIdentifier: request.Records[0].EventID,
		}}
		return resp, err
	}
	err = sgs.s3Client.UploadFile(BUCKET, fmt.Sprintf("assets/%s/TTSResponse.json", entryID), bytes.NewReader(ttsResponseBytes), "application/json")
	if err != nil {
		log.Printf("Failed to upload TTS response: %v", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{{
			ItemIdentifier: request.Records[0].EventID,
		}}
		return resp, err
	}
	decodedAudio, err := subtitleclient.ConvertB64ToAudio(ttsResponse.Audio)
	if err != nil {
		log.Printf("Failed to decode audio: %v", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{{
			ItemIdentifier: request.Records[0].EventID,
		}}
		return resp, err
	}
	// Upload the audio to S3
	err = sgs.s3Client.UploadFile(BUCKET, fmt.Sprintf("assets/%s/Audio.aac", entryID), bytes.NewReader(decodedAudio), "audio/aac")
	if err != nil {
		log.Printf("Failed to upload audio: %v", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{{
			ItemIdentifier: request.Records[0].EventID,
		}}
		return resp, err
	}
	lines := subtitleclient.GenerateSubtitleLines(ttsResponse.WordTimeStamps)
	assContent := subtitleclient.GenerateASSContent(lines)
	err = sgs.s3Client.UploadFile(BUCKET, fmt.Sprintf("assets/%s/Subtitle.ass", entryID), bytes.NewReader([]byte(assContent)), "application/x-ass")
	if err != nil {
		log.Printf("Failed to upload subtitles: %v", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{{
			ItemIdentifier: request.Records[0].EventID,
		}}
		return resp, err
	}
	return resp, nil
}

func main() {
	// Initialize the service
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
	s3Client := s3client.NewS3Client(awsSession)
	TTSClient := subtitleclient.NewSubtitleClient()
	sgs := SubtitleGenerationService{
		s3Client:  s3Client,
		TTSClient: TTSClient,
	}
	lambda.Start(sgs.handler)
}
