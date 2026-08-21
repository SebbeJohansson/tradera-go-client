# Tradera API Client for Go

A generated Go client for Tradera's REST API v4. The package provides aggregate and service-scoped clients with generated request types, response types, and endpoint methods.

## Features

- **Generated API surface** - endpoint methods and models come from Tradera's OpenAPI contract
- **Service clients** - Search, Public, Listing, Restricted, Order, and Buyer clients
- **Typed responses** - generated `...WithResponse` methods decode documented JSON responses
- **Authentication** - application and optional user credentials are added to every request
- **Configurable transport** - custom base URL, HTTP client, headers, timeout, retries, and rate limiting
- **Structured errors** - non-successful responses return `*tradera.APIError`

## Official Tradera API Documentation

- [REST API Getting Started](https://api.tradera.com/v4/docs/index.html)
- [REST API Reference](https://api.tradera.com/v4/swagger/index.html)
- [Developer Center](https://api.tradera.com/)

## Looking for SOAP API v3?

Tradera's SOAP API remains at v3 and is intended for existing integrations. The previous Go SOAP client source is preserved on the [`archive/v3`](https://github.com/SebbeJohansson/tradera-go-client/tree/archive/v3) branch. New integrations should use this REST v4 client.

## Project Background

This project was originally fully AI-generated based on [pristabell/tradera-api-client](https://github.com/SebbeJohansson/tradera-api-client). It has now deviated from the original TypeScript client, but is still meant to be the variant for GO.

## Installation

```bash
go get github.com/SebbeJohansson/tradera-go-client/v4
```

## Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	tradera "github.com/SebbeJohansson/tradera-go-client/v4"
	"github.com/SebbeJohansson/tradera-go-client/v4/generated/rest/search"
)

func main() {
	client, err := tradera.NewClient(tradera.DefaultConfig(1234, "your-app-key"))
	if err != nil {
		log.Fatal(err)
	}

	query := "vintage camera"
	response, err := client.Search().SearchWithResponse(
		context.Background(),
		&search.SearchParams{Query: &query},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Result: %#v\n", response.JSON200)
}
```

`Client` is the aggregate client. Its generated methods are promoted directly, and the same operations are also grouped into service clients:

### Public API

```go
response, err := client.GetItemWithResponse(ctx, 123456789)
if err != nil {
	log.Fatal(err)
}

item := response.JSON200
```

### Service-Scoped Clients

Service clients can be obtained from the aggregate client or constructed independently, matching the service classes in the TypeScript client:

```go
publicClient, err := tradera.NewPublicClient(
	tradera.DefaultConfig(1234, "your-app-key"),
)
if err != nil {
	log.Fatal(err)
}

response, err := publicClient.GetItemWithResponse(ctx, 123456789)
```

### User Authentication

Restricted, Order, and Buyer operations generally require user credentials. The same config can create either an aggregate or a service-scoped client:

```go
config := tradera.DefaultConfig(1234, "your-app-key").
	WithUserAuth(5678, "your-user-token")

client, err := tradera.NewRestrictedClient(config)
```

The client sends these REST headers:

- `X-App-Id`
- `X-App-Key`
- `X-User-Id`, when user authentication is configured
- `X-User-Token`, when user authentication is configured

## Available Clients

| Aggregate method | Standalone constructor | Generated package | Purpose |
| --- | --- | --- | --- |
| `client.Raw()` | `NewClient` | `generated/rest` | All REST v4 operations |
| `client.Search()` | `NewSearchClient` | `generated/rest/search` | Item search operations |
| `client.Public()` | `NewPublicClient` | `generated/rest/public` | Public items, users, categories, and reference data |
| `client.Listing()` | `NewListingClient` | `generated/rest/listing` | Listing restart information |
| `client.Restricted()` | `NewRestrictedClient` | `generated/rest/restricted` | Authenticated seller and listing operations |
| `client.Order()` | `NewOrderClient` | `generated/rest/order` | Seller order operations |
| `client.Buyer()` | `NewBuyerClient` | `generated/rest/buyer` | Buyer operations |

Methods ending in `WithResponse` return generated wrappers containing the raw HTTP response, response body, status helpers, and decoded fields such as `JSON200`.

## Errors

Non-2xx responses are returned as `*tradera.APIError` with the status code, response body, headers, and request ID when provided by Tradera.

```go
var apiError *tradera.APIError
if errors.As(err, &apiError) {
	fmt.Printf("status=%d request=%s body=%s\n", apiError.StatusCode, apiError.RequestID, apiError.Body)
}
```

## Generated Code

The OpenAPI contract is pinned under `openapi/`. Because the upstream document does not provide `operationId` values, `openapi/client-operations.json` maintains reviewed method names and service assignments.

```bash
go run ./internal/updatecontract
go generate ./...
git diff --exit-code -- generated/rest
```

Everything under `generated/rest/` is generated and must not be edited directly. Some search request/result schemas are open maps because those schemas are empty in Tradera's current OpenAPI document.

## Development

```bash
go generate ./...
go test ./...
go vet ./...
```

## License

MIT
