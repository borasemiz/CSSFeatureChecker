package output

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/caniuse-scraper/scraper"
)

func TestWriteCSV(t *testing.T) {
	var buf bytes.Buffer
	results := []scraper.Result{
		{ID: "css-grid", Coverage: 93.50, Status: "rec"},
		{ID: "css-flexbox", Coverage: 98.20, Status: "rec"},
	}

	if err := WriteCSV(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(rows) != 3 { // header + 2 results
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[0][0] != "ID" || rows[0][1] != "Coverage" || rows[0][2] != "Status" {
		t.Errorf("unexpected header: %v", rows[0])
	}
	if rows[1][0] != "css-grid" || rows[1][1] != "93.50" || rows[1][2] != "rec" {
		t.Errorf("unexpected row 1: %v", rows[1])
	}
}

func TestWriteCSV_EmptyResults(t *testing.T) {
	var buf bytes.Buffer

	if err := WriteCSV(&buf, []scraper.Result{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows, _ := csv.NewReader(&buf).ReadAll()
	if len(rows) != 1 { // header only
		t.Fatalf("expected 1 row (header only), got %d", len(rows))
	}
}
