package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	authRouter "github.com/Kanishk-K/UniteDownloader/Backend/pkg/auth-service/router"
	authService "github.com/Kanishk-K/UniteDownloader/Backend/pkg/auth-service/services"
	jobSchedulerRouter "github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/router"
	jobSchedulerService "github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/services"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/authutil"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/dynamoClient/services"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/handlerutil"
)

func main() {
	cwd, _ := os.Getwd()
	fmt.Println("Current working directory:", cwd)
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: No .env file found. Using existing environment variables.")
	}

	// Get the PORT from environment variables
	host := os.Getenv("HOST")
	port := ""
	if host == "" {
		port = ":8080"
		host = "http://localhost"
		fmt.Println("No HOST environment variable detected, defaulting to", host)
		fmt.Println("No PORT environment variable detected, defaulting to", port)
	}

	GoogleOauthConfig := &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	JWTClient := authutil.NewAuthClient([]byte(os.Getenv("JWT_PRIVATE")))

	// Setup dynamo client
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
	dynamoClient := services.NewDynamoClient(awsSession)

	// Establish redis connection, ensure close is called at the end
	redisUrl := os.Getenv("REDIS_URL")
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisUrl})
	defer client.Close()
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisUrl})
	defer inspector.Close()

	// Create an authentication service
	authService := authService.NewAuthService(GoogleOauthConfig, JWTClient)
	authServiceRouter := authRouter.NewAuthServiceRouter(authService)

	// Create a new JobSchedulerService
	jobSchedulerService := jobSchedulerService.NewJobSchedulerService(client, inspector, dynamoClient)
	jobSchedulerRouter := jobSchedulerRouter.NewJobSchedulerRouter(jobSchedulerService, JWTClient)

	handlerutil.RegisterRoutes(
		authServiceRouter, jobSchedulerRouter,
	)

	fmt.Println("Starting server: ", host+port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
