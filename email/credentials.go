package email

import (
	"errors"
	"os"
)

type credentials struct {
	from     string
	to       string
	host     string
	port     string
	username string
	password string
}

func getCredentials() (*credentials, error) {
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

	username := os.Getenv("SMTP_USERNAME")
	if username == "" {
		return nil, errors.New("SMTP_USERNAME is missing")
	}

	password := os.Getenv("SMTP_PASSWORD")
	if password == "" {
		return nil, errors.New("SMTP_PASSWORD is missing")
	}

	return &credentials{
		from:     from,
		to:       to,
		host:     host,
		port:     port,
		username: username,
		password: password,
	}, nil
}
