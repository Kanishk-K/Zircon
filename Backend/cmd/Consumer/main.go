package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"

	"github.com/openai/openai-go"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"

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
	pollyClient := polly.New(
		session.Must(
			session.NewSessionWithOptions(
				session.Options{
					SharedConfigState: session.SharedConfigEnable,
				},
			),
		),
	)

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
	mux.HandleFunc(tasks.TypeSummarizeTranscription, tasks.NewSummarizeTranscriptionProcess(openAIClient).HandleSummarizeTranscriptionTask)
	mux.HandleFunc(tasks.TypeTTSSummary, tasks.NewTTSSummaryProcess(pollyClient).HandleTTSSummaryTask)
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
