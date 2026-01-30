// Command getitem demonstrates how to get detailed information about a Tradera item.
//
// Usage:
//
//	export TRADERA_APP_ID=12345
//	export TRADERA_APP_KEY=your-app-key
//	go run . 123456789
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tradera "github.com/SebbeJohansson/tradera-go-client"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <item-id>\n", os.Args[0])
		os.Exit(1)
	}

	itemID, err := strconv.ParseInt(os.Args[1], 10, 32)
	if err != nil {
		log.Fatalf("Invalid item ID: %v", err)
	}

	// Get credentials from environment
	appID, err := strconv.Atoi(os.Getenv("TRADERA_APP_ID"))
	if err != nil || appID == 0 {
		log.Fatal("TRADERA_APP_ID environment variable must be set to a valid integer")
	}

	appKey := os.Getenv("TRADERA_APP_KEY")
	if appKey == "" {
		log.Fatal("TRADERA_APP_KEY environment variable must be set")
	}

	// Create client
	client, err := tradera.NewClient(tradera.DefaultConfig(appID, appKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get item details
	fmt.Printf("Fetching item %d...\n\n", itemID)

	item, err := client.Public().GetItem(ctx, int32(itemID))
	if err != nil {
		log.Fatalf("Failed to get item: %v", err)
	}

	if item == nil {
		fmt.Println("Item not found")
		return
	}

	// Display item details
	fmt.Printf("Title:       %s\n", item.ShortDescription)
	fmt.Printf("Item ID:     %d\n", item.ID)
	fmt.Printf("Category:    %d\n", item.CategoryID)
	fmt.Println()

	fmt.Printf("Current bid: %d SEK\n", item.MaxBid)
	fmt.Printf("Total bids:  %d\n", item.TotalBids)
	if item.BuyItNowPrice != nil && *item.BuyItNowPrice > 0 {
		fmt.Printf("Buy It Now:  %d SEK\n", *item.BuyItNowPrice)
	}
	fmt.Println()

	fmt.Printf("Start date:  %s\n", item.StartDate.Format(time.RFC3339))
	fmt.Printf("End date:    %s\n", item.EndDate.Format(time.RFC3339))

	if item.Seller != nil {
		fmt.Println()
		fmt.Printf("Seller:      %s (ID: %d)\n", item.Seller.Alias, item.Seller.ID)
		fmt.Printf("Rating:      %d\n", item.Seller.TotalRating)
	}

	if item.LongDescription != "" {
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println(item.LongDescription)
	}
}
