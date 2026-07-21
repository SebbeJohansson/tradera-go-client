package tradera

import (
	"fmt"
	"net/http"

	"github.com/SebbeJohansson/tradera-go-client/generated/rest/listing"
)

// ListingClient provides listing restart operations.
type ListingClient struct{ *listing.ClientWithResponses }

// NewListingClient creates a client containing only Listing service operations.
func NewListingClient(config Config) (*ListingClient, error) {
	config, httpClient, err := prepareClient(config)
	if err != nil {
		return nil, err
	}
	return newListingClient(config.BaseURL, httpClient)
}

func newListingClient(baseURL string, httpClient *http.Client) (*ListingClient, error) {
	client, err := listing.NewClientWithResponses(baseURL, listing.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create listing client: %w", err)
	}
	return &ListingClient{ClientWithResponses: client}, nil
}
