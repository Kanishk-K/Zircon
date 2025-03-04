package services

import (
	"errors"
	"fmt"
	"log"
	"net/url"
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
	ValidateQuery(jobInfo *models.JobQueueRequest) error
	ScheduleJob(jobInfo *models.JobQueueRequest) error
	CheckStatus(jobStatusInfo *models.JobStatusRequest) (*models.JobStatusResponse, error)
	JobProcessedContent(jobInfo *models.JobStatusRequest) (*models.JobExistingResponse, error)
}

// This contains the content that JobSchedulerService will need
// such as database connections, etc.
type JobSchedulerService struct {
	asynqClient    *asynq.Client
	asynqInspector *asynq.Inspector
	dynamoClient   services.DynamoMethods
}

var validVideoChoices = map[string]bool{
	"subway":    true,
	"minecraft": true,
}

// Creates a new JobScheduler which is required to have
// the methods defined in JobSchedulerServiceMethods
// It is provided inputs to supply the struct with the necessary
// data it needs to function
func NewJobSchedulerService(asynqClient *asynq.Client, asynqInspector *asynq.Inspector, dynamoClient services.DynamoMethods) JobSchedulerServiceMethods {
	return &JobSchedulerService{
		asynqClient:    asynqClient,
		asynqInspector: asynqInspector,
		dynamoClient:   dynamoClient,
	}
}

// In order to validate the request we need to do the following:
// 1. Check that the user is allowed to submit jobs
// 2. Check that the transcript link is from an authorized source
// 3. Check that the background video selected is from one of the available options
// 4. Validate that the form is filled out correctly
func (js *JobSchedulerService) ValidateQuery(jobInfo *models.JobQueueRequest) error {
	// Step 1 [TODO]: Check that the user is allowed to submit jobs

	// Step 2: Ensure the transcript link is from an authorized source (https://cdnapi.kaltura.com)
	url, err := url.Parse(jobInfo.TranscriptLink)
	if err != nil {
		log.Printf("Failed to parse URL: %s", jobInfo.TranscriptLink)
		return err
	}
	if url.Host != "cdnapi.kaltura.com" {
		return fmt.Errorf("transcript link is not from an authorized source %s", jobInfo.TranscriptLink)
	}

	// Step 3: Ensure the background video is from one of the available options
	if _, ok := validVideoChoices[jobInfo.BackgroundVideo]; !ok && (jobInfo.BackgroundVideo != "") {
		return fmt.Errorf("background video is not from an authorized source %s", jobInfo.BackgroundVideo)
	}

	// Step 4: Validate that the form is filled out correctly
	if jobInfo.BackgroundVideo != "" {
		jobInfo.Summarize = true
	}

	return nil
}

func (js *JobSchedulerService) ScheduleJob(jobInfo *models.JobQueueRequest) error {
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

	// Create a new task to generate notes
	if jobInfo.Notes && !jobParams.NotesGenerated {
		notesInfo := &models.NotesInformation{
			EntryID:        jobInfo.EntryID,
			TranscriptLink: jobInfo.TranscriptLink,
		}
		task, err := tasks.NewGenerateNotesTask(notesInfo)
		if err != nil {
			log.Println("Failed to create notes task: ", err)
			return err
		}
		_, err = js.asynqClient.Enqueue(task, asynq.TaskID(fmt.Sprintf("note:%s", jobInfo.EntryID)), asynq.Queue("default"), asynq.Retention(time.Hour))
		switch {
		case errors.Is(err, asynq.ErrTaskIDConflict):
			log.Println("Task already exists, skipping")
		case err != nil:
			log.Println("Failed to enqueue notes task: ", err)
			return err
		}
	}

	// Create a new task to transcribe the video
	if jobInfo.Summarize && !jobParams.SummaryGenerated {
		summarizeInfo := &models.SummarizeInformation{
			EntryID:         jobInfo.EntryID,
			TranscriptLink:  jobInfo.TranscriptLink,
			BackgroundVideo: jobInfo.BackgroundVideo,
		}
		task, err := tasks.NewSummarizeTranscriptionTask(summarizeInfo)
		if err != nil {
			log.Println("Failed to create summarize task: ", err)
			return err
		}
		_, err = js.asynqClient.Enqueue(task, asynq.TaskID(fmt.Sprintf("summary:%s", jobInfo.EntryID)), asynq.Queue("default"), asynq.Retention(time.Hour))
		switch {
		case errors.Is(err, asynq.ErrTaskIDConflict):
			log.Println("Task already exists, skipping")
		case err != nil:
			log.Println("Failed to enqueue summarize task: ", err)
			return err
		}
	} else if jobInfo.BackgroundVideo != "" && !slices.Contains(jobParams.VideosAvailable, jobInfo.BackgroundVideo) {
		videoInfo := &models.GenerateVideoInformation{
			EntryID:           jobInfo.EntryID,
			BackgroundVideo:   jobInfo.BackgroundVideo,
			GenerateSubtitles: !jobParams.SubtitlesGenerated,
		}
		task, err := tasks.NewGenerateVideoTask(videoInfo)
		if err != nil {
			log.Println("Failed to create video task: ", err)
			return err
		}
		_, err = js.asynqClient.Enqueue(task, asynq.TaskID(fmt.Sprintf("video:%s", jobInfo.EntryID)), asynq.Queue("low"), asynq.Retention(time.Hour))
		switch {
		case errors.Is(err, asynq.ErrTaskIDConflict):
			log.Println("Task already exists, skipping")
		case err != nil:
			log.Println("Failed to enqueue video task: ", err)
			return err
		}
	}

	return nil
}

func (js *JobSchedulerService) CheckStatus(jobStatusInfo *models.JobStatusRequest) (*models.JobStatusResponse, error) {
	response := &models.JobStatusResponse{}

	noteJob, err := js.asynqInspector.GetTaskInfo("default", fmt.Sprintf("note:%s", jobStatusInfo.EntryID))
	switch {
	case errors.Is(err, asynq.ErrQueueNotFound):
		return nil, fmt.Errorf("queue %s not found", "default")
	case errors.Is(err, asynq.ErrTaskNotFound):
		response.NotesStatus = 0
	default:
		response.NotesStatus = noteJob.State
	}

	summaryJob, err := js.asynqInspector.GetTaskInfo("default", fmt.Sprintf("summary:%s", jobStatusInfo.EntryID))
	switch {
	case errors.Is(err, asynq.ErrQueueNotFound):
		return nil, fmt.Errorf("queue %s not found", "default")
	case errors.Is(err, asynq.ErrTaskNotFound):
		response.SummarizeStatus = 0
	default:
		response.SummarizeStatus = summaryJob.State
	}

	videoJob, err := js.asynqInspector.GetTaskInfo("low", fmt.Sprintf("video:%s", jobStatusInfo.EntryID))
	switch {
	case errors.Is(err, asynq.ErrQueueNotFound):
		return nil, fmt.Errorf("queue %s not found", "low")
	case errors.Is(err, asynq.ErrTaskNotFound):
		response.VideoStatus = 0
	default:
		response.VideoStatus = videoJob.State
	}

	return response, nil
}

func (js *JobSchedulerService) JobProcessedContent(jobInfo *models.JobStatusRequest) (*models.JobExistingResponse, error) {
	jobParams, err := js.dynamoClient.GetJob(jobInfo.EntryID)
	if err != nil {
		return nil, err
	}
	return &models.JobExistingResponse{
		NotesGenerated:   jobParams.NotesGenerated,
		SummaryGenerated: jobParams.SummaryGenerated,
		VideosAvailable:  jobParams.VideosAvailable,
	}, nil
}
