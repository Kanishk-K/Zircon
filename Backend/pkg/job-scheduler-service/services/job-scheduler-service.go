package services

import (
	"log"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/tasks"
	"github.com/hibiken/asynq"
)

// These are the methods that the JobSchedulerService should implement
type JobSchedulerServiceMethods interface {
	QueueJob() error
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

func (js *JobSchedulerService) QueueJob() error {
	task, err := tasks.NewEmailDeliveryTask(42, "some:template:id")
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
