package output

import (
	"encoding/csv"
	"io"
	"strconv"

	"github.com/caniuse-scraper/scraper"
)

type CSVResultOutputWriter interface {
	WriteCSVHeader() error
	WriteResult(result scraper.Result) error
}

type csvResultOutputWriter struct {
	writer *csv.Writer
}

func (w *csvResultOutputWriter) WriteCSVHeader() error {
	return w.writer.Write([]string{"ID", "Coverage", "Status"})
}

func (w *csvResultOutputWriter) WriteResult(result scraper.Result) error {
	return w.writer.Write([]string{
		result.ID,
		strconv.FormatFloat(result.Coverage, 'f', 2, 64),
		result.Status,
	})
}

func WriteCSV(w io.Writer, results []scraper.Result) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	cw.Write([]string{"ID", "Coverage", "Status"})
	for _, r := range results {
		cw.Write([]string{r.ID, strconv.FormatFloat(r.Coverage, 'f', 2, 64), r.Status})
	}

	return cw.Error()
}

func MakeCSVResultOutputWriter(writer io.Writer) CSVResultOutputWriter {
	return &csvResultOutputWriter{
		writer: csv.NewWriter(writer),
	}
}
