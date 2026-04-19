package upload

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct {
	calledWithBucket string
	calledWithKey    string
	returnErr        error
}

func (m *mockS3Client) PutObject(
	_ context.Context,
	input *s3.PutObjectInput,
	_ ...func(*s3.Options),
) (*s3.PutObjectOutput, error) {
	m.calledWithBucket = *input.Bucket
	m.calledWithKey = *input.Key
	return &s3.PutObjectOutput{}, m.returnErr
}

func TestToS3_Success(t *testing.T) {
	mock := &mockS3Client{}
	err := ToS3(context.Background(), mock, "my-bucket", "output.csv", []byte("data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calledWithBucket != "my-bucket" {
		t.Errorf("expected bucket %q, got %q", "my-bucket", mock.calledWithBucket)
	}
	if mock.calledWithKey != "output.csv" {
		t.Errorf("expected key %q, got %q", "output.csv", mock.calledWithKey)
	}
}

func TestToS3_Error(t *testing.T) {
	mock := &mockS3Client{returnErr: errors.New("s3 error")}
	err := ToS3(context.Background(), mock, "my-bucket", "output.csv", []byte("data"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
