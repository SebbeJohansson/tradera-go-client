package tradera

import (
	"fmt"
	"net/http"

	"github.com/SebbeJohansson/tradera-go-client/v4/generated/rest/restricted"
)

// RestrictedClient provides authenticated seller and listing operations.
type RestrictedClient struct {
	*restricted.ClientWithResponses
}

// NewRestrictedClient creates a client containing only Restricted service operations.
func NewRestrictedClient(config Config) (*RestrictedClient, error) {
	config, httpClient, err := prepareClient(config)
	if err != nil {
		return nil, err
	}
	return newRestrictedClient(config.BaseURL, httpClient)
}

func newRestrictedClient(baseURL string, httpClient *http.Client) (*RestrictedClient, error) {
	client, err := restricted.NewClientWithResponses(baseURL, restricted.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create restricted client: %w", err)
	}
	return &RestrictedClient{ClientWithResponses: client}, nil
}
