package scraper

import (
	"strings"
	"testing"
)

const sampleJSON = `{
	"eras": { "e-2": "2 versions back" },
	"agents": { "chrome": { "browser": "Chrome" } },
	"data": {
		"css-grid": {
			"title": "CSS Grid Layout",
			"spec": "https://drafts.csswg.org/css-grid/",
			"status": "cr",
			"categories": ["CSS"],
			"stats": { "chrome": { "57": "y" } },
			"usage_perc_y": 85.23,
			"usage_perc_a": 5.12
		},
		"flexbox": {
			"title": "CSS Flexible Box Layout Module",
			"spec": "https://drafts.csswg.org/css-flexbox/",
			"status": "cr",
			"categories": ["CSS"],
			"stats": { "chrome": { "21": "y" } },
			"usage_perc_y": 97.10,
			"usage_perc_a": 0.50
		}
	}
}`

func TestFeatureIteratorFromJSON(t *testing.T) {
	iterator := MakeFeatureIteratorFromJSON(strings.NewReader(sampleJSON))

	// First feature — check every field
	feature, err := iterator.Next()
	if err != nil {
		t.Fatalf("unexpected error on first Next(): %v", err)
	}
	if feature.ID != "css-grid" {
		t.Errorf("ID: got %q, want %q", feature.ID, "css-grid")
	}
	if feature.Title != "CSS Grid Layout" {
		t.Errorf("Title: got %q, want %q", feature.Title, "CSS Grid Layout")
	}
	if feature.Spec != "https://drafts.csswg.org/css-grid/" {
		t.Errorf("Spec: got %q, want %q", feature.Spec, "https://drafts.csswg.org/css-grid/")
	}
	if feature.Status != "cr" {
		t.Errorf("Status: got %q, want %q", feature.Status, "cr")
	}
	if len(feature.Categories) != 1 || feature.Categories[0] != "CSS" {
		t.Errorf("Categories: got %v, want [CSS]", feature.Categories)
	}
	if feature.Stats["chrome"]["57"] != "y" {
		t.Errorf("Stats: got %v, want chrome/57=y", feature.Stats)
	}
	if feature.UsagePercY != 85.23 {
		t.Errorf("UsagePercY: got %v, want 85.23", feature.UsagePercY)
	}
	if feature.UsagePercA != 5.12 {
		t.Errorf("UsagePercA: got %v, want 5.12", feature.UsagePercA)
	}

	// Second feature — confirm iteration advances correctly
	feature, err = iterator.Next()
	if err != nil {
		t.Fatalf("unexpected error on second Next(): %v", err)
	}
	if feature.ID != "flexbox" {
		t.Errorf("ID: got %q, want %q", feature.ID, "flexbox")
	}

	// Exhaustion — must return end_features
	_, err = iterator.Next()
	if err == nil || err.Error() != "end_features" {
		t.Fatalf("when all the features are consumed, `end_features` error should be thrown")
	}
}
