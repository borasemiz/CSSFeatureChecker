package compare

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/caniuse-scraper/scraper"
)

// ParseCSV reads output.csv bytes into a map of feature ID → coverage.
func ParseCSV(data []byte) (map[string]float64, error) {
	r := csv.NewReader(bytes.NewReader(data))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	m := make(map[string]float64, len(rows)-1)
	for _, row := range rows[1:] { // skip header
		if len(row) < 2 {
			continue
		}
		val, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			continue
		}
		m[row[0]] = val
	}
	return m, nil
}

// NewlyAbove90 returns features whose coverage was below 90 in the old CSV
// but is now >= 90 in the fresh scraped results.
func NewlyAbove90(old map[string]float64, fresh []scraper.Result) []scraper.Result {
	var crossed []scraper.Result
	for _, r := range fresh {
		prev, exists := old[r.ID]
		if exists && prev < 90.0 && r.Coverage >= 90.0 {
			crossed = append(crossed, r)
		}
	}
	return crossed
}
