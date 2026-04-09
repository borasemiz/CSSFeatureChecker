package output

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/caniuse-scraper/scraper"
)

func TestWriteCSV_CreatesFile(t *testing.T) {
	path := t.TempDir() + "/test.csv"
	results := []scraper.Result{
		{ID: "css-grid", Coverage: 93.50, Status: "rec"},
		{ID: "css-flexbox", Coverage: 98.20, Status: "rec"},
	}

	if err := WriteCSV(path, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(rows) != 3 { // header + 2 results
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	// header
	if rows[0][0] != "ID" || rows[0][1] != "Coverage" || rows[0][2] != "Status" {
		t.Errorf("unexpected header: %v", rows[0])
	}

	// first data row
	if rows[1][0] != "css-grid" || rows[1][1] != "93.50" || rows[1][2] != "rec" {
		t.Errorf("unexpected row 1: %v", rows[1])
	}
}

func TestWriteCSV_EmptyResults(t *testing.T) {
	path := t.TempDir() + "/empty.csv"

	if err := WriteCSV(path, []scraper.Result{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	defer f.Close()

	rows, _ := csv.NewReader(f).ReadAll()
	if len(rows) != 1 { // header only
		t.Fatalf("expected 1 row (header only), got %d", len(rows))
	}
}

func TestWriteCSV_InvalidPath(t *testing.T) {
	err := WriteCSV("/nonexistent/dir/test.csv", []scraper.Result{})
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}
