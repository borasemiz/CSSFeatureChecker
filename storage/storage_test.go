package storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct {
	calledWithBucket string
	calledWithKey    string
	returnErr        error
	returnBody       string
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

func (m *mockS3Client) GetObject(
	_ context.Context,
	input *s3.GetObjectInput,
	_ ...func(*s3.Options),
) (*s3.GetObjectOutput, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	m.calledWithBucket = *input.Bucket
	m.calledWithKey = *input.Key
	return &s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader(m.returnBody)),
	}, nil
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

func TestFromS3_Success(t *testing.T) {
	mock := &mockS3Client{returnBody: "hello csv"}
	body, err := FromS3(context.Background(), mock, "my-bucket", "output.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("unexpected error reading body: %v", err)
	}
	if string(data) != "hello csv" {
		t.Errorf("expected %q, got %q", "hello csv", string(data))
	}
	if mock.calledWithBucket != "my-bucket" {
		t.Errorf("expected bucket %q, got %q", "my-bucket", mock.calledWithBucket)
	}
	if mock.calledWithKey != "output.csv" {
		t.Errorf("expected key %q, got %q", "output.csv", mock.calledWithKey)
	}
}

func TestFromS3_Error(t *testing.T) {
	mock := &mockS3Client{returnErr: errors.New("s3 error")}
	_, err := FromS3(context.Background(), mock, "my-bucket", "output.csv")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
