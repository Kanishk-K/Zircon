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
	return asynq.NewTask(TypeSummarizeTranscription, payload, asynq.Queue("critical"), asynq.MaxRetry(0), asynq.Timeout(60*time.Minute)), nil
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
	summary, err := p.generateSummary(ctx, &transcriptData)
	if err != nil {
		log.Printf("Failed to generate summary: %v", err)
		return err
	}

	// Upload summary to S3
	log.Printf("Summary: %s", summary)

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

func (p *SummarizeTranscriptionProcess) generateSummary(ctx context.Context, transcriptData *string) (string, error) {
	chatCompletion, err := p.LLMClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(
				"You are an instructor at a university. You are tasked to summarize a lecture. Write it in the perspective as if you were explaining it verbally to a student. Do not include any formatting including latex, images, or code. Do not include a preface of any kind. Provide thorough explanations of each concept discussed. Provide context to all examples.  Begin your response with a high level introduction to the topics that will be discussed.",
			),
			openai.UserMessage(*transcriptData),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})

	if err != nil {
		log.Printf("Failed to generate summary: %v", err)
		return "", err
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
