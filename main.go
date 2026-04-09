package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const dataURL = "https://raw.githubusercontent.com/Fyrd/caniuse/refs/heads/main/fulldata-json/data-2.0.json"

type Feature struct {
	Title      string                       `json:"title"`
	Spec       string                       `json:"spec"`
	Status     string                       `json:"status"`
	Categories []string                     `json:"categories"`
	Stats      map[string]map[string]string `json:"stats"`
	UsagePercY float64                      `json:"usage_perc_y"`
	UsagePercA float64                      `json:"usage_perc_a"`
}

type CaniuseData struct {
	Data map[string]Feature `json:"data"`
}

type Result struct {
	ID       string
	Title    string
	Coverage float64
	Spec     string
	Status   string
	URL      string
}

func featureURL(id string) string {
	return "https://caniuse.com/" + id
}

func fetchData(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func isCSSFeature(categories []string) bool {
	for _, cat := range categories {
		if strings.HasPrefix(strings.ToUpper(cat), "CSS") {
			return true
		}
	}
	return false
}

func main() {
	threshold := 90.0

	fmt.Printf("Fetching caniuse data from %s...\n", dataURL)
	body, err := fetchData(dataURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching data: %v\n", err)
		os.Exit(1)
	}

	var caniuse CaniuseData
	if err := json.Unmarshal(body, &caniuse); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	var results []Result
	for id, feature := range caniuse.Data {
		if !isCSSFeature(feature.Categories) {
			continue
		}
		// coverage = full support + partial support
		coverage := feature.UsagePercY + feature.UsagePercA
		if coverage >= threshold {
			results = append(results, Result{
				ID:       id,
				Title:    feature.Title,
				Coverage: coverage,
				Spec:     feature.Spec,
				Status:   feature.Status,
				URL:      featureURL(id),
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Coverage > results[j].Coverage
	})

	fmt.Printf("\nCSS features with >= %.0f%% browser coverage:\n", threshold)
	fmt.Printf("%-50s %-10s %-10s %s\n", "Feature", "Coverage", "Status", "URL")
	fmt.Println(strings.Repeat("-", 75))

	for _, r := range results {
		fmt.Printf("%-50s %-10.2f %-10s %s\n", r.Title, r.Coverage, r.Status, r.URL)
	}

	fmt.Printf("\nTotal: %d features\n", len(results))
}
