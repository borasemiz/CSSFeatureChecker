package compare

import (
	"testing"

	"github.com/caniuse-scraper/scraper"
)

func TestParseCSV(t *testing.T) {
	csv := []byte("ID,Coverage,Status\ncss-grid,93.50,rec\ncss-flexbox,88.00,rec\n")

	m, err := ParseCSV(csv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m))
	}
	if m["css-grid"] != 93.50 {
		t.Errorf("expected 93.50 for css-grid, got %f", m["css-grid"])
	}
	if m["css-flexbox"] != 88.00 {
		t.Errorf("expected 88.00 for css-flexbox, got %f", m["css-flexbox"])
	}
}

func TestNewlyAbove90(t *testing.T) {
	old := map[string]float64{
		"css-grid":       88.0, // was below, now above → should appear
		"css-flexbox":    95.0, // was already above → should NOT appear
		"css-animations": 70.0, // was below, still below → should NOT appear
	}
	fresh := []scraper.Result{
		{ID: "css-grid", Title: "CSS Grid", Coverage: 92.0},
		{ID: "css-flexbox", Title: "CSS Flexbox", Coverage: 97.0},
		{ID: "css-animations", Title: "CSS Animations", Coverage: 75.0},
	}

	results := NewlyAbove90(old, fresh)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "css-grid" {
		t.Errorf("expected css-grid, got %s", results[0].ID)
	}
}

func TestNewlyAbove90_NoneQualify(t *testing.T) {
	old := map[string]float64{
		"css-grid": 95.0,
	}
	fresh := []scraper.Result{
		{ID: "css-grid", Coverage: 96.0},
	}

	results := NewlyAbove90(old, fresh)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestNewlyAbove90_FeatureNotInOld(t *testing.T) {
	old := map[string]float64{}
	fresh := []scraper.Result{
		{ID: "css-grid", Coverage: 92.0},
	}

	results := NewlyAbove90(old, fresh)
	if len(results) != 0 {
		t.Errorf("expected 0 results for unknown feature, got %d", len(results))
	}
}
