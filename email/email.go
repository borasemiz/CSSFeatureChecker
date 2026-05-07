package email

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"strings"

	"github.com/caniuse-scraper/scraper"
)

type Sender interface {
	Send() error
	WriteResult(result scraper.Result) error
}

type client struct {
	smtpClient *smtp.Client
	writer     io.WriteCloser
}

func (c *client) WriteResult(result scraper.Result) error {
	_, err := c.writer.Write(buildResult(result))
	return err
}

func (c *client) Send() error {
	if err := c.writer.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	return c.smtpClient.Quit()
}

func MakeClient() (Sender, error) {
	credentials, err := getCredentials()
	if err != nil {
		return nil, err
	}

	smtpClient, err := initSMTPClient(credentials)
	if err != nil {
		return nil, err
	}

	w, err := smtpClient.Data()
	if err != nil {
		return nil, fmt.Errorf("smtp data: %w", err)
	}

	if _, err := w.Write(buildHeader(credentials.from, credentials.to)); err != nil {
		return nil, fmt.Errorf("smtp write: %w", err)
	}

	return &client{
		writer:     w,
		smtpClient: smtpClient,
	}, nil
}

func buildHeader(from, to string) []byte {
	var sb strings.Builder
	fmt.Fprintf(&sb, "From: %s\r\n", from)
	fmt.Fprintf(&sb, "To: %s\r\n", to)
	sb.WriteString("Subject: CSS Features Newly Above 90% Coverage\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString("The following CSS features have newly crossed 90% browser coverage:\n\n")
	return []byte(sb.String())
}

func buildResult(feature scraper.Result) []byte {
	var sb strings.Builder
	fmt.Fprintf(
		&sb,
		"- %s: %.2f%%\n  %s\n\n",
		feature.Title,
		feature.Coverage,
		feature.URL,
	)
	return []byte(sb.String())
}

func initSMTPClient(cr *credentials) (*smtp.Client, error) {
	smtpClient, err := smtp.Dial(cr.host + ":" + cr.port)
	if err != nil {
		return nil, fmt.Errorf("smtp dial: %w", err)
	}

	if err := smtpClient.StartTLS(&tls.Config{ServerName: cr.host}); err != nil {
		return nil, fmt.Errorf("starttls: %w", err)
	}

	auth := smtp.PlainAuth("", cr.username, cr.password, cr.host)
	if err := smtpClient.Auth(auth); err != nil {
		return nil, fmt.Errorf("smtp auth: %w", err)
	}

	if err := smtpClient.Mail(cr.from); err != nil {
		return nil, fmt.Errorf("smtp mail: %w", err)
	}

	if err := smtpClient.Rcpt(cr.to); err != nil {
		return nil, fmt.Errorf("smtp rcpt: %w", err)
	}

	return smtpClient, nil
}
