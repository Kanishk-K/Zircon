package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	apiresponse "github.com/Kanishk-K/UniteDownloader/Backend/pkg/apiResponse"
	dynamo "github.com/Kanishk-K/UniteDownloader/Backend/pkg/dynamoClient"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/jobutil"
	kalturaclient "github.com/Kanishk-K/UniteDownloader/Backend/pkg/kalturaClient"
	s3client "github.com/Kanishk-K/UniteDownloader/Backend/pkg/s3Client"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/openai/openai-go"
	"golang.org/x/sync/errgroup"
)

const BUCKET = "lecture-processor"

const (
	StatusNew     = "NEW"
	StatusExists  = "EXISTS"
	StatusSkipped = "SKIPPED"
)

type JobSchedulerService struct {
	dynamoClient dynamo.DynamoMethods
	s3Client     s3client.S3Methods
	LLMClient    *openai.Client
	isProd       bool
}

var validVideoChoices = map[string]bool{
	"":               true,
	"subway_surfers": true,
	"minecraft":      true,
	"baking":         true,
	"makeup":         true,
	"sand":           true,
	"soap":           true,
}

func validateRequest(requestBody *jobutil.JobQueueRequest) error {
	// Step 1: Ensure the background video is from one of the available options
	if _, ok := validVideoChoices[requestBody.BackgroundVideo]; !ok {
		return fmt.Errorf("background video is not from an authorized source %s", requestBody.BackgroundVideo)
	}
	// Step 2: Ensure the entry ID is not empty
	if requestBody.EntryID == "" {
		return errors.New("entry ID is empty")
	}
	// Step 3: Ensure the title is not empty
	if requestBody.Title == "" {
		return errors.New("title is empty")
	}

	return nil
}

func downloadTranscript(downloadLink string) (*string, error) {
	// Download the transcript
	resp, err := http.Get(downloadLink)
	if err != nil {
		log.Printf("Failed to download transcript: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the transcript
	transcriptDataBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read transcript: %v", err)
		return nil, err
	}
	transcriptData := string(transcriptDataBytes)
	return &transcriptData, nil
}

func (jss JobSchedulerService) generateNotes(transcriptData *string, entryID string) error {
	chatCompletion, err := jss.LLMClient.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(
				"You are an assistant that generates notes for a lecture from a transcript.\n" +
					"\n" +
					"GOALS:\n" +
					"- Explain content in detail.\n" +
					"- Use simple language.\n" +
					"- Express abstract ideas in an accessible manner.\n" +
					"\n" +
					"IMPORTANT: Exclusively generate notes in markdown format using paragraphs, titles, lists, codeblocks, and tables.\n" +
					"IMPORTANT: Do NOT include images, links, checklists, diagrams, or LaTeX.\n" +
					"IMPORTANT: Be sure to always indicate coding language in code blocks.\n" +
					"\n" +
					"TRANSCRIPT:\n",
			),
			openai.UserMessage(*transcriptData),
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
	})
	if err != nil {
		log.Printf("API call to generate notes failed: %v", err)
		return err
	}
	output := chatCompletion.Choices[0].Message.Content
	err = jss.s3Client.UploadFile(BUCKET, fmt.Sprintf("assets/%s/Notes.md", entryID), bytes.NewReader([]byte(output)), "text/markdown")
	if err != nil {
		log.Printf("Failed to upload notes: %v", err)
		return err
	}
	return nil
}

func (jss JobSchedulerService) generateSummary(transcriptData *string, entryID string) error {
	chatCompletion, err := jss.LLMClient.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(
				"You are an assistant that summarizes university lectures.\n" +
					"\n" +
					"GOALS:\n" +
					"- Explain content in detail.\n" +
					"- Use simple language.\n" +
					"- Express abstract ideas in an accessible manner.\n" +
					"\n" +
					"IMPORTANT: Only respond in plain text. No bullet points, code, or structured sections.\n" +
					"IMPORTANT: Explore each concept thoroughly and step-by-step. You may use approachable analogies to make concepts accessible, if absolutely required.\n" +
					"IMPORTANT: Do not include a preface in your response, just the summary.\n" +
					"\n" +
					"TRANSCRIPT:\n",
			),
			openai.UserMessage(*transcriptData),
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
	})
	if err != nil {
		log.Printf("API call to generate notes failed: %v", err)
		return err
	}
	output := chatCompletion.Choices[0].Message.Content
	err = jss.s3Client.UploadFile(BUCKET, fmt.Sprintf("assets/%s/Summary.txt", entryID), bytes.NewReader([]byte(output)), "text/plain")
	if err != nil {
		log.Printf("Failed to upload notes: %v", err)
		return err
	}
	return nil
}

