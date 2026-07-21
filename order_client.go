package tradera

import (
	"fmt"
	"net/http"

	"github.com/SebbeJohansson/tradera-go-client/generated/rest/order"
)

// OrderClient provides seller order operations.
type OrderClient struct{ *order.ClientWithResponses }

// NewOrderClient creates a client containing only Order service operations.
func NewOrderClient(config Config) (*OrderClient, error) {
	config, httpClient, err := prepareClient(config)
	if err != nil {
		return nil, err
	}
	return newOrderClient(config.BaseURL, httpClient)
}

func newOrderClient(baseURL string, httpClient *http.Client) (*OrderClient, error) {
	client, err := order.NewClientWithResponses(baseURL, order.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create order client: %w", err)
	}
	return &OrderClient{ClientWithResponses: client}, nil
}
