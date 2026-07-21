package tradera

import (
	"fmt"
	"net/http"

	"github.com/SebbeJohansson/tradera-go-client/generated/rest/buyer"
)

// BuyerClient provides buyer operations.
type BuyerClient struct{ *buyer.ClientWithResponses }

// NewBuyerClient creates a client containing only Buyer service operations.
func NewBuyerClient(config Config) (*BuyerClient, error) {
	config, httpClient, err := prepareClient(config)
	if err != nil {
		return nil, err
	}
	return newBuyerClient(config.BaseURL, httpClient)
}

func newBuyerClient(baseURL string, httpClient *http.Client) (*BuyerClient, error) {
	client, err := buyer.NewClientWithResponses(baseURL, buyer.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create buyer client: %w", err)
	}
	return &BuyerClient{ClientWithResponses: client}, nil
}
