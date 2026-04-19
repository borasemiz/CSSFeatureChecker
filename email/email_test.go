package email

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/caniuse-scraper/scraper"
)

type mockSESClient struct {
	capturedInput *sesv2.SendEmailInput
	returnErr     error
}

func (m *mockSESClient) SendEmail(
	_ context.Context,
	input *sesv2.SendEmailInput,
	_ ...func(*sesv2.Options),
) (*sesv2.SendEmailOutput, error) {
	m.capturedInput = input
	return &sesv2.SendEmailOutput{}, m.returnErr
}

func TestSend_Success(t *testing.T) {
	mock := &mockSESClient{}
	features := []scraper.Result{
		{Title: "CSS Grid Layout", Coverage: 92.50, URL: "https://caniuse.com/css-grid"},
	}

	err := Send(context.Background(), mock, "from@example.com", "to@example.com", features)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *mock.capturedInput.FromEmailAddress != "from@example.com" {
		t.Errorf("expected from %q, got %q", "from@example.com", *mock.capturedInput.FromEmailAddress)
	}
	if mock.capturedInput.Destination.ToAddresses[0] != "to@example.com" {
		t.Errorf("expected to %q, got %q", "to@example.com", mock.capturedInput.Destination.ToAddresses[0])
	}
}

func TestSend_BodyContainsFeatureDetails(t *testing.T) {
	mock := &mockSESClient{}
	features := []scraper.Result{
		{Title: "CSS Grid Layout", Coverage: 92.50, URL: "https://caniuse.com/css-grid"},
		{Title: "CSS Flexbox", Coverage: 98.10, URL: "https://caniuse.com/flexbox"},
	}

	if err := Send(context.Background(), mock, "from@example.com", "to@example.com", features); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := *mock.capturedInput.Content.Simple.Body.Text.Data
	for _, f := range features {
		if !strings.Contains(body, f.Title) {
			t.Errorf("body missing feature title %q", f.Title)
		}
		if !strings.Contains(body, f.URL) {
			t.Errorf("body missing feature URL %q", f.URL)
		}
	}
}

func TestSend_Error(t *testing.T) {
	mock := &mockSESClient{returnErr: errors.New("ses error")}
	err := Send(context.Background(), mock, "from@example.com", "to@example.com", []scraper.Result{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
