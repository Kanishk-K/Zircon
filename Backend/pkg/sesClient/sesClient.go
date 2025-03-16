package sesclient

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SESMethods interface {
	SendEmail(to string, subject string, entryID string, backgroundVideo string) error
}

type SESClient struct {
	client *sesv2.Client
	domain string
}

func NewSESClient(awsSession aws.Config) SESMethods {
	return &SESClient{
		client: sesv2.NewFromConfig(awsSession),
		domain: os.Getenv("DOMAIN"),
	}
}

func (sc *SESClient) SendEmail(to string, subject string, entryID string, backgroundVideo string) error {
	caser := cases.Title(language.English)
	emailInput := &sesv2.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{
				fmt.Sprintf("%s@umn.edu", to),
			},
		},
		FromEmailAddress: aws.String(fmt.Sprintf("Zircon <noreply@%s>", sc.domain)),
		Content: &types.EmailContent{
			Template: &types.Template{
				TemplateName: aws.String("zircon_job_complete_template"),
				TemplateData: aws.String(fmt.Sprintf(`{ "Subject": "%s", "BackgroundVideo": "%s"}`, subject, caser.String(backgroundVideo))),
			},
		},
	}
	_, err := sc.client.SendEmail(context.Background(), emailInput)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}

	return nil
}
