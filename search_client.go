package tradera

import (
	"fmt"
	"net/http"

	"github.com/SebbeJohansson/tradera-go-client/v4/generated/rest/search"
)

// SearchClient provides item search operations.
type SearchClient struct{ *search.ClientWithResponses }

// NewSearchClient creates a client containing only Search service operations.
func NewSearchClient(config Config) (*SearchClient, error) {
	config, httpClient, err := prepareClient(config)
	if err != nil {
		return nil, err
	}
	return newSearchClient(config.BaseURL, httpClient)
}

func newSearchClient(baseURL string, httpClient *http.Client) (*SearchClient, error) {
	client, err := search.NewClientWithResponses(baseURL, search.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create search client: %w", err)
	}
	return &SearchClient{ClientWithResponses: client}, nil
}
