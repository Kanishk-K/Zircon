package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cognitoclient "github.com/Kanishk-K/UniteDownloader/Backend/pkg/cognitoClient"
	dynamo "github.com/Kanishk-K/UniteDownloader/Backend/pkg/dynamoClient"
	s3client "github.com/Kanishk-K/UniteDownloader/Backend/pkg/s3Client"
	sesclient "github.com/Kanishk-K/UniteDownloader/Backend/pkg/sesClient"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hibiken/asynq"
)

func main() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")},
		asynq.Config{
			Concurrency: 1,
			Queues: map[string]int{
				"high":   3,
				"medium": 2,
				"low":    1,
			},
			StrictPriority: true,
		},
	)
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
	dynamoClient := dynamo.NewDynamoClient(awsSession)
	sesClient := sesclient.NewSESClient(awsSession)
	cognitoClient := cognitoclient.NewCognitoClient(awsSession)

	vg := tasks.NewGenerateVideoProcess(s3Client, dynamoClient, sesClient, cognitoClient)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.VideoGenerationTask, vg.HandleVideoGenerationTask)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
