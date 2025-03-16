package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"

	dynamo "github.com/Kanishk-K/UniteDownloader/Backend/pkg/dynamoClient"
	s3client "github.com/Kanishk-K/UniteDownloader/Backend/pkg/s3Client"
	sesclient "github.com/Kanishk-K/UniteDownloader/Backend/pkg/sesClient"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hibiken/asynq"
)

const BUCKET = "lecture-processor"
const VideoGenerationTask = "videoGeneration"

type VideoGenerationPayload struct {
	EntryID         string `json:"entryID"`
	RequestedBy     string `json:"requestedBy"`
	BackgroundVideo string `json:"backgroundVideo"`
}

type GenerateVideoProcess struct {
	s3Client     s3client.S3Methods
	dynamoClient dynamo.DynamoMethods
	sesClient    sesclient.SESMethods
}

func NewGenerateVideoProcess(s3Client s3client.S3Methods, dynamoClient dynamo.DynamoMethods, sesClient sesclient.SESMethods) *GenerateVideoProcess {
	return &GenerateVideoProcess{s3Client, dynamoClient, sesClient}
}

func NewVideoGenerationTask(entryID string, requestedBy string, backgroundVideo string) (*asynq.Task, error) {
	taskInfo := VideoGenerationPayload{
		EntryID:         entryID,
		RequestedBy:     requestedBy,
		BackgroundVideo: backgroundVideo,
	}
	payload, err := json.Marshal(taskInfo)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(VideoGenerationTask, payload), nil
}

func (p *GenerateVideoProcess) HandleVideoGenerationTask(ctx context.Context, t *asynq.Task) error {
	var payload VideoGenerationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	workingDir, err := os.MkdirTemp("", payload.EntryID)
	if err != nil {
		log.Printf("Error creating temp directory: %v", err)
		return err
	}
	defer os.RemoveAll(workingDir)

	mp3Fp, err := os.CreateTemp(workingDir, "audio-*.mp3")
	if err != nil {
		log.Printf("Error creating temp audio file: %v", err)
		return err
	}
	defer mp3Fp.Close()
	defer os.Remove(mp3Fp.Name())

	subtitlesFp, err := os.CreateTemp(workingDir, "subtitles-*.ass")
	if err != nil {
		log.Printf("Error creating temp subtitles file: %v", err)
	}
	defer subtitlesFp.Close()
	defer os.Remove(subtitlesFp.Name())

	mp3Bytes, err := p.s3Client.ReadFile(BUCKET, fmt.Sprintf("/assets/%s/Audio.mp3", payload.EntryID))
	if err != nil {
		log.Printf("Error reading audio file from S3: %v", err)
		return err
	}
	_, err = io.Copy(mp3Fp, mp3Bytes)
	if err != nil {
		log.Printf("Error putting bytes into mp3 file: %v", err)
		return err
	}
	err = mp3Bytes.Close()
	if err != nil {
		log.Printf("Failed to close mp3 reader: %v", err)
		return err
	}

	subtitleBytes, err := p.s3Client.ReadFile(BUCKET, fmt.Sprintf("/assets/%s/Subtitle.ass", payload.EntryID))
	if err != nil {
		log.Printf("Error reading subtitle file from S3: %v", err)
		return err
	}
	_, err = io.Copy(subtitlesFp, subtitleBytes)
	if err != nil {
		log.Printf("Error putting bytes into subtitle file: %v", err)
		return err
	}
	err = subtitleBytes.Close()
	if err != nil {
		log.Printf("Failed to close subtitle reader: %v", err)
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current working directory: %v", err)
		return err
	}
	backgroundVideo := filepath.Join(dir, "static", fmt.Sprintf("%s.mp4", payload.BackgroundVideo))
	logoPng := filepath.Join(dir, "static", "logo.png")
	minuteOffset := rand.IntN(30)
	secondsOffset := rand.IntN(60)
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-stream_loop",
		"-1",
		"-ss",
		fmt.Sprintf("00:%02d:%02d", minuteOffset, secondsOffset),
		"-i",
		backgroundVideo,
		"-i",
		filepath.Base(mp3Fp.Name()),
		"-i",
		logoPng,
		"-filter_complex",
		// Commented is the FFMPEG 7.1 version. Docker uses 6.1
		// fmt.Sprintf("ass='%s'[subs];[2]format=rgba,colorchannelmixer=aa=0.3[logo];[logo][0]scale=w=oh*dar:h=rh/12[logo_scaled];[subs][logo_scaled]overlay=x=W-w-10:y=10[output];", filepath.Base(subtitlesFp.Name())),
		fmt.Sprintf("ass='%s'[subs];[2]format=rgba,colorchannelmixer=aa=0.3[logo];[subs][logo]overlay=main_w-overlay_w-10:10[output];", filepath.Base(subtitlesFp.Name())),
		"-map",
		"[output]",
		"-map",
		"1:a",
		"-c:v",
		"libx264",
		"-c:a",
		"copy",
		"-crf",
		"30",
		"-shortest",
		"output.mp4",
	)
	cmd.Dir = workingDir
	// Put cmd errors into stdout
	// cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	log.Printf("Generating video for %s", payload.EntryID)
	err = cmd.Run()
	if err != nil {
		log.Printf("Error in running ffmpeg command: %v", err)
		return err
	}

	outputFp, err := os.Open(filepath.Join(workingDir, "output.mp4"))
	if err != nil {
		log.Printf("Failed to open output Mp4 file: %v", err)
		return err
	}
	defer outputFp.Close()
	defer os.Remove(outputFp.Name())

	err = p.s3Client.UploadFile(BUCKET, fmt.Sprintf("/assets/%s/%s.mp4", payload.EntryID, payload.BackgroundVideo), outputFp, "video/mp4")
	if err != nil {
		log.Printf("Failed to upload video to S3: %v", err)
		return err
	}

	log.Printf("Completed video for %s", payload.EntryID)

	updated, err := p.dynamoClient.AddVideoToJob(payload.EntryID, payload.BackgroundVideo)
	if err != nil {
		log.Printf("Failed to update job data: %v", err)
		return err
	}

	err = p.sesClient.SendEmail(
		payload.RequestedBy,
		updated.Attributes["title"].(*types.AttributeValueMemberS).Value,
		payload.EntryID,
		payload.BackgroundVideo,
	)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}
