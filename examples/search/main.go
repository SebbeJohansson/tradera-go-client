// Command search demonstrates how to search for items on Tradera.
//
// Usage:
//
//	export TRADERA_APP_ID=12345
//	export TRADERA_APP_KEY=your-app-key
//	go run . "vintage camera"
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tradera "github.com/pristabell/tradera-api-client"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <search-query>\n", os.Args[0])
		os.Exit(1)
	}

	query := os.Args[1]

	// Get credentials from environment
	appID, err := strconv.Atoi(os.Getenv("TRADERA_APP_ID"))
	if err != nil || appID == 0 {
		log.Fatal("TRADERA_APP_ID environment variable must be set to a valid integer")
	}

	appKey := os.Getenv("TRADERA_APP_KEY")
	if appKey == "" {
		log.Fatal("TRADERA_APP_KEY environment variable must be set")
	}

	// Create client with rate limiting and retries
	config := tradera.DefaultConfig(appID, appKey)
	config.RateLimit = 5
	config.RetryEnabled = true
	config.MaxRetries = 3

	client, err := tradera.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Perform search
	fmt.Printf("Searching for: %s\n\n", query)

	result, err := client.Search().Search(ctx, query, 0)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	fmt.Printf("Found %d items (showing first page)\n", result.TotalNumberOfItems)
	fmt.Println(strings.Repeat("-", 60))

	for _, item := range result.Items {
		endTime := item.EndDate.ToGoTime().Format("2006-01-02 15:04")
		fmt.Printf("ID: %-12d Price: %-8d SEK  Ends: %s\n", item.ID, item.NextBid, endTime)
		fmt.Printf("    %s\n\n", truncate(item.ShortDescription, 55))
	}

	if int(result.TotalNumberOfItems) > len(result.Items) {
		fmt.Printf("\n... and %d more items\n", int(result.TotalNumberOfItems)-len(result.Items))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
