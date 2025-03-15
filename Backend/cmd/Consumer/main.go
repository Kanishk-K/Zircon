package main

import (
	"log"
	"os"

	dynamo "github.com/Kanishk-K/UniteDownloader/Backend/pkg/dynamoClient"
	s3client "github.com/Kanishk-K/UniteDownloader/Backend/pkg/s3Client"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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

	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(region),
		},
	}))
	s3Client := s3client.NewS3Client(awsSession)
	dynamoClient := dynamo.NewDynamoClient(awsSession)

	vg := tasks.NewGenerateVideoProcess(s3Client, dynamoClient)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.VideoGenerationTask, vg.HandleVideoGenerationTask)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
