package scraper

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsCSSFeature(t *testing.T) {
	tests := []struct {
		name       string
		categories []string
		want       bool
	}{
		{"CSS category", []string{"CSS"}, true},
		{"CSS3 category", []string{"CSS3"}, true},
		{"CSS2 category", []string{"CSS2"}, true},
		{"lowercase css", []string{"css"}, true},
		{"mixed case", []string{"Css3"}, true},
		{"CSS among others", []string{"HTML5", "CSS", "JS"}, true},
		{"non-CSS only", []string{"HTML5", "JS"}, false},
		{"empty categories", []string{}, false},
		{"JS API only", []string{"JS API"}, false},
		{"SVG only", []string{"SVG"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCSSFeature(tt.categories)
			if got != tt.want {
				t.Errorf("IsCSSFeature(%v) = %v, want %v", tt.categories, got, tt.want)
			}
		})
	}
}

func TestFetchData_Success(t *testing.T) {
	body := `{"data":{}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	defer srv.Close()

	rc, err := FetchData(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("unexpected error reading body: %v", err)
	}
	if string(got) != body {
		t.Errorf("got %q, want %q", string(got), body)
	}
}

func TestFetchData_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := FetchData(srv.URL)
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

func TestParse(t *testing.T) {
	r := strings.NewReader(`{"data":{"css-grid":{"title":"CSS Grid Layout","status":"rec","categories":["CSS"],"usage_perc_y":92.0,"usage_perc_a":1.0}}}`)

	data, err := Parse(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f, ok := data.Data["css-grid"]
	if !ok {
		t.Fatal("expected css-grid in data")
	}
	if f.Title != "CSS Grid Layout" {
		t.Errorf("expected title %q, got %q", "CSS Grid Layout", f.Title)
	}
}

func TestFeatureURL(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{"simple id", "css-grid", "https://caniuse.com/css-grid"},
		{"id with numbers", "css3-colors", "https://caniuse.com/css3-colors"},
		{"id with multiple hyphens", "css-logical-props", "https://caniuse.com/css-logical-props"},
		{"empty id", "", "https://caniuse.com/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FeatureURL(tt.id)
			if got != tt.want {
				t.Errorf("FeatureURL(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestFilterCSS(t *testing.T) {
	data := CaniuseData{
		Data: map[string]Feature{
			"css-grid": {
				Title:      "CSS Grid Layout",
				Status:     "rec",
				Categories: []string{"CSS"},
				UsagePercY: 92.0,
				UsagePercA: 1.0,
			},
			"css-variables": {
				Title:      "CSS Variables",
				Status:     "rec",
				Categories: []string{"CSS"},
				UsagePercY: 60.0,
				UsagePercA: 5.0,
			},
			"fetch": {
				Title:      "Fetch API",
				Status:     "ls",
				Categories: []string{"JS API"},
				UsagePercY: 95.0,
				UsagePercA: 0.0,
			},
		},
	}

	results := FilterCSS(data, 90.0)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "css-grid" {
		t.Errorf("expected css-grid, got %s", results[0].ID)
	}
	if results[0].Coverage != 93.0 {
		t.Errorf("expected coverage 93.0, got %f", results[0].Coverage)
	}
	if results[0].URL != "https://caniuse.com/css-grid" {
		t.Errorf("unexpected URL: %s", results[0].URL)
	}
}

func TestFilterCSS_SortedDescending(t *testing.T) {
	data := CaniuseData{
		Data: map[string]Feature{
			"css-grid": {
				Title:      "CSS Grid Layout",
				Categories: []string{"CSS"},
				UsagePercY: 92.0,
			},
			"css-flexbox": {
				Title:      "CSS Flexbox",
				Categories: []string{"CSS"},
				UsagePercY: 98.0,
			},
			"css-transitions": {
				Title:      "CSS Transitions",
				Categories: []string{"CSS"},
				UsagePercY: 95.0,
			},
		},
	}

	results := FilterCSS(data, 90.0)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i := 1; i < len(results); i++ {
		if results[i].Coverage > results[i-1].Coverage {
			t.Errorf("results not sorted: %f > %f at index %d", results[i].Coverage, results[i-1].Coverage, i)
		}
	}
}
