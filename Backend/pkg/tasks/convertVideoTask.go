package tasks

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/hibiken/asynq"
)

const TypeTranscribeVideo = "video:transcribe"

func NewTranscribeVideoTask(jobInfo *models.JobInformation) (*asynq.Task, error) {
	payload, err := json.Marshal(jobInfo)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTranscribeVideo, payload, asynq.Queue("critical"), asynq.Timeout(60*time.Minute)), nil
}

func HandleTranscribeVideoTask(ctx context.Context, t *asynq.Task) error {
	p := models.JobInformation{}
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf("Tasked to convert video titled: %s", p.Title)

	// Create a temporary file to store the mp3 file
	mp3f, err := os.CreateTemp("", "mp3-*.mp3")
	if err != nil {
		log.Printf("Failed to create temporary file: %v", err)
		return err
	}
	// We close the file as FFMPEG will just overwrite the file location
	// We still want to keep the location as we will upload the contents for API processing.
	err = mp3f.Close()
	if err != nil {
		log.Printf("Failed to close temporary file: %v", err)
		os.Remove(mp3f.Name())
	}
	defer os.Remove(mp3f.Name())
	err = convertMP4toMP3(mp3f, p.DownloadLink)
	if err != nil {
		log.Printf("Failed to convert mp4 to mp3: %v", err)
		return err
	}
	// Submit the mp3 file to the API for processing

	log.Printf("Successfully converted video titled: %s", p.Title)

	return nil
}

func convertMP4toMP3(MP3fp *os.File, downloadLink string) error {
	// Create a temporary file to store the mp4 file
	mp4f, err := os.CreateTemp("", "mp4-*.mp4")
	if err != nil {
		log.Printf("Failed to create temporary file: %v", err)
		return err
	}
	defer os.Remove(mp4f.Name())
	defer mp4f.Close()
	// Download the mp4 file
	resp, err := http.Get(downloadLink)
	if err != nil {
		log.Printf("Failed to download mp4 file: %v", err)
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(mp4f, resp.Body)
	if err != nil {
		log.Printf("Failed to write mp4 file to temporary file: %v", err)
		return err
	}

	// Convert the mp4 file to mp3
	cmd := exec.Command("ffmpeg", "-y", "-i", mp4f.Name(), "-vn", MP3fp.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("Failed to convert mp4 to mp3: %v", err)
		return err
	}

	return nil
}
