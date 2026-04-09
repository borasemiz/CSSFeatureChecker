package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/caniuse-scraper/scraper"
)

func main() {
	threshold := 90.0

	fmt.Printf("Fetching caniuse data")
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

	results := scraper.FilterCSS(data, threshold)

	fmt.Printf("\nCSS features with >= %.0f%% browser coverage:\n", threshold)
	fmt.Printf("%-50s %-10s %-10s %s\n", "Feature", "Coverage", "Status", "URL")
	fmt.Println(strings.Repeat("-", 100))

	for _, r := range results {
		fmt.Printf("%-50s %-10.2f %-10s %s\n", r.Title, r.Coverage, r.Status, r.URL)
	}

	fmt.Printf("\nTotal: %d features\n", len(results))
}
