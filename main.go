package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	sestypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/caniuse-scraper/compare"
	"github.com/caniuse-scraper/output"
	"github.com/caniuse-scraper/scraper"
	"github.com/caniuse-scraper/storage"
)

const (
	csvKey    = "output.csv"
	csvTmp    = "/tmp/output.csv"
	sesSender = "noreply@example.com"
	sesTo     = "team@example.com"
)

func handler(ctx context.Context) error {
	bucket := os.Getenv("BUCKET_NAME")
	if bucket == "" {
		return fmt.Errorf("BUCKET_NAME env var not set")
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	// 1. Fetch current caniuse data
	body, err := scraper.FetchData(scraper.DataURL)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer body.Close()

	data, err := scraper.Parse(body)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	freshResults := scraper.FilterCSS(data, 0)

	// 2. Fetch existing output.csv from S3
	var crossed []scraper.Result
	csvBody, err := storage.FromS3(ctx, s3Client, bucket, csvKey)
	if err != nil {
		var noSuchKey *s3types.NoSuchKey
		if !errors.As(err, &noSuchKey) {
			return fmt.Errorf("fetch csv from s3: %w", err)
		}
		// First run — no previous CSV to compare against
	} else {
		defer csvBody.Close()

		// 3. Compare
		old, err := compare.ParseCSV(csvBody)
		if err != nil {
			return fmt.Errorf("parse csv: %w", err)
		}

		crossed = compare.NewlyAbove90(old, freshResults)
	}

	// 4. Write new output.csv and upload to S3
	if err := output.WriteCSV(csvTmp, freshResults); err != nil {
		return fmt.Errorf("write csv: %w", err)
	}

	csv, err := os.ReadFile(csvTmp)
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}

	if err := storage.ToS3(ctx, s3Client, bucket, csvKey, csv); err != nil {
		return fmt.Errorf("upload csv: %w", err)
	}

	// 5. Email results via SES
	if len(crossed) > 0 {
		if err := sendEmail(ctx, sesv2.NewFromConfig(cfg), crossed); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
	}

	return nil
}

func sendEmail(ctx context.Context, client *sesv2.Client, features []scraper.Result) error {
	var sb strings.Builder
	sb.WriteString("The following CSS features have newly crossed 90% browser coverage:\n\n")
	for _, f := range features {
		fmt.Fprintf(&sb, "- %s: %.2f%%\n  %s\n\n", f.Title, f.Coverage, f.URL)
	}

	_, err := client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(sesSender),
		Destination: &sestypes.Destination{
			ToAddresses: []string{sesTo},
		},
		Content: &sestypes.EmailContent{
			Simple: &sestypes.Message{
				Subject: &sestypes.Content{
					Data: aws.String("CSS Features Newly Above 90% Coverage"),
				},
				Body: &sestypes.Body{
					Text: &sestypes.Content{
						Data: aws.String(sb.String()),
					},
				},
			},
		},
	})

	return err
}

func main() {
	lambda.Start(handler)
}
