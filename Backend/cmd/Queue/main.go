package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hibiken/asynq"
)

type QueueService struct {
	jobQueue *asynq.Client
}

/*
This path should be protected by the following dynamodb filter:
{
  "eventName": ["INSERT"],
}
*/

func handler(request events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
	resp := events.DynamoDBEventResponse{}
	// Print the request for debugging
	entryID := request.Records[0].Change.NewImage["entryID"].String()
	backgroundVideo := request.Records[0].Change.NewImage["requestedVideo"].String()
	if entryID == "" || backgroundVideo == "" {
		log.Printf("EntryID or background video is empty")
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{
			{
				ItemIdentifier: request.Records[0].EventID,
			},
		}
		return resp, fmt.Errorf("entryID or background video is empty")
	}
	priority := asynq.Queue("low")
	log.Printf("Processing request for entryID: %s\n", entryID)
	log.Printf("Request: %v\n", request.Records[0])

	// Initialize the service
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_URL")})
	if client == nil {
		log.Printf("Could not connect to Redis")
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{
			{
				ItemIdentifier: request.Records[0].EventID,
			},
		}
		return resp, fmt.Errorf("could not connect to Redis")
	}
	defer client.Close()
	log.Printf("Background video: %s\n", backgroundVideo)
	log.Printf("Priority: %s\n", priority)
	qs := QueueService{jobQueue: client}
	task, err := tasks.NewVideoGenerationTask(entryID, backgroundVideo)
	if err != nil {
		log.Printf("Could not create the task: %s\n", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{
			{
				ItemIdentifier: request.Records[0].EventID,
			},
		}
		return resp, err
	}
	_, err = qs.jobQueue.Enqueue(
		task,
		priority,
		asynq.MaxRetry(0),
		asynq.TaskID(fmt.Sprintf("%s:%s", entryID, backgroundVideo)),
		asynq.Unique(time.Hour*24),
		asynq.Retention(time.Hour*24*7),
	)
	if err != nil {
		log.Printf("Could not enqueue the task: %s\n", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{
			{
				ItemIdentifier: request.Records[0].EventID,
			},
		}
		return resp, nil
	}
	log.Printf("Enqueued the task for entryID: %s\n", entryID)
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
