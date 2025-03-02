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
  "eventName": ["MODIFY"],
  "dynamodb": {
	"NewImage": {
		videosAvailable: [{"exists": true}]
	}
}
*/

func parseVideoInformation(record events.DynamoDBEventRecord) (string, asynq.Option, error) {
	// Two options for video generation
	// 1. This is the first time the video is being generated (medium priority)
	// 2. This is a generation of a new video with a different background (low priority)
	oldImage := record.Change.OldImage
	newImage := record.Change.NewImage
	if oldImage["videosAvailable"].IsNull() {
		// This is the first time the video is being generated
		return newImage["videosAvailable"].StringSet()[0], asynq.Queue("high"), nil
	} else {
		// This is a generation of a new video with a different background
		// Set difference the newImage and oldImage to determine the background video
		oldVideos := oldImage["videosAvailable"].StringSet()
		newVideos := newImage["videosAvailable"].StringSet()
		if len(oldVideos) == len(newVideos) {
			// No new videos have been added
			return "", nil, fmt.Errorf("no new videos have been added")
		}
		oldVideoMap := make(map[string]bool)
		for _, video := range oldVideos {
			oldVideoMap[video] = true
		}
		for _, video := range newVideos {
			if _, ok := oldVideoMap[video]; !ok {
				// This is the new video that was added
				return video, asynq.Queue("low"), nil
			}
		}
	}
	return "", nil, fmt.Errorf("could not determine the background video")
}

func handler(request events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
	resp := events.DynamoDBEventResponse{}
	// Print the request for debugging
	entryID := request.Records[0].Change.NewImage["entryID"].String()
	log.Printf("Processing request for entryID: %s\n", entryID)

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
	backgroundVideo, priority, err := parseVideoInformation(request.Records[0])
	if err != nil {
		log.Printf("Could not determine the background video: %s\n", err)
		resp.BatchItemFailures = []events.DynamoDBBatchItemFailure{
			{
				ItemIdentifier: request.Records[0].EventID,
			},
		}
		return resp, err
	}
	log.Printf("Background video: %s\n", backgroundVideo)
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
