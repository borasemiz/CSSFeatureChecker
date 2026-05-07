package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const csvKey = "output.csv"

type Storage interface {
	Retrieve() (io.ReadCloser, error)
	Save(io.Reader) error
}

type s3Storage struct {
	bucket  string
	config  aws.Config
	client  *s3.Client
	context context.Context
}

func (s *s3Storage) Retrieve() (io.ReadCloser, error) {
	out, err := s.client.GetObject(s.context, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(csvKey),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *s3Storage) Save(reader io.Reader) error {
	_, err := s.client.PutObject(s.context, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(csvKey),
		Body:   reader,
	})
	return err
}

func GetStorage(ctx context.Context) (Storage, error) {
	bucket := os.Getenv("BUCKET_NAME")
	if bucket == "" {
		return nil, fmt.Errorf("BUCKET_NAME env var not set")
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	return &s3Storage{
		bucket:  bucket,
		config:  cfg,
		client:  s3.NewFromConfig(cfg),
		context: ctx,
	}, nil
}
