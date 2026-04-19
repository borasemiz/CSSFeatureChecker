package output

import (
	"encoding/csv"
	"io"
	"strconv"

	"github.com/caniuse-scraper/scraper"
)

func WriteCSV(w io.Writer, results []scraper.Result) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	cw.Write([]string{"ID", "Coverage", "Status"})
	for _, r := range results {
		cw.Write([]string{r.ID, strconv.FormatFloat(r.Coverage, 'f', 2, 64), r.Status})
	}

	return cw.Error()
}
