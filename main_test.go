package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/caniuse-scraper/scraper"
	"github.com/joho/godotenv"
)

func TestHandler(t *testing.T) {
	//t.Skip()
	godotenv.Load()
	err := handler(context.Background())

	if err != nil {
		t.Fatalf("error happened %s", err.Error())
	}
}

// --- Mock implementations ---

type mockStorage struct {
	retrieveData string
	retrieveErr  error
	savedData    []byte
	saveErr      error
}

func newMockStorage(data string, err error) *mockStorage {
	return &mockStorage{
		retrieveData: data,
		retrieveErr:  err,
	}
}

func (m *mockStorage) Retrieve() (io.ReadCloser, error) {
	if m.retrieveErr != nil {
		return nil, m.retrieveErr
	}
	return io.NopCloser(strings.NewReader(m.retrieveData)), nil
}

func (m *mockStorage) Save(r io.Reader) error {
	data, _ := io.ReadAll(r)
	m.savedData = data
	return m.saveErr
}

type mockIterator struct {
	features []*scraper.Feature
	idx      int
}

func (m *mockIterator) Next() (*scraper.Feature, error) {
	if m.idx >= len(m.features) {
		return nil, scraper.ErrorEndFeatures
	}
	f := m.features[m.idx]
	m.idx++
	return f, nil
}

type errIterator struct{ err error }

func (e *errIterator) Next() (*scraper.Feature, error) { return nil, e.err }

type mockEmailer struct {
	results []scraper.Result
	sendErr error
}

func (m *mockEmailer) WriteResult(r scraper.Result) error {
	m.results = append(m.results, r)
	return nil
}

func (m *mockEmailer) Send() error { return m.sendErr }

// --- Test helpers ---

func cssFeature(id string, coverage float64) *scraper.Feature {
	return &scraper.Feature{
		ID:         id,
		Title:      id + " feature",
		Categories: []string{"CSS"},
		UsagePercY: coverage,
		Status:     "ls",
	}
}

func nonCSSFeature(id string, coverage float64) *scraper.Feature {
	return &scraper.Feature{
		ID:         id,
		Title:      id + " feature",
		Categories: []string{"JS API"},
		UsagePercY: coverage,
		Status:     "ls",
	}
}

// makeOldCSV builds a minimal CSV string in the format expected by compare.ParseCSV.
func makeOldCSV(features map[string]float64) string {
	var sb strings.Builder
	sb.WriteString("ID,Coverage,Status\n")
	for id, cov := range features {
		fmt.Fprintf(&sb, "%s,%.2f,ls\n", id, cov)
	}
	return sb.String()
}

// --- Tests ---

// İlk çalıştırmada (eski CSV yok), coverage >= 90 olan CSS feature email edilmeli.
func TestDo_FirstRun_FeatureAbove90_GetsEmailed(t *testing.T) {
	st := newMockStorage("", &s3types.NoSuchKey{})
	iter := &mockIterator{features: []*scraper.Feature{cssFeature("css-grid", 95.0)}}
	mailer := &mockEmailer{}

	if err := do(st, iter, mailer); err != nil {
		t.Fatalf("do() error: %v", err)
	}

	if len(mailer.results) != 1 || mailer.results[0].ID != "css-grid" {
		t.Errorf("expected css-grid emailed, got %v", mailer.results)
	}
}

// İlk çalıştırmada, coverage < 90 olan CSS feature email edilmemeli.
func TestDo_FirstRun_FeatureBelow90_NotEmailed(t *testing.T) {
	st := newMockStorage("", &s3types.NoSuchKey{})
	iter := &mockIterator{features: []*scraper.Feature{cssFeature("css-transitions", 85.0)}}
	mailer := &mockEmailer{}

	if err := do(st, iter, mailer); err != nil {
		t.Fatalf("do() error: %v", err)
	}

	if len(mailer.results) != 0 {
		t.Errorf("expected no emails, got %v", mailer.results)
	}
}

