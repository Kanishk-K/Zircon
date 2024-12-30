package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	cwd, _ := os.Getwd()
	fmt.Println("Current working directory:", cwd)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found. Using existing environment variables.")
	}

	var LLM_BASE_URL string
	if os.Getenv(("DEV_ENV")) == "true" {
		LLM_BASE_URL = "https://api.openai.com"
	} else {
		LLM_BASE_URL = "https://api.groq.com/openai/v1"
	}
	openAIClient := openai.NewClient(
		option.WithAPIKey(os.Getenv("LLM_API_KEY")),
		option.WithBaseURL(LLM_BASE_URL),
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
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
