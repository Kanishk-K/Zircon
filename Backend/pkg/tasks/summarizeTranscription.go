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
	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/dynamoClient/services"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hibiken/asynq"
	"github.com/openai/openai-go"
)

const TypeSummarizeTranscription = "summary:genSummary"

type SummarizeTranscriptionProcess struct {
	LLMClient    *openai.Client
	s3Client     *s3.S3
	dynamoClient services.DynamoMethods
	asynqClient  *asynq.Client
}

func NewSummarizeTranscriptionProcess(client *openai.Client, s3Client *s3.S3, dynamoClient services.DynamoMethods, asynqClient *asynq.Client) *SummarizeTranscriptionProcess {
	return &SummarizeTranscriptionProcess{
		LLMClient:    client,
		s3Client:     s3Client,
		dynamoClient: dynamoClient,
		asynqClient:  asynqClient,
	}
}

func NewSummarizeTranscriptionTask(jobInfo *models.SummarizeInformation) (*asynq.Task, error) {
	payload, err := json.Marshal(jobInfo)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSummarizeTranscription, payload, asynq.Queue("critical"), asynq.MaxRetry(0), asynq.Timeout(60*time.Minute)), nil
}

// We are tasked to create a summary, if we are here there does not exist an element in the database.
func (p *SummarizeTranscriptionProcess) HandleSummarizeTranscriptionTask(ctx context.Context, t *asynq.Task) error {
	data := models.SummarizeInformation{}
	if err := json.Unmarshal(t.Payload(), &data); err != nil {
		return err
	}

	log.Printf("Generating summary for: %s", data.EntryID)

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
	log.Printf("Uploaded summary to S3: %s", data.EntryID)

	// Update DynamoDB
	p.dynamoClient.UpdateSummary(data.EntryID)

	log.Printf("Finished processing: %s", data.EntryID)

	if data.BackgroundVideo != "" {
		// Generate the video
		task, err := NewGenerateVideoTask(&models.GenerateVideoInformation{
			EntryID:           data.EntryID,
			BackgroundVideo:   data.BackgroundVideo,
			GenerateSubtitles: true,
		})
		if err != nil {
			log.Println("Failed to create video task: ", err)
			return err
		}
		_, err = p.asynqClient.Enqueue(task, asynq.TaskID(
			fmt.Sprintf("video:%s", data.EntryID)),
			asynq.Queue("default"),
			asynq.TaskID(fmt.Sprintf("video:%s", data.EntryID)),
			asynq.Retention(time.Hour),
		)
		if err != nil {
			log.Println("Failed to enqueue video task: ", err)
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
				"You are an instructor at a university. You are tasked to teach a lecture. Be extensive in your descriptions. Write it in the perspective as if you were explaining it verbally to a student. Do not include any formatting including latex, images, markdown, lists, or code. Do not include a preface of any kind. Provide thorough explanations of each concept discussed. Provide context to all examples.",
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
