# tradera-api-client

A comprehensive Go client library for the Tradera SOAP API.

## Project Background

This project is fully AI-generated based on [pristabell/tradera-api-client](https://github.com/SebbeJohansson/tradera-api-client). It provides a high-level, idiomatic Go interface for interacting with Tradera's web services.

## Installation

```bash
go get github.com/SebbeJohansson/tradera-go-client
```

## Available Clients

The library provides access to all 6 Tradera API services via dedicated clients:

- **SearchClient**: Item search operations
- **PublicClient**: Public data (items, categories, users)
- **ListingClient**: Listing information and status updates
- **RestrictedClient**: Seller operations (requires user authentication)
- **OrderClient**: Order and transaction management (requires user authentication)
- **BuyerClient**: Buyer operations and feedback (requires user authentication)

## Generated Code

The code in the [`generated/`](file:///c:/Users/sebbe/Projects/pristabell/go-tradera-api-client/generated) directory is automatically generated from the Tradera WSDL files using `gowsdl`. These files contain the underlying SOAP structures and service definitions.

## Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"github.com/SebbeJohansson/tradera-go-client"
)

func main() {
    // Create a new client with your AppID and AppKey
    cfg := tradera.DefaultConfig(1234, "your-app-key")
    client, err := tradera.NewClient(cfg)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Example: Search for items
    result, err := client.Search().Search(ctx, "vintage camera", 0)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d items\n", result.TotalNumberOfItems)
}
```

## License

MIT