// Eski verinde < 90 olan, yeni verinde >= 90'a çıkan feature email edilmeli.
func TestDo_CrossedThreshold_GetsEmailed(t *testing.T) {
	old := makeOldCSV(map[string]float64{"css-grid": 85.0})
	st := newMockStorage(old, nil)
	iter := &mockIterator{features: []*scraper.Feature{cssFeature("css-grid", 92.0)}}
	mailer := &mockEmailer{}

	if err := do(st, iter, mailer); err != nil {
		t.Fatalf("do() error: %v", err)
	}

	if len(mailer.results) != 1 || mailer.results[0].ID != "css-grid" {
		t.Errorf("expected css-grid emailed, got %v", mailer.results)
	}
}

// Eski verinde zaten >= 90 olan feature tekrar email edilmemeli.
func TestDo_AlreadyAbove90_NotEmailed(t *testing.T) {
	old := makeOldCSV(map[string]float64{"css-grid": 92.0})
	st := newMockStorage(old, nil)
	iter := &mockIterator{features: []*scraper.Feature{cssFeature("css-grid", 95.0)}}
	mailer := &mockEmailer{}

	if err := do(st, iter, mailer); err != nil {
		t.Fatalf("do() error: %v", err)
	}

	if len(mailer.results) != 0 {
		t.Errorf("expected no emails, got %v", mailer.results)
	}
}

// CSS olmayan feature ne email edilmeli ne de kaydedilmeli.
func TestDo_NonCSSFeature_Skipped(t *testing.T) {
	st := newMockStorage("", &s3types.NoSuchKey{})
	iter := &mockIterator{features: []*scraper.Feature{nonCSSFeature("fetch", 99.0)}}
	mailer := &mockEmailer{}

	if err := do(st, iter, mailer); err != nil {
		t.Fatalf("do() error: %v", err)
	}

	if len(mailer.results) != 0 {
		t.Errorf("expected no emails, got %v", mailer.results)
	}
}

// Karma feature listesinde: >= 90 CSS email edilmeli, diğerleri atlanmalı.
func TestDo_MixedFeatures_CorrectRouting(t *testing.T) {
	st := newMockStorage("", &s3types.NoSuchKey{})
	iter := &mockIterator{features: []*scraper.Feature{
		cssFeature("css-grid", 95.0),        // email edilmeli
		cssFeature("css-transitions", 80.0), // email edilmemeli (< 90)
		nonCSSFeature("fetch", 99.0),        // CSS değil, atlanmalı
	}}
	mailer := &mockEmailer{}

	if err := do(st, iter, mailer); err != nil {
		t.Fatalf("do() error: %v", err)
	}

	if len(mailer.results) != 1 || mailer.results[0].ID != "css-grid" {
		t.Errorf("expected only css-grid emailed, got %v", mailer.results)
	}
}

// NoSuchKey dışında bir storage hatası do()'nun hata döndürmesine yol açmalı.
func TestDo_StorageRetrieveError_ReturnsError(t *testing.T) {
	st := newMockStorage("", errors.New("connection refused"))
	iter := &mockIterator{}
	mailer := &mockEmailer{}

	err := do(st, iter, mailer)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Iterator beklenmedik hata döndürürse do() hata döndürmeli.
func TestDo_IteratorError_ReturnsError(t *testing.T) {
	st := newMockStorage("", &s3types.NoSuchKey{})
	iter := &errIterator{err: errors.New("iterator failed")}
	mailer := &mockEmailer{}

	err := do(st, iter, mailer)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// emailer.Send() hata döndürürse do() hata döndürmeli.
func TestDo_EmailSendError_ReturnsError(t *testing.T) {
	st := newMockStorage("", &s3types.NoSuchKey{})
	iter := &mockIterator{}
	mailer := &mockEmailer{sendErr: errors.New("smtp error")}

	err := do(st, iter, mailer)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// storage.Save() hata döndürürse do() hata döndürmeli.
func TestDo_StorageSaveError_ReturnsError(t *testing.T) {
	st := &mockStorage{
		retrieveErr: &s3types.NoSuchKey{},
		saveErr:     errors.New("s3 put failed"),
	}
	iter := &mockIterator{}
	mailer := &mockEmailer{}

	err := do(st, iter, mailer)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
