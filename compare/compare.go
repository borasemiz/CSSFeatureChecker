package compare

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/caniuse-scraper/scraper"
)

// ParseCSV streams r row by row into a map of feature ID → coverage.
// It never loads the full contents into memory.
func ParseCSV(r io.Reader) (map[string]float64, error) {
	reader := csv.NewReader(r)

	// skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	m := make(map[string]float64)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}
		val, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return nil, fmt.Errorf("parse coverage for %q: %w", row[0], err)
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
