// Command seller demonstrates how to use authenticated seller operations on Tradera.
//
// Usage:
//
//	export TRADERA_APP_ID=12345
//	export TRADERA_APP_KEY=your-app-key
//	export TRADERA_USER_ID=67890
//	export TRADERA_TOKEN=your-oauth-token
//	go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tradera "github.com/pristabell/tradera-api-client"
)

func main() {
	// Get credentials from environment
	appID, err := strconv.Atoi(os.Getenv("TRADERA_APP_ID"))
	if err != nil || appID == 0 {
		log.Fatal("TRADERA_APP_ID environment variable must be set to a valid integer")
	}

	appKey := os.Getenv("TRADERA_APP_KEY")
	if appKey == "" {
		log.Fatal("TRADERA_APP_KEY environment variable must be set")
	}

	userID, err := strconv.Atoi(os.Getenv("TRADERA_USER_ID"))
	if err != nil || userID == 0 {
		log.Fatal("TRADERA_USER_ID environment variable must be set to a valid integer")
	}

	token := os.Getenv("TRADERA_TOKEN")
	if token == "" {
		log.Fatal("TRADERA_TOKEN environment variable must be set")
	}

	// Create authenticated client
	config := tradera.DefaultConfig(appID, appKey)
	config.UserID = userID
	config.Token = token

	client, err := tradera.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get user info
	fmt.Println("Fetching user info...")
	userInfo, err := client.Restricted().GetUserInfo(ctx)
	if err != nil {
		log.Fatalf("Failed to get user info: %v", err)
	}

	fmt.Printf("Logged in as: %s (ID: %d)\n", userInfo.Alias, userInfo.ID)
	fmt.Printf("Email: %s\n", userInfo.Email)
	fmt.Println()

	// Get seller transactions
	fmt.Println("Fetching recent transactions...")
	transactions, err := client.Restricted().GetSellerTransactions(ctx)
	if err != nil {
		log.Fatalf("Failed to get transactions: %v", err)
	}

	if len(transactions) == 0 {
		fmt.Println("No transactions found")
	} else {
		fmt.Printf("Found %d transactions:\n\n", len(transactions))
		for _, tx := range transactions {
			status := ""
			if tx.IsMarkedAsPaidConfirmed {
				status += "[PAID] "
			}
			if tx.IsMarkedAsShipped {
				status += "[SHIPPED] "
			}

			fmt.Printf("  Transaction #%d: %d SEK %s\n", tx.ID, tx.Amount, status)
			fmt.Printf("    Date: %s\n", tx.Date)
			if tx.BuyerAlias != "" {
				fmt.Printf("    Buyer: %s\n", tx.BuyerAlias)
			}
			fmt.Println()
		}
	}

	// Get seller orders
	fmt.Println("Fetching seller orders...")
	orders, err := client.Order().GetSellerOrders(ctx)
	if err != nil {
		log.Fatalf("Failed to get orders: %v", err)
	}

	if len(orders) == 0 {
		fmt.Println("No orders found")
	} else {
		fmt.Printf("Found %d orders:\n\n", len(orders))
		for _, order := range orders {
			fmt.Printf("  Order #%d\n", order.ID)
			fmt.Printf("    Buyer: %s (ID: %d)\n", order.BuyerAlias, order.BuyerID)
			fmt.Println()
		}
	}
}
