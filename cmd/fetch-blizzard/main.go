package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
)

func main() {
	outputFile := flag.String("o", "data/blizzard_minions.json", "Output file")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fmt.Println("Launching browser...")
	pw, err := playwright.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to launch playwright: %v\n", err)
		os.Exit(1)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to launch browser: %v\n", err)
		os.Exit(1)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create page: %v\n", err)
		os.Exit(1)
	}
	defer page.Close()

	fmt.Println("Navigating to Blizzard card browser...")
	_, err = page.Goto("https://hs.blizzard.cn/battlegrounds/?sort=tier%3Aasc&bgCardType=minion", playwright.PageGotoOptions{
		Timeout: playwright.Float(60000),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to navigate: %v\n", err)
		os.Exit(1)
	}

	// Wait for content to load
	fmt.Println("Waiting for content...")
	page.WaitForSelector("body", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(60000)})
	page.WaitForTimeout(5000)

	// Extract minion data using JavaScript
	fmt.Println("Extracting minion data...")
	result, err := page.Evaluate(`() => {
		const minions = [];
		
		// Try multiple selectors that might contain card data
		const cardSelectors = [
			'.card-item',
			'.minion-card',
			'[class*="card"]',
			'.hearthstone-card'
		];
		
		let cards = [];
		for (const selector of cardSelectors) {
			cards = document.querySelectorAll(selector);
			if (cards.length > 0) break;
		}
		
		// If no cards found, try to get data from page context
		if (cards.length === 0) {
			// Look for any element with card-related data attributes
			const allElements = document.querySelectorAll('[class*="tier"], [class*="card"]');
			console.log('Found elements:', allElements.length);
		}
		
		// Extract text content that might be minion names
		const pageText = document.body.innerText;
		
		// Try to find card grid or list
		const gridSelectors = ['.card-grid', '.card-list', '.cards', '[class*="grid"]'];
		for (const selector of gridSelectors) {
			const grid = document.querySelector(selector);
			if (grid) {
				console.log('Found grid:', selector);
			}
		}
		
		return {
			totalCards: cards.length,
			pageText: pageText.substring(0, 2000),
			url: window.location.href
		};
	}`)
	
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to evaluate: %v\n", err)
		os.Exit(1)
	}
	
	data, _ := json.MarshalIndent(result, "", "  ")
	os.WriteFile(*outputFile, data, 0644)
	fmt.Printf("Saved to %s\n", *outputFile)
	fmt.Println(string(data))
}