func (jss JobSchedulerService) handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		IsBase64Encoded: false,
	}
	respBody := make(map[string]interface{})
	requestBody := jobutil.JobQueueRequest{}
	err := json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		apiresponse.APIErrorResponse(500, "Failed to decode request body", &resp)
		return resp, nil
	}
	err = validateRequest(&requestBody)
	if err != nil {
		apiresponse.APIErrorResponse(500, "Submitted request was not valid", &resp)
		return resp, nil
	}
	/*
		Buisness logic goes here
	*/
	var subject string
	if jss.isProd {
		subject = request.RequestContext.Authorizer["claims"].(map[string]any)["cognito:username"].(string)
	} else {
		subject = "DEV USER"
	}
	log.Print("Subject: ", subject)
	// Add the job if it doesn't exist
	err = jss.dynamoClient.CreateJobIfNotExists(requestBody.EntryID, requestBody.Title, subject)
	if err != nil {
		var ccfe *types.ConditionalCheckFailedException
		if errors.As(err, &ccfe) {
			// Job already exists
			respBody["contentGeneration"] = StatusExists
			log.Printf("Job already exists, updating job status should more items be added.")
		} else {
			// Some other error occurred
			apiresponse.APIErrorResponse(500, "Failed to create job", &resp)
			return resp, err
		}
	} else {
		respBody["contentGeneration"] = StatusNew
		err = jss.dynamoClient.AddScheduledJobToUser(subject, requestBody.EntryID)
		if err != nil {
			_ = jss.dynamoClient.DeleteJobByUser(requestBody.EntryID, subject)
			apiresponse.APIErrorResponse(500, "User not permitted to create more requests", &resp)
			return resp, nil
		}

		transcriptLink, err := kalturaclient.GetTranscriptLink(requestBody.EntryID)
		if err != nil {
			_ = jss.dynamoClient.DeleteJobByUser(requestBody.EntryID, subject)
			_ = jss.dynamoClient.DeregisterJobFromUser(subject, requestBody.EntryID)
			apiresponse.APIErrorResponse(500, "Failed to get transcript link", &resp)
		}
		transcriptString, err := downloadTranscript(transcriptLink)
		if err != nil {
			_ = jss.dynamoClient.DeleteJobByUser(requestBody.EntryID, subject)
			_ = jss.dynamoClient.DeregisterJobFromUser(subject, requestBody.EntryID)
			apiresponse.APIErrorResponse(500, "Failed to download transcript", &resp)
			return resp, err
		}

		if len(*transcriptString) > 480000 {
			// Transcript is too long, reject the request
			_ = jss.dynamoClient.DeleteJobByUser(requestBody.EntryID, subject)
			_ = jss.dynamoClient.DeregisterJobFromUser(subject, requestBody.EntryID)
			apiresponse.APIErrorResponse(500, "Transcript is too long", &resp)
			return resp, nil
		}

		var errGroup errgroup.Group
		errGroup.Go(func() error {
			return jss.generateNotes(transcriptString, requestBody.EntryID)
		})
		errGroup.Go(func() error {
			return jss.generateSummary(transcriptString, requestBody.EntryID)
		})
		if err := errGroup.Wait(); err != nil {
			_ = jss.dynamoClient.DeleteJobByUser(requestBody.EntryID, subject)
			_ = jss.dynamoClient.DeregisterJobFromUser(subject, requestBody.EntryID)
			apiresponse.APIErrorResponse(500, "Failed to generate notes or summary", &resp)
			return resp, err
		}
	}

	if requestBody.BackgroundVideo != "" {
		respBody["videoGeneration"] = StatusNew
		// Request subtitle generation
		err = jss.dynamoClient.GenerateSubtitles(requestBody.EntryID, requestBody.BackgroundVideo)
		if err != nil {
			apiresponse.APIErrorResponse(500, "Failed to update job status", &resp)
			return resp, err
		}

		// Request video generation
		err = jss.dynamoClient.CreateVideoRequest(requestBody.EntryID, requestBody.BackgroundVideo, subject)
		if err != nil {
			var ccfe *types.ConditionalCheckFailedException
			if errors.As(err, &ccfe) {
				// Video requested previously
				respBody["videoGeneration"] = StatusExists
			} else {
				apiresponse.APIErrorResponse(500, "Failed to create new video request", &resp)
				return resp, err
			}
		}
	} else {
		respBody["videoGeneration"] = StatusSkipped
	}

	apiresponse.APISuccessResponse(respBody, &resp)

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

	LLMClient := openai.NewClient()

	jss := JobSchedulerService{
		dynamoClient: dynamoClient,
		s3Client:     s3Client,
		LLMClient:    LLMClient,
	}

	jss.isProd = os.Getenv("AWS_SAM_LOCAL") != "true"

	lambda.Start(jss.handler)
}
