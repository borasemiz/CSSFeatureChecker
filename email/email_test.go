package email

import (
	"strings"
	"testing"

	"github.com/caniuse-scraper/scraper"
)

func TestGetClient_MissingFrom(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "")
	_, err := MakeClient()
	if err == nil {
		t.Fatal("expected error when SMTP_EMAIL_FROM is not set")
	}
}

func TestGetClient_MissingTo(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "noreply@example.com")
	t.Setenv("EMAIL_TO", "")
	_, err := MakeClient()
	if err == nil {
		t.Fatal("expected error when EMAIL_TO is not set")
	}
}

func TestGetClient_MissingHost(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "noreply@example.com")
	t.Setenv("EMAIL_TO", "team@example.com")
	t.Setenv("SMTP_HOST", "")
	_, err := MakeClient()
	if err == nil {
		t.Fatal("expected error when SMTP_HOST is not set")
	}
}

func TestGetClient_MissingPort(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "noreply@example.com")
	t.Setenv("EMAIL_TO", "team@example.com")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "")
	_, err := MakeClient()
	if err == nil {
		t.Fatal("expected error when SMTP_PORT is not set")
	}
}

func TestGetClient_MissingUsername(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "noreply@example.com")
	t.Setenv("EMAIL_TO", "team@example.com")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "321")
	t.Setenv("SMTP_USERNAME", "")

	_, err := MakeClient()
	if err == nil {
		t.Fatal("expected error when SMTP_USERNAME is not set")
	}
}

func TestGetClient_MissingPassword(t *testing.T) {
	t.Setenv("SMTP_EMAIL_FROM", "noreply@example.com")
	t.Setenv("EMAIL_TO", "team@example.com")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "321")
	t.Setenv("SMTP_USERNAME", "yser")
	t.Setenv("SMTP_PASSWORD", "")

	_, err := MakeClient()
	if err == nil {
		t.Fatal("expected error when SMTP_PASSWORD is not set")
	}
}

func TestBuildHeader(t *testing.T) {
	body := string(buildHeader("from@example.com", "to@example.com"))

	for _, header := range []string{"From: from@example.com", "To: to@example.com", "Subject:", "MIME-Version:", "Content-Type:"} {
		if !strings.Contains(body, header) {
			t.Errorf("message missing header %q", header)
		}
	}
}

func TestBuildResult_ContainsFeatureDetails(t *testing.T) {
	features := scraper.Result{
		Title:    "CSS Grid Layout",
		Coverage: 92.50,
		URL:      "https://caniuse.com/css-grid",
	}

	body := string(buildResult(features))

	if !strings.Contains(body, features.Title) {
		t.Errorf("message missing feature title %q", features.Title)
	}

	if !strings.Contains(body, features.URL) {
		t.Errorf("message missing feature URL %q", features.URL)
	}
}
