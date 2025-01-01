package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hibiken/asynq"
)

const TypeGenerateVideo = "summary:video"

var validChoices = map[string]bool{
	"subway":    true,
	"minecraft": true,
}

type GenerateVideoProcess struct {
	s3Client *s3.S3
}

func NewGenerateVideoProcess(s3Client *s3.S3) *GenerateVideoProcess {
	return &GenerateVideoProcess{
		s3Client: s3Client,
	}
}

func NewGenerateVideoTask(jobInfo *models.GenerateVideoInformation) (*asynq.Task, error) {
	payload, err := json.Marshal(jobInfo)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeGenerateVideo, payload, asynq.Queue("critical"), asynq.MaxRetry(0), asynq.Timeout(60*time.Minute)), nil
}

func (p *GenerateVideoProcess) HandleGenerateVideoTask(ctx context.Context, t *asynq.Task) error {
	data := models.GenerateVideoInformation{}
	if err := json.Unmarshal(t.Payload(), &data); err != nil {
		log.Printf("Error unmarshalling payload: %v", err)
		return err
	}

	if _, ok := validChoices[data.BackgroundVideo]; !ok {
		log.Printf("Invalid theme choice: %s", data.BackgroundVideo)
		return fmt.Errorf("invalid theme choice: %s", data.BackgroundVideo)
	}

	log.Printf("Processing video generation for entry: %s", data.EntryID)

	// Download Mp3 file from s3
	mp3Result, err := p.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("%s/Audio.mp3", data.EntryID)),
	})
	if err != nil {
		log.Printf("Error downloading mp3 file: %v", err)
		return err
	}

	// Download Subtitle file from s3
	subtitleResult, err := p.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("%s/Subtitles.ass", data.EntryID)),
	})
	if err != nil {
		log.Printf("Error downloading ass file: %v", err)
		return err
	}

	tempDir, err := os.MkdirTemp("", data.EntryID)
	if err != nil {
		log.Printf("Error in creating temp directory %v", err)
		return err
	}
	// defer os.RemoveAll(tempDir)

	// Create location for mp3 file and copy data
	tempMp3Ptr, err := os.CreateTemp(tempDir, "vidGen-*.mp3")
	if err != nil {
		log.Printf("Error in creating temporary mp3 file")
		return err
	}
	defer os.Remove(tempMp3Ptr.Name())
	defer tempMp3Ptr.Close()

	_, err = io.Copy(tempMp3Ptr, mp3Result.Body)
	if err != nil {
		log.Printf("Failure in copying data to temporary mp3 file.")
		return err
	}

	// Create location for subtitle file and copy data
	subtitlePtr, err := os.CreateTemp(tempDir, "vidGen-*.ass")
	if err != nil {
		log.Printf("Error in creating temporary mp3 file")
		return err
	}
	defer os.Remove(subtitlePtr.Name())
	defer subtitlePtr.Close()

	_, err = io.Copy(subtitlePtr, subtitleResult.Body)
	if err != nil {
		log.Printf("Failure in copying data to temporary ass file.")
		return err
	}

	// Generate the video by using ffmpeg.
	// [COMMAND] : ffmpeg -i video.mp4 -i Audio.mp3 -vf "subtitles=Subtitles.ass" -filter:a "atempo=1.25" -c:v libx264 -crf 30 -c:a aac -shortest output.mp4
	// [TODO] : This function does not compress videos a whole bunch. There likely are ways to reduce filesize which can be optimized later
	// [EXPLANATION]
	// `-i video.mp4` : sets first input as the background video
	// `-i Audio.mp3` : sets the second input as the audio
	// `-vf "subtitles=Subtitles.ass"` : visual filter, forces subtitles to be a part of the actual video.
	// `-filter:a "atempo=1.25"` : audio filter, sets the speed of the audio as 1.25
	// `c:v libx264` : sets the video codec as lib264 which is for mp4
	// `c:a aac` : sets the audio codec as aac which is used for mp3
	// `-crf 30` : indicated the compression ratio, higher values are reduced quality (lower bitrate)
	// -shortest : uses the shorter of the mp4 and mp3 to dictate the final length (usually mp4)

	outputPtr, err := os.CreateTemp(tempDir, "output-*.mp4")
	if err != nil {
		log.Printf("Failed to create an output location for the file.")
		return err
	}
	// defer os.Remove(outputPtr.Name())
	defer outputPtr.Close() // No need to actually modify the file, we just need the location for an S3 upload later.

	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Unable to get current working directory.")
		return err
	}
	backgroundVideo := filepath.Join(dir, "static", fmt.Sprintf("%s.mp4", data.BackgroundVideo))

	cmd := exec.Command("ffmpeg", "-y", "-i", backgroundVideo, "-i", filepath.Base(tempMp3Ptr.Name()), "-vf", fmt.Sprintf("subtitles='%s'", filepath.Base(subtitlePtr.Name())), "-filter:a", "atempo=1.25", "-c:v", "libx264", "-c:a", "aac", "-crf", "30", "-shortest", filepath.Base(outputPtr.Name()))
	cmd.Dir = tempDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Printf("Error in running ffmpeg command: %v", err)
		return err
	}

	log.Printf("Completed video generation for entry: %s", data.EntryID)
	return nil
}
