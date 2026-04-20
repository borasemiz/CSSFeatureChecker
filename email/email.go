package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/caniuse-scraper/scraper"
)

type Sender interface {
	Send(features []scraper.Result) error
}

type client struct {
	from string
	to   string
	host string
	addr string
	auth smtp.Auth
}

func (c *client) Send(features []scraper.Result) error {
	conn, err := tls.Dial("tcp", c.addr, &tls.Config{ServerName: c.host})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}

	smtpClient, err := smtp.NewClient(conn, c.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer smtpClient.Close()

	if err := smtpClient.Auth(c.auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := smtpClient.Mail(c.from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}

	if err := smtpClient.Rcpt(c.to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err := w.Write(buildMessage(c.from, c.to, features)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	return w.Close()
}

func GetClient() (Sender, error) {
	from := os.Getenv("SMTP_EMAIL_FROM")
	if from == "" {
		return nil, errors.New("SMTP_EMAIL_FROM is missing")
	}

	to := os.Getenv("EMAIL_TO")
	if to == "" {
		return nil, errors.New("EMAIL_TO is missing")
	}

	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return nil, errors.New("SMTP_HOST is missing")
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		return nil, errors.New("SMTP_PORT is missing")
	}

	return &client{
		from: from,
		to:   to,
		host: host,
		addr: host + ":" + port,
		auth: smtp.PlainAuth("", os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"), host),
	}, nil
}

func buildMessage(from, to string, features []scraper.Result) []byte {
	var sb strings.Builder
	fmt.Fprintf(&sb, "From: %s\r\n", from)
	fmt.Fprintf(&sb, "To: %s\r\n", to)
	sb.WriteString("Subject: CSS Features Newly Above 90% Coverage\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString("The following CSS features have newly crossed 90% browser coverage:\n\n")
	for _, f := range features {
		fmt.Fprintf(&sb, "- %s: %.2f%%\n  %s\n\n", f.Title, f.Coverage, f.URL)
	}
	return []byte(sb.String())
}
