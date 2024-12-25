package services

import (
	"log"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"
	"github.com/hibiken/asynq"
)

// These are the methods that the JobSchedulerService should implement
type JobSchedulerServiceMethods interface {
	QueueDownload(UserID int, VideoID string, SourceURL string) error
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

func (js *JobSchedulerService) QueueDownload(UserID int, VideoID string, SourceURL string) error {
	task, err := tasks.NewConvertVideoTask(UserID, VideoID, SourceURL)
	if err != nil {
		log.Printf("could not create task: %v", err)
	}
	info, err := js.asynqClient.Enqueue(task)
	if err != nil {
		log.Printf("could not enqueue task: %v", err)
	}
	log.Printf("enqueued task ID %s", info.ID)
	return err
}
