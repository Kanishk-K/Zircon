package sesclient

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SESMethods interface {
	SendEmail(to string, subject string, entryID string, backgroundVideo string) error
}

type SESClient struct {
	client *ses.SES
	domain string
}

func NewSESClient(awsSession *session.Session) SESMethods {
	return &SESClient{
		client: ses.New(awsSession),
		domain: os.Getenv("DOMAIN"),
	}
}

func (sc *SESClient) SendEmail(to string, subject string, entryID string, backgroundVideo string) error {
	caser := cases.Title(language.English)
	emailInput := &ses.SendTemplatedEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(fmt.Sprintf("%s@umn.edu", to)),
			},
		},
		Template:     aws.String("zircon_job_complete_template"),
		TemplateData: aws.String(fmt.Sprintf(`{ "Subject": "%s", "BackgroundVideo": "%s"}`, subject, caser.String(backgroundVideo))),
		Source:       aws.String(fmt.Sprintf("Zircon <noreply@%s>", sc.domain)),
	}
	_, err := sc.client.SendTemplatedEmail(emailInput)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}

	return nil
}
