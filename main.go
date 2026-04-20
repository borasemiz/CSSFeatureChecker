package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/caniuse-scraper/compare"
	"github.com/caniuse-scraper/email"
	"github.com/caniuse-scraper/output"
	"github.com/caniuse-scraper/scraper"
	"github.com/caniuse-scraper/storage"
)

const csvKey = "output.csv"

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

	// 4. Write new output.csv directly into a buffer and upload to S3
	var buf bytes.Buffer
	if err := output.WriteCSV(&buf, freshResults); err != nil {
		return fmt.Errorf("write csv: %w", err)
	}

	if err := storage.ToS3(ctx, s3Client, bucket, csvKey, &buf); err != nil {
		return fmt.Errorf("upload csv: %w", err)
	}

	// 5. Email results via SMTP
	if len(crossed) > 0 {
		emailClient, err := email.GetClient()
		if err != nil {
			return fmt.Errorf("email client: %w", err)
		}

		if err := emailClient.Send(crossed); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
