package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	apiresponse "github.com/Kanishk-K/UniteDownloader/Backend/pkg/api-response"
	dynamo "github.com/Kanishk-K/UniteDownloader/Backend/pkg/dynamoClient"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/jobutil"
	s3client "github.com/Kanishk-K/UniteDownloader/Backend/pkg/s3Client"
	subtitleclient "github.com/Kanishk-K/UniteDownloader/Backend/pkg/subtitleClient"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

const BUCKET = "lecture-processor"

type SubtitleGenerationService struct {
	dynamoClient dynamo.DynamoMethods
	s3Client     s3client.S3Methods
	TTSClient    subtitleclient.SubtitleGenerationMethods
}

func handler(request jobutil.JobQueueRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		IsBase64Encoded: false,
	}
	// Check if a video is requested, if not then this api need not be called
	if request.BackgroundVideo == "" {
		log.Println("No video requested")
		apiresponse.APIErrorResponse(400, "No video requested", &resp)
		return resp, nil
	}

	// Initialize the service
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(region),
		},
	}))
	dynamoClient := dynamo.NewDynamoClient(awsSession)
	s3Client := s3client.NewS3Client(awsSession)
	TTSClient := subtitleclient.NewSubtitleClient()
	sgs := SubtitleGenerationService{
		dynamoClient: dynamoClient,
		s3Client:     s3Client,
		TTSClient:    TTSClient,
	}
	// Check if a subtitle has already been requested
	err := sgs.dynamoClient.EnableSubtitleGeneration(request.EntryID)
	if err != nil {
		apiresponse.APIErrorResponse(400, "Subtitle already generated", &resp)
		return resp, nil
	}
	// Read the summary from S3
	summary, err := sgs.s3Client.ReadFile(BUCKET, fmt.Sprintf("/assets/%s/Summary.txt", request.EntryID))
	if err != nil {
		sgs.dynamoClient.DisableSubtitleGeneration(request.EntryID)
		apiresponse.APIErrorResponse(500, "Failed to read summary", &resp)
		return resp, nil
	}
	defer summary.Close()
	summaryBytes, err := io.ReadAll(summary)
	if err != nil {
		sgs.dynamoClient.DisableSubtitleGeneration(request.EntryID)
		apiresponse.APIErrorResponse(500, "Failed to read summary", &resp)
		return resp, nil
	}
	ttsResponse, err := sgs.TTSClient.GenerateTTS(string(summaryBytes))
	// ttsResponse, err := sgs.TTSClient.GenerateTTS("Hello world, I am a Zircon test!")
	if err != nil {
		sgs.dynamoClient.DisableSubtitleGeneration(request.EntryID)
		apiresponse.APIErrorResponse(500, "Failed to generate TTS", &resp)
		return resp, nil
	}
	// Upload the tts response to S3
	ttsResponseBytes, err := json.Marshal(ttsResponse)
	if err != nil {
		sgs.dynamoClient.DisableSubtitleGeneration(request.EntryID)
		apiresponse.APIErrorResponse(500, "Failed to marshal tts response", &resp)
		return resp, nil
	}
	err = sgs.s3Client.UploadFile(BUCKET, fmt.Sprintf("/assets/%s/TTSResponse.json", request.EntryID), bytes.NewReader(ttsResponseBytes), "application/json")
	if err != nil {
		sgs.dynamoClient.DisableSubtitleGeneration(request.EntryID)
		apiresponse.APIErrorResponse(500, "Failed to upload tts response", &resp)
		return resp, nil
	}
	decodedAudio, err := subtitleclient.ConvertB64ToAudio(ttsResponse.Audio)
	if err != nil {
		sgs.dynamoClient.DisableSubtitleGeneration(request.EntryID)
		apiresponse.APIErrorResponse(500, "Failed to decode audio", &resp)
		return resp, nil
	}
	// Upload the audio to S3
	err = sgs.s3Client.UploadFile(BUCKET, fmt.Sprintf("/assets/%s/Audio.mp3", request.EntryID), bytes.NewReader(decodedAudio), "audio/mp3")
	if err != nil {
		sgs.dynamoClient.DisableSubtitleGeneration(request.EntryID)
		apiresponse.APIErrorResponse(500, "Failed to upload audio", &resp)
		return resp, nil
	}
	lines := subtitleclient.GenerateSubtitleLines(ttsResponse.WordTimeStamps)
	assContent := subtitleclient.GenerateASSContent(lines)
	err = sgs.s3Client.UploadFile(BUCKET, fmt.Sprintf("/assets/%s/Subtitle.ass", request.EntryID), bytes.NewReader([]byte(assContent)), "application/x-ass")
	if err != nil {
		sgs.dynamoClient.DisableSubtitleGeneration(request.EntryID)
		apiresponse.APIErrorResponse(500, "Failed to upload subtitle", &resp)
		return resp, nil
	}
	apiresponse.APIErrorResponse(200, "Successful Invocation", &resp)
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
