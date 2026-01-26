package tradera_test

import (
	"context"
	"fmt"
	"log"
	"time"

	tradera "github.com/pristabell/tradera-api-client"
)

// This example shows how to create a basic client and search for items.
func Example_basicSearch() {
	// Create a client with your App ID and App Key
	client, err := tradera.NewClient(tradera.DefaultConfig(12345, "your-app-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Search for items
	result, err := client.Search().Search(ctx, "vintage camera", 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d items\n", result.TotalNumberOfItems)
	for _, item := range result.Items {
		fmt.Printf("- %s (ID: %d)\n", item.ShortDescription, item.ID)
	}
}

// This example shows how to get detailed information about a specific item.
func Example_getItem() {
	client, err := tradera.NewClient(tradera.DefaultConfig(12345, "your-app-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get item details by ID
	item, err := client.Public().GetItem(ctx, 123456789)
	if err != nil {
		log.Fatal(err)
	}

	if item != nil {
		fmt.Printf("Item: %s\n", item.ShortDescription)
		fmt.Printf("Price: %d SEK\n", item.MaxBid)
		fmt.Printf("Ends: %s\n", item.EndDate.Format(time.RFC3339))
	}
}

// This example shows how to browse categories.
func Example_getCategories() {
	client, err := tradera.NewClient(tradera.DefaultConfig(12345, "your-app-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get all categories
	categories, err := client.Public().GetCategories(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Categories:")
	for _, cat := range categories {
		fmt.Printf("- %s (ID: %d)\n", cat.Name, cat.ID)
	}
}

// This example shows how to configure the client with rate limiting and retries.
func Example_withMiddleware() {
	config := tradera.Config{
		AppID:  12345,
		AppKey: "your-app-key",

		// Rate limit to 5 requests per second
		RateLimit: 5,

		// Enable automatic retries
		RetryEnabled:   true,
		MaxRetries:     3,
		RetryBaseDelay: 500 * time.Millisecond,

		// Cache responses for 5 minutes
		CacheTTL: 5 * time.Minute,

		// Request timeout
		Timeout: 30 * time.Second,
	}

	client, err := tradera.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Use client as normal - middleware is applied automatically
	ctx := context.Background()
	result, err := client.Search().Search(ctx, "test", 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d items\n", result.TotalNumberOfItems)
}

// This example shows how to use authenticated operations for sellers.
func Example_sellerOperations() {
	// Create a client with user authentication
	config := tradera.DefaultConfig(12345, "your-app-key")
	config.UserID = 67890
	config.Token = "user-oauth-token"

	client, err := tradera.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get seller's transactions
	transactions, err := client.Restricted().GetSellerTransactions(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d transactions\n", len(transactions))
	for _, tx := range transactions {
		fmt.Printf("- Transaction %d: %d SEK\n", tx.ID, tx.Amount)
	}
}

// This example shows how to use the OAuth token flow.
func Example_oauthFlow() {
	client, err := tradera.NewClient(tradera.DefaultConfig(12345, "your-app-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Step 1: Get a login URL for the user to authorize your app
	// The user should be directed to: https://api.tradera.com/v3/Authenticate.aspx?appId=YOUR_APP_ID
	// After authorization, they receive a secret key

	// Step 2: Exchange the user ID and secret key for an access token
	userID := int32(12345)          // User ID from authorization
	secretKey := "secret-from-auth" // Secret key from authorization callback

	token, err := client.Public().FetchToken(ctx, userID, secretKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Access token: %s\n", token)

	// Step 3: Create a new client with the user's credentials
	userConfig := tradera.DefaultConfig(12345, "your-app-key")
	userConfig.UserID = int(userID)
	userConfig.Token = token

	userClient, err := tradera.NewClient(userConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer userClient.Close()

	// Now you can use authenticated endpoints
	userInfo, err := userClient.Restricted().GetUserInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Logged in as: %s\n", userInfo.Alias)
}

// This example shows advanced search with filters.
func Example_advancedSearch() {
	client, err := tradera.NewClient(tradera.DefaultConfig(12345, "your-app-key"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()

	// Search with advanced filters
	minPrice := int32(1000)
	maxPrice := int32(5000)
	params := tradera.SearchAdvancedRequest{
		SearchWords:  "iPhone",
		CategoryID:   345262, // Electronics > Phones
		PriceMinimum: &minPrice,
		PriceMaximum: &maxPrice,
		OrderBy:      "EndDateAscending",
	}

	result, err := client.Search().SearchAdvanced(ctx, params)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d iPhones between 1000-5000 SEK\n", result.TotalNumberOfItems)
}
