package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-lambda-go/lambda"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/caniuse-scraper/compare"
	"github.com/caniuse-scraper/email"
	"github.com/caniuse-scraper/output"
	"github.com/caniuse-scraper/scraper"
	"github.com/caniuse-scraper/storage"
)

func do(storage storage.Storage, featureIterator scraper.FeatureInterator, emailer email.Sender) error {
	savingDone := make(chan error)
	saveToStorageReader, saveToStorageWriter := io.Pipe()
	defer saveToStorageReader.Close()
	csvWriter := output.MakeCSVResultOutputWriter(saveToStorageWriter)
	csvWriter.WriteCSVHeader()

	go func() {
		savingDone <- storage.Save(saveToStorageReader)
	}()

	oldFeatures := make(map[string]float64)
	oldStream, err := storage.Retrieve()
	if err != nil {
		// First run — no previous CSV to compare against
		var noSuchKey *s3types.NoSuchKey
		if !errors.As(err, &noSuchKey) {
			return fmt.Errorf("fetch csv from s3: %w", err)
		}
	} else {
		defer oldStream.Close()

		// 3. Compare
		oldFeatures, err = compare.ParseCSV(oldStream)
		if err != nil {
			return fmt.Errorf("parse csv: %w", err)
		}
	}

	for {
		newFeature, err := featureIterator.Next()
		if err != nil && err == scraper.ErrorEndFeatures {
			break
		} else if err != nil {
			return fmt.Errorf("error processing new features: %v", err)
		}

		if !scraper.IsCSSFeature(newFeature.Categories) {
			continue
		}

		oldFeature, ok := oldFeatures[newFeature.ID]
		result := scraper.Result{
			ID:       newFeature.ID,
			Title:    newFeature.Title,
			Coverage: newFeature.UsagePercY,
			Spec:     newFeature.Spec,
			Status:   newFeature.Status,
			URL:      scraper.FeatureURL(newFeature.ID),
		}
		if (!ok && result.Coverage >= 90) || (result.Coverage >= 90 && oldFeature < 90) {
			emailer.WriteResult(result)
		} else if result.Coverage < 90 {
			csvWriter.WriteResult(result)
		}
	}

	if err := saveToStorageWriter.Close(); err != nil {
		return err
	}

	if err := emailer.Send(); err != nil {
		return err
	}

	if err := <-savingDone; err != nil {
		return err
	}

	return nil
}

func handler(ctx context.Context) error {
	storage, err := storage.GetStorage(ctx)
	if err != nil {
		return err
	}

	// 1. Fetch current caniuse data
	body, err := scraper.FetchData(scraper.DataURL)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer body.Close()

	featureIterator := scraper.MakeFeatureIteratorFromJSON(body)

	emailer, err := email.MakeClient()
	if err != nil {
		return err
	}

	if err := do(storage, featureIterator, emailer); err != nil {
		return err
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
