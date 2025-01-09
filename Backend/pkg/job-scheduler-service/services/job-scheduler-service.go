package services

import (
	"log"
	"slices"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	dynamoModels "github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/dynamoClient/models"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/dynamoClient/services"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"
	"github.com/hibiken/asynq"
)

// These are the methods that the JobSchedulerService should implement
type JobSchedulerServiceMethods interface {
	ScheduleJob(jobInfo *models.JobInformation) error
}

// This contains the content that JobSchedulerService will need
// such as database connections, etc.
type JobSchedulerService struct {
	asynqClient  *asynq.Client
	dynamoClient services.DynamoMethods
}

// Creates a new JobScheduler which is required to have
// the methods defined in JobSchedulerServiceMethods
// It is provided inputs to supply the struct with the necessary
// data it needs to function
func NewJobSchedulerService(asynqClient *asynq.Client, dynamoClient services.DynamoMethods) JobSchedulerServiceMethods {
	return &JobSchedulerService{
		asynqClient:  asynqClient,
		dynamoClient: dynamoClient,
	}
}

func (js *JobSchedulerService) ScheduleJob(jobInfo *models.JobInformation) error {
	// TODO: Validate that the user is allowed to submit jobs
	// TODO: Register the job in Redis

	// Register the job on DynamoDB
	var jobParams *dynamoModels.JobDocument
	jobParams, err := js.dynamoClient.GetJob(jobInfo.EntryID)
	if err != nil {
		log.Printf("Failed to find %s in Job Database using default", jobInfo.EntryID)
		jobParams = &dynamoModels.JobDocument{
			// Default (uncreated values)
			EntryID:            jobInfo.EntryID,
			GeneratedOn:        time.Now().UTC().String(),
			GeneratedBy:        jobInfo.UserID,
			NotesGenerated:     false,
			SummaryGenerated:   false,
			SubtitlesGenerated: false,
			VideosAvailable:    []string{},
		}
		err = js.dynamoClient.NewJob(jobParams)
		if err != nil {
			log.Printf("Failed to upload new Job to Dynamo")
			return err
		}
	}

	// Create a new task to transcribe the video
	if jobInfo.Summarize && !jobParams.SummaryGenerated {
		summarizeInfo := &models.SummarizeInformation{
			EntryID:         jobInfo.EntryID,
			UserID:          jobInfo.UserID,
			TranscriptLink:  jobInfo.TranscriptLink,
			Title:           jobInfo.Title,
			BackgroundVideo: jobInfo.BackgroundVideo,
		}
		task, err := tasks.NewSummarizeTranscriptionTask(summarizeInfo)
		if err != nil {
			log.Println("Failed to create summarize task: ", err)
			return err
		}
		// Enqueue the task
		_, err = js.asynqClient.Enqueue(task)
		if err != nil {
			log.Println("Failed to enqueue summarize task: ", err)
			return err
		}
	} else if jobInfo.BackgroundVideo != "" && !slices.Contains(jobParams.VideosAvailable, jobInfo.BackgroundVideo) {
		videoInfo := &models.GenerateVideoInformation{
			EntryID:         jobInfo.EntryID,
			UserID:          jobInfo.UserID,
			Title:           jobInfo.Title,
			BackgroundVideo: jobInfo.BackgroundVideo,
		}
		task, err := tasks.NewGenerateVideoTask(videoInfo)
		if err != nil {
			log.Println("Failed to create video task: ", err)
			return err
		}
		// Enqueue the task
		_, err = js.asynqClient.Enqueue(task)
		if err != nil {
			log.Println("Failed to enqueue video task: ", err)
			return err
		}
	}

	if jobInfo.Notes {
		log.Printf("Tasked to generate notes for video titled: %s", jobInfo.Title)
	}
	return nil
}
