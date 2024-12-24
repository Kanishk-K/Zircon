package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"

	jobSchedulerRouter "github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/router"
	jobSchedulerService "github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/services"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/handlerutil"
)

func main() {
	cwd, _ := os.Getwd()
	fmt.Println("Current working directory:", cwd)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found. Using existing environment variables.")
	}

	// Get the PORT from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		fmt.Println("No PORT environment variable detected, defaulting to", port)
	}

	// Establish redis connection, ensure close is called at the end
	redisUrl := os.Getenv("REDIS_URL")
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisUrl})
	defer client.Close()

	// Create a new JobSchedulerService
	jobSchedulerService := jobSchedulerService.NewJobSchedulerService(client)
	jobSchedulerRouter := jobSchedulerRouter.NewJobSchedulerRouter(jobSchedulerService)

	handlerutil.RegisterRoutes(
		jobSchedulerRouter,
	)

	fmt.Println("Starting server on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
