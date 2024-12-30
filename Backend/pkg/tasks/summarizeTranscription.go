package tasks

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/hibiken/asynq"
	"github.com/openai/openai-go"
)

const TypeSummarizeTranscription = "transcribe:summarize"

type SummarizeTranscriptionProcess struct {
	LLMClient *openai.Client
}

func NewSummarizeTranscriptionProcess(client *openai.Client) *SummarizeTranscriptionProcess {
	return &SummarizeTranscriptionProcess{
		LLMClient: client,
	}
}

func NewSummarizeTranscriptionTask(jobInfo *models.SummarizeInformation) (*asynq.Task, error) {
	payload, err := json.Marshal(jobInfo)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSummarizeTranscription, payload, asynq.Queue("critical"), asynq.Timeout(60*time.Minute)), nil
}

func (p *SummarizeTranscriptionProcess) HandleSummarizeTranscriptionTask(ctx context.Context, t *asynq.Task) error {
	data := models.SummarizeInformation{}
	if err := json.Unmarshal(t.Payload(), &data); err != nil {
		return err
	}
	log.Printf("Tasked to summarize transcript titled: %s", data.Title)

	var transcriptData string
	if err := downloadTranscript(&transcriptData, data.TranscriptLink); err != nil {
		log.Printf("Failed to download transcript: %v", err)
		return err
	}

	// Generate the summary
	p.generateSummary(&transcriptData)

	log.Printf("Completed summarizing transcript titled: %s", data.Title)
	return nil
}

func downloadTranscript(transcriptData *string, downloadLink string) error {
	// Download the transcript
	resp, err := http.Get(downloadLink)
	if err != nil {
		log.Printf("Failed to download transcript: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read the transcript
	transcriptDataBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read transcript: %v", err)
		return err
	}
	*transcriptData = string(transcriptDataBytes)

	return nil
}

func (p *SummarizeTranscriptionProcess) generateSummary(transcriptData *string) error {
	log.Printf("Generating summary for transcript: %s", *transcriptData)
	return nil
}
