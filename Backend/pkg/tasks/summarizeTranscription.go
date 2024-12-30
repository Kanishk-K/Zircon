package tasks

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/hibiken/asynq"
)

const TypeSummarizeTranscription = "transcribe:summarize"

func NewTranscribeVideoTask(jobInfo *models.JobInformation) (*asynq.Task, error) {
	payload, err := json.Marshal(jobInfo)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSummarizeTranscription, payload, asynq.Queue("critical"), asynq.Timeout(60*time.Minute)), nil
}

func HandleTranscribeVideoTask(ctx context.Context, t *asynq.Task) error {
	p := models.JobInformation{}
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf("Tasked to convert video titled: %s", p.Title)

	return nil
}

func downloadTranscript(fp *os.File) error {
	// Download the transcript
	return nil
}
