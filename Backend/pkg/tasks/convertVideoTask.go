package tasks

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/hibiken/asynq"
)

const TypeConvertVideo = "video:convert"

func NewConvertVideoTask(UserID int, VideoID string, SourceURL string) (*asynq.Task, error) {
	payload, err := json.Marshal(models.VideoDownload{
		UserID:    UserID,
		VideoID:   VideoID,
		SourceURL: SourceURL,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeConvertVideo, payload, asynq.Queue("critical"), asynq.MaxRetry(3), asynq.Timeout(60*time.Minute)), nil
}

func HandleConvertVideoTask(ctx context.Context, t *asynq.Task) error {
	p := models.VideoDownload{}
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf("Tasked to convert video %q for user %d under video ID %q", p.SourceURL, p.UserID, p.VideoID)
	// convert video
	return nil
}
