package tradera

import (
	"fmt"
	"net/http"

	"github.com/SebbeJohansson/tradera-go-client/generated/rest/public"
)

// PublicClient provides public item, user, category, and reference data operations.
type PublicClient struct{ *public.ClientWithResponses }

// NewPublicClient creates a client containing only Public service operations.
func NewPublicClient(config Config) (*PublicClient, error) {
	config, httpClient, err := prepareClient(config)
	if err != nil {
		return nil, err
	}
	return newPublicClient(config.BaseURL, httpClient)
}

func newPublicClient(baseURL string, httpClient *http.Client) (*PublicClient, error) {
	client, err := public.NewClientWithResponses(baseURL, public.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create public client: %w", err)
	}
	return &PublicClient{ClientWithResponses: client}, nil
}
