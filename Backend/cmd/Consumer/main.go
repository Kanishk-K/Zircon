package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/dynamoClient/services"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"

	"github.com/openai/openai-go"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	cwd, _ := os.Getwd()
	fmt.Println("Current working directory:", cwd)
	if err := godotenv.Load(
		"../.env",
	); err != nil {
		log.Println("Warning: No .env file found. Using existing environment variables.")
	}

	openAIClient := openai.NewClient()

	// This searches for the credentials file in the default location ($HOME/.aws/credentials)
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
	pollyClient := polly.New(awsSession)
	s3Client := s3.New(awsSession)
	dynamoClient := services.NewDynamoClient(awsSession)

	redisUrl := os.Getenv("REDIS_URL")
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisUrl})
	defer asynqClient.Close()

	var srv *asynq.Server
	if os.Getenv("API_ONLY") == "true" {
		srv = asynq.NewServer(
			asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")},
			asynq.Config{
				// Specify how many concurrent workers to use
				Concurrency: 20,
				// We do not want to process any video tasks.
				Queues: map[string]int{"default": 2},
			},
		)
	} else {
		srv = asynq.NewServer(
			asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")},
			asynq.Config{
				// Specify how many concurrent workers to use
				Concurrency: 1,
				// We are okay processing both video and non-video tasks. However, do one at a time.
				Queues: map[string]int{"default": 2, "low": 1},
			},
		)
	}

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeGenerateNotes, tasks.NewGenerateNotesProcess(openAIClient, s3Client, dynamoClient).HandleGenerateNotesTask)
	mux.HandleFunc(tasks.TypeSummarizeTranscription, tasks.NewSummarizeTranscriptionProcess(openAIClient, s3Client, dynamoClient, asynqClient).HandleSummarizeTranscriptionTask)
	if os.Getenv("API_ONLY") != "true" {
		mux.HandleFunc(tasks.TypeGenerateVideo, tasks.NewGenerateVideoProcess(pollyClient, s3Client, dynamoClient).HandleGenerateVideoTask)
	}
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
