package email

import (
	"strings"
	"testing"

	"github.com/caniuse-scraper/scraper"
)

type mockSender struct {
	capturedFeatures []scraper.Result
}

func (m *mockSender) Send(features []scraper.Result) error {
	m.capturedFeatures = features
	return nil
}

func TestSend_Success(t *testing.T) {
	mock := &mockSender{}
	features := []scraper.Result{{Title: "CSS Grid Layout", Coverage: 92.50}}

	if err := mock.Send(features); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.capturedFeatures) != 1 {
		t.Errorf("expected 1 feature, got %d", len(mock.capturedFeatures))
	}
}

func TestGetClient_MissingFrom(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "")
	_, err := GetClient()
	if err == nil {
		t.Fatal("expected error when SMTP_EMAIL_FROM is not set")
	}
}

func TestGetClient_MissingTo(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "noreply@example.com")
	t.Setenv("EMAIL_TO", "")
	_, err := GetClient()
	if err == nil {
		t.Fatal("expected error when EMAIL_TO is not set")
	}
}

func TestGetClient_MissingHost(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "noreply@example.com")
	t.Setenv("EMAIL_TO", "team@example.com")
	t.Setenv("SMTP_HOST", "")
	_, err := GetClient()
	if err == nil {
		t.Fatal("expected error when SMTP_HOST is not set")
	}
}

func TestGetClient_MissingPort(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "noreply@example.com")
	t.Setenv("EMAIL_TO", "team@example.com")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "")
	_, err := GetClient()
	if err == nil {
		t.Fatal("expected error when SMTP_PORT is not set")
	}
}

func TestGetClient_Success(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "noreply@example.com")
	t.Setenv("EMAIL_TO", "team@example.com")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")

	c, err := GetClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected client, got nil")
	}
}

func TestBuildMessage_ContainsHeaders(t *testing.T) {
	body := string(buildMessage("from@example.com", "to@example.com", []scraper.Result{}))

	for _, header := range []string{"From: from@example.com", "To: to@example.com", "Subject:", "MIME-Version:", "Content-Type:"} {
		if !strings.Contains(body, header) {
			t.Errorf("message missing header %q", header)
		}
	}
}

func TestBuildMessage_ContainsFeatureDetails(t *testing.T) {
	features := []scraper.Result{
		{Title: "CSS Grid Layout", Coverage: 92.50, URL: "https://caniuse.com/css-grid"},
		{Title: "CSS Flexbox", Coverage: 98.10, URL: "https://caniuse.com/flexbox"},
	}

	body := string(buildMessage("from@example.com", "to@example.com", features))

	for _, f := range features {
		if !strings.Contains(body, f.Title) {
			t.Errorf("message missing feature title %q", f.Title)
		}
		if !strings.Contains(body, f.URL) {
			t.Errorf("message missing feature URL %q", f.URL)
		}
	}
}
