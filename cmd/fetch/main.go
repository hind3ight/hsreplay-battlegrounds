package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hind3ight/hsreplay-battlegrounds/fetcher"
)

func main() {
	outputFile := flag.String("o", "", "Output file (default: stdout)")
	listOnly := flag.Bool("list", false, "Only fetch comp list, not details")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fmt.Println("Initializing Playwright...")
	f, err := fetcher.NewPlaywrightFetcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create fetcher: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if *listOnly {
		fmt.Println("Fetching comp list...")
		comps, err := f.FetchCompList(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch comp list: %v\n", err)
			os.Exit(1)
		}

		data, _ := json.MarshalIndent(comps, "", "  ")
		if *outputFile != "" {
			os.WriteFile(*outputFile, data, 0644)
			fmt.Printf("Saved %d comps to %s\n", len(comps), *outputFile)
		} else {
			fmt.Println(string(data))
		}
		return
	}

	fmt.Println("Fetching all comps with details...")
	result, err := f.FetchAllComps(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch comps: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	if *outputFile != "" {
		os.WriteFile(*outputFile, data, 0644)
		fmt.Printf("Saved %d comps to %s\n", len(result.Comps), *outputFile)
	} else {
		fmt.Println(string(data))
	}
}
