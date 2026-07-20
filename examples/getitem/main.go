// Command getitem demonstrates how to retrieve a Tradera item with REST API v4.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	tradera "github.com/SebbeJohansson/tradera-go-client"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <item-id>", os.Args[0])
	}

	itemID, err := strconv.ParseInt(os.Args[1], 10, 32)
	if err != nil {
		log.Fatal(err)
	}
	appID, err := strconv.Atoi(os.Getenv("TRADERA_APP_ID"))
	if err != nil || appID == 0 {
		log.Fatal("TRADERA_APP_ID must be a valid integer")
	}

	client, err := tradera.NewClient(tradera.DefaultConfig(appID, os.Getenv("TRADERA_APP_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	response, err := client.Public().GetItemWithResponse(context.Background(), int32(itemID))
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
