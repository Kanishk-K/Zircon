package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hibiken/asynq"
	"github.com/openai/openai-go"
)

const TypeSummarizeTranscription = "transcribe:summarize"

type SummarizeTranscriptionProcess struct {
	LLMClient   *openai.Client
	s3Client    *s3.S3
	asynqClient *asynq.Client
}

func NewSummarizeTranscriptionProcess(client *openai.Client, s3Client *s3.S3, asynqClient *asynq.Client) *SummarizeTranscriptionProcess {
	return &SummarizeTranscriptionProcess{
		LLMClient:   client,
		s3Client:    s3Client,
		asynqClient: asynqClient,
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

	// Check if the location already exists in s3. If it does then skip this processing
	_, err := p.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String("lecture-processor"),
		Key:    aws.String(fmt.Sprintf("%s/Summary.txt", data.EntryID)),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				break
			default:
				log.Printf("Failed to check if object exists: %v", err)
				return err
			}
		}
	}

	if err != nil {
		log.Printf("Did not find on S3, tasked to process: %s", data.Title)

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
		_, err = p.s3Client.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String("lecture-processor"),
			Key:         aws.String(fmt.Sprintf("%s/Summary.txt", data.EntryID)),
			ContentType: aws.String("text/plain"),
			Body:        bytes.NewReader([]byte(summary)),
		})
		if err != nil {
			log.Printf("Failed to upload summary to S3: %v", err)
			return err
		}
		log.Printf("Uploaded summary to S3: %s", data.Title)
	} else {
		log.Printf("Found on S3, skipping processing: %s", data.Title)
	}

	// Check if TTS Summary was requested, if so then schedule the task
	if data.BackgroundVideo != "none" {
		jobInfo := &models.TTSSummaryInformation{
			EntryID:         data.EntryID,
			Title:           data.Title,
			BackgroundVideo: data.BackgroundVideo,
		}
		task, err := NewTTSSummaryTask(jobInfo)
		if err != nil {
			log.Println("Failed to create task: ", err)
			return err
		}
		// Enqueue the task
		_, err = p.asynqClient.Enqueue(task)
		if err != nil {
			log.Println("Failed to enqueue task", err)
			return err
		}
	}

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
				"You are an instructor at a university. You are tasked to teach a lecture. Be extensive in your descriptions. Write it in the perspective as if you were explaining it verbally to a student. Do not include any formatting including latex, images, or code. Do not include a preface of any kind. Provide thorough explanations of each concept discussed. Provide context to all examples.  Begin your response with a high level introduction to the topics that will be discussed.",
			),
			openai.UserMessage(*transcriptData),
		}),
		Model: openai.F(openai.ChatModelGPT4oMini),
	})

	if err != nil {
		log.Printf("Failed to generate summary: %v", err)
		return "", err
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
