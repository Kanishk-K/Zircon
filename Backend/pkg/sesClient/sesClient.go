package sesclient

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

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

func TitleVideo(backgroundVideo string) string {
	replacedVideo := strings.ReplaceAll(backgroundVideo, "_", " ")
	caser := cases.Title(language.English)
	return caser.String(replacedVideo)
}

func (sc *SESClient) SendEmail(to string, subject string, entryID string, backgroundVideo string) error {
	emailInput := &sesv2.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{
				to,
			},
		},
		FromEmailAddress: aws.String(fmt.Sprintf("Zircon <noreply@%s>", sc.domain)),
		Content: &types.EmailContent{
			Template: &types.Template{
				TemplateName: aws.String("zircon_job_complete_template"),
				TemplateData: aws.String(
					fmt.Sprintf(
						`{ "Subject": "%s", "VideoTitle": "%s", "VideoType": "%s", "EntryID": "%s" }`,
						subject,
						TitleVideo(backgroundVideo),
						backgroundVideo,
						entryID,
					),
				),
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
