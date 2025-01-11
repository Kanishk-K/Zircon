package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/job-scheduler-service/models"

	"github.com/Kanishk-K/UniteDownloader/Backend/pkg/shared/dynamoClient/services"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hibiken/asynq"
	"github.com/openai/openai-go"
)

const TypeGenerateNotes = "notes:genNotes"

type GenerateNotesProcess struct {
	LLMClient    *openai.Client
	s3Client     *s3.S3
	dynamoClient services.DynamoMethods
}

func NewGenerateNotesProcess(client *openai.Client, s3Client *s3.S3, dynamoClient services.DynamoMethods) *GenerateNotesProcess {
	return &GenerateNotesProcess{
		LLMClient:    client,
		s3Client:     s3Client,
		dynamoClient: dynamoClient,
	}
}

func NewGenerateNotesTask(jobInfo *models.NotesInformation) (*asynq.Task, error) {
	payload, err := json.Marshal(jobInfo)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeGenerateNotes, payload, asynq.Queue("critical"), asynq.MaxRetry(0), asynq.Timeout(60*time.Minute)), nil
}

func (p *GenerateNotesProcess) HandleGenerateNotesTask(ctx context.Context, t *asynq.Task) error {
	data := models.NotesInformation{}
	if err := json.Unmarshal(t.Payload(), &data); err != nil {
		return err
	}

	log.Printf("Generating notes for: %s", data.EntryID)

	var transcriptData string
	if err := downloadTranscript(&transcriptData, data.TranscriptLink); err != nil {
		log.Printf("Failed to download transcript: %v", err)
		return err
	}

	notes, err := p.GenerateNotes(ctx, &transcriptData)
	if err != nil {
		log.Printf("Failed to generate notes: %v", err)
		return err
	}

	// Upload notes to S3
	_, err = p.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String("lecture-processor"),
		Key:         aws.String(fmt.Sprintf("%s/Notes.md", data.EntryID)),
		ContentType: aws.String("text/markdown"),
		Body:        bytes.NewReader([]byte(notes)),
	})
	if err != nil {
		log.Printf("Failed to upload summary to S3: %v", err)
		return err
	}

	p.dynamoClient.UpdateNotes(data.EntryID)

	log.Printf("Finished processing notes: %s", data.EntryID)

	return nil
}

func (p *GenerateNotesProcess) GenerateNotes(ctx context.Context, transcriptData *string) (string, error) {
	chatCompletion, err := p.LLMClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(
				"You are an instructor at a university. You are provided a transcript of a lecture. You are tasked with generated detailed notes for the lecture entirely in markdown. Do not include any unnecessary information or prefaces. When given a topic explain it extensively as if the audience has no prior knowledge of the subject.",
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
