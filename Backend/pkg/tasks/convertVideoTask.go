package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	log.Printf("Tasked to convert video for user %d under video ID %q", p.UserID, p.VideoID)
	// Download the m3u8 file directly from the source URL
	err := convertm3u8ToMp4(p.VideoID, p.SourceURL)
	if err != nil {
		return err
	}
	log.Printf("Successfully converted video for user %d under video ID %q", p.UserID, p.VideoID)

	return nil
}

func convertm3u8ToMp4(videoID string, url string) error {
	// Download the m3u8 file directly from the source URL into a temporary file
	m3u8f, err := os.CreateTemp("", "m3u8-*.m3u8")
	if err != nil {
		log.Printf("Failed to create temporary file: %v", err)
		return err
	}
	defer os.Remove(m3u8f.Name())
	defer m3u8f.Close()
	fmt.Println("Created temporary file: ", m3u8f.Name())

	// Download the m3u8 file
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to download m3u8 file: %v", err)
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(m3u8f, resp.Body)
	if err != nil {
		log.Printf("Failed to write m3u8 file to temporary file: %v", err)
		return err
	}

	// Create space for a new mp4 file.
	mp4f, err := os.CreateTemp("", "mp4-*.mp4")
	if err != nil {
		log.Printf("Failed to create temporary file: %v", err)
		return err
	}
	mp4f.Close()
	defer os.Remove(mp4f.Name())

	// Convert the m3u8 file to mp4
	cmd := exec.Command("ffmpeg", "-protocol_whitelist", "file,http,https,tcp,tls,crypto", "-y", "-i", m3u8f.Name(), "-c", "copy", mp4f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("Failed to convert m3u8 to mp4: %v", err)
		return err
	}

	return nil
}
