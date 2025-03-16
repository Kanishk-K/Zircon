package s3client

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Methods interface {
	UploadFile(bucket string, key string, file io.ReadSeeker, filetype string) error
	ReadFile(bucket string, key string) (io.ReadCloser, error)
}

type S3Client struct {
	client *s3.S3
}

func NewS3Client(awsSession *session.Session) S3Methods {
	return &S3Client{
		client: s3.New(awsSession),
	}
}

func (sc *S3Client) UploadFile(bucket string, key string, file io.ReadSeeker, filetype string) error {
	_, err := sc.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(filetype),
		Body:        file,
	})
	if err != nil {
		return err
	}
	return nil
}

func (sc *S3Client) ReadFile(bucket string, key string) (io.ReadCloser, error) {
	resp, err := sc.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
