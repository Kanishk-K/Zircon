package main

import (
	"fmt"
	"log"
	"os"

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

	redisUrl := os.Getenv("REDIS_URL")
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisUrl})
	defer asynqClient.Close()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 1,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 3,
				"default":  2,
				"low":      1,
			},
			StrictPriority: true,
			// See the godoc for other configuration options
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeSummarizeTranscription, tasks.NewSummarizeTranscriptionProcess(openAIClient, s3Client, asynqClient).HandleSummarizeTranscriptionTask)
	mux.HandleFunc(tasks.TypeTTSSummary, tasks.NewTTSSummaryProcess(pollyClient, s3Client, asynqClient).HandleTTSSummaryTask)
	mux.HandleFunc(tasks.TypeGenerateVideo, tasks.NewGenerateVideoProcess(s3Client).HandleGenerateVideoTask)
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
