// Command search demonstrates how to search for items with Tradera REST API v4.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	tradera "github.com/SebbeJohansson/tradera-go-client/v4"
	"github.com/SebbeJohansson/tradera-go-client/v4/generated/rest/search"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <search-query>", os.Args[0])
	}

	appID, err := strconv.Atoi(os.Getenv("TRADERA_APP_ID"))
	if err != nil || appID == 0 {
		log.Fatal("TRADERA_APP_ID must be a valid integer")
	}

	query := os.Args[1]
	client, err := tradera.NewClient(tradera.DefaultConfig(appID, os.Getenv("TRADERA_APP_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.Search().SearchWithResponse(context.Background(), &search.SearchParams{Query: &query})
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
