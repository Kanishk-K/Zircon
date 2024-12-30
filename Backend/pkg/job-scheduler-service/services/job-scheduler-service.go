package services

import (
	"log"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"
	"github.com/hibiken/asynq"
)

// These are the methods that the JobSchedulerService should implement
type JobSchedulerServiceMethods interface {
	ScheduleJob(jobInfo *models.JobInformation) error
	// ViewJobStatus() error
}

// This contains the content that JobSchedulerService will need
// such as database connections, etc.
type JobSchedulerService struct {
	asynqClient *asynq.Client
}

// Creates a new JobScheduler which is required to have
// the methods defined in JobSchedulerServiceMethods
// It is provided inputs to supply the struct with the necessary
// data it needs to function
func NewJobSchedulerService(asynqClient *asynq.Client) JobSchedulerServiceMethods {
	return &JobSchedulerService{
		asynqClient: asynqClient,
	}
}

func (js *JobSchedulerService) ScheduleJob(jobInfo *models.JobInformation) error {
	// TODO: Validate that the user is allowed to submit jobs
	// TODO: Register the job in Redis
	// Create a new task to transcribe the video
	task, err := tasks.NewTranscribeVideoTask(jobInfo)
	if err != nil {
		log.Println("Failed to create task: ", err)
		return err
	}
	// Enqueue the task
	_, err = js.asynqClient.Enqueue(task)
	if err != nil {
		log.Println("Failed to enqueue task: ", err)
		return err
	}
	return nil

}
