// Command seller demonstrates authenticated Tradera REST API v4 operations.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	tradera "github.com/SebbeJohansson/tradera-go-client/v4"
	"github.com/SebbeJohansson/tradera-go-client/v4/generated/rest/restricted"
)

func main() {
	appID, err := strconv.Atoi(os.Getenv("TRADERA_APP_ID"))
	if err != nil || appID == 0 {
		log.Fatal("TRADERA_APP_ID must be a valid integer")
	}
	userID, err := strconv.Atoi(os.Getenv("TRADERA_USER_ID"))
	if err != nil || userID == 0 {
		log.Fatal("TRADERA_USER_ID must be a valid integer")
	}

	config := tradera.DefaultConfig(appID, os.Getenv("TRADERA_APP_KEY")).WithUserAuth(userID, os.Getenv("TRADERA_TOKEN"))
	client, err := tradera.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.Restricted().GetSellerTransactionsWithResponse(
		context.Background(),
		&restricted.GetSellerTransactionsParams{},
	)
	if err != nil {
		log.Fatal(err)
	}
	if response.JSON200 == nil {
		log.Fatalf("unexpected response status: %d", response.StatusCode())
	}

	if err := json.NewEncoder(os.Stdout).Encode(response.JSON200); err != nil {
		log.Fatal(err)
	}
}
