package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/caniuse-scraper/scraper"
)

func WriteCSV(path string, results []scraper.Result) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"ID", "Coverage", "Status"})
	for _, r := range results {
		w.Write([]string{r.ID, strconv.FormatFloat(r.Coverage, 'f', 2, 64), r.Status})
	}

	return w.Error()
}
