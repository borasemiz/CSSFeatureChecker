package main

import (
	"fmt"
	"os"

	"github.com/caniuse-scraper/output"
	"github.com/caniuse-scraper/scraper"
)

func main() {
	fmt.Println("Fetching caniuse data...")
	body, err := scraper.FetchData(scraper.DataURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching data: %v\n", err)
		os.Exit(1)
	}

	data, err := scraper.Parse(body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing data: %v\n", err)
		os.Exit(1)
	}

	results := scraper.FilterCSS(data, 0)

	if err := output.WriteCSV("output.csv", results); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wrote %d features to output.csv\n", len(results))
}
