package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const DataURL = "https://raw.githubusercontent.com/Fyrd/caniuse/refs/heads/main/fulldata-json/data-2.0.json"

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

func FeatureURL(id string) string {
	return "https://caniuse.com/" + id
}

func FetchData(url string) (io.ReadCloser, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return resp.Body, nil
}

func IsCSSFeature(categories []string) bool {
	for _, cat := range categories {
		if strings.HasPrefix(strings.ToUpper(cat), "CSS") {
			return true
		}
	}
	return false
}

func FilterCSS(data CaniuseData, threshold float64) []Result {
	var results []Result
	for id, feature := range data.Data {
		if !IsCSSFeature(feature.Categories) {
			continue
		}
		coverage := feature.UsagePercY + feature.UsagePercA
		if coverage >= threshold {
			results = append(results, Result{
				ID:       id,
				Title:    feature.Title,
				Coverage: coverage,
				Spec:     feature.Spec,
				Status:   feature.Status,
				URL:      FeatureURL(id),
			})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Coverage > results[j].Coverage
	})
	return results
}

func Parse(r io.Reader) (CaniuseData, error) {
	var data CaniuseData
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return CaniuseData{}, fmt.Errorf("parse error: %w", err)
	}
	return data, nil
}
