package main

import (
	"net/http"
	"net/http/httptest"
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
			got := isCSSFeature(tt.categories)
			if got != tt.want {
				t.Errorf("isCSSFeature(%v) = %v, want %v", tt.categories, got, tt.want)
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

	got, err := fetchData(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	_, err := fetchData(srv.URL)
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
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
			got := featureURL(tt.id)
			if got != tt.want {
				t.Errorf("featureURL(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
