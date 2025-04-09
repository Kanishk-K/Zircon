package s3client

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Methods interface {
	UploadFile(bucket string, key string, file io.ReadSeeker, filetype string) error
	ReadFile(bucket string, key string) (io.ReadCloser, error)
	DeleteFile(bucket string, key string) error
}

type S3Client struct {
	client *s3.Client
}

func NewS3Client(awsSession aws.Config) S3Methods {
	return &S3Client{
		client: s3.NewFromConfig(awsSession),
	}
}

func (sc *S3Client) UploadFile(bucket string, key string, file io.ReadSeeker, filetype string) error {
	_, err := sc.client.PutObject(context.Background(), &s3.PutObjectInput{
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
	resp, err := sc.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (sc *S3Client) DeleteFile(bucket string, key string) error {
	_, err := sc.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}
