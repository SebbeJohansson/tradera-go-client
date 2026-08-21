// Package tradera provides a generated Go client for Tradera's REST API v4.
package tradera

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/SebbeJohansson/tradera-go-client/v4/generated/rest"
	"github.com/SebbeJohansson/tradera-go-client/v4/middleware"
)

// Client provides aggregate and service-scoped generated REST clients.
type Client struct {
	*rest.ClientWithResponses

	config Config

	search     *SearchClient
	public     *PublicClient
	listing    *ListingClient
	restricted *RestrictedClient
	order      *OrderClient
	buyer      *BuyerClient
}

// NewClient creates a Tradera REST API v4 client.
func NewClient(config Config) (*Client, error) {
	config, httpClient, err := prepareClient(config)
	if err != nil {
		return nil, err
	}

	rawClient, err := rest.NewClientWithResponses(config.BaseURL, rest.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create aggregate client: %w", err)
	}
	searchClient, err := newSearchClient(config.BaseURL, httpClient)
	if err != nil {
		return nil, err
	}
	publicClient, err := newPublicClient(config.BaseURL, httpClient)
	if err != nil {
		return nil, err
	}
	listingClient, err := newListingClient(config.BaseURL, httpClient)
	if err != nil {
		return nil, err
	}
	restrictedClient, err := newRestrictedClient(config.BaseURL, httpClient)
	if err != nil {
		return nil, err
	}
	orderClient, err := newOrderClient(config.BaseURL, httpClient)
	if err != nil {
		return nil, err
	}
	buyerClient, err := newBuyerClient(config.BaseURL, httpClient)
	if err != nil {
		return nil, err
	}

	return &Client{
		ClientWithResponses: rawClient,
		config:              config,
		search:              searchClient,
		public:              publicClient,
		listing:             listingClient,
		restricted:          restrictedClient,
		order:               orderClient,
		buyer:               buyerClient,
	}, nil
}

// Raw returns the aggregate generated client containing every REST v4 operation.
func (client *Client) Raw() *rest.ClientWithResponses { return client.ClientWithResponses }

// Search returns the Search service client.
func (client *Client) Search() *SearchClient { return client.search }

// Public returns the Public service client.
func (client *Client) Public() *PublicClient { return client.public }

// Listing returns the Listing service client.
func (client *Client) Listing() *ListingClient { return client.listing }

// Restricted returns the Restricted service client.
func (client *Client) Restricted() *RestrictedClient { return client.restricted }

// Order returns the Order service client.
func (client *Client) Order() *OrderClient { return client.order }

// Buyer returns the Buyer service client.
func (client *Client) Buyer() *BuyerClient { return client.buyer }

// Config returns the effective client configuration.
func (client *Client) Config() Config { return client.config }

func prepareClient(config Config) (Config, *http.Client, error) {
	if err := config.Validate(); err != nil {
		return Config{}, nil, err
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}
	return config, configuredHTTPClient(config), nil
}

func configuredHTTPClient(config Config) *http.Client {
	httpClient := &http.Client{}
	if config.HTTPClient != nil {
		*httpClient = *config.HTTPClient
	}
	if config.Timeout > 0 {
		httpClient.Timeout = config.Timeout
	}

	baseTransport := httpClient.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	transport := &apiTransport{
		base:    baseTransport,
		config:  config,
		headers: config.Headers.Clone(),
	}
	if config.RateLimit > 0 {
		transport.rateLimiter = middleware.NewRateLimiter(config.RateLimit)
	}
	if config.RetryEnabled {
		transport.retryer = middleware.NewRetryer(middleware.RetryConfig{
			MaxRetries:  config.MaxRetries,
			BaseDelay:   config.RetryBaseDelay,
			MaxDelay:    30 * time.Second,
			Multiplier:  2,
			Jitter:      0.2,
			ShouldRetry: IsRetryable,
		})
	}
	httpClient.Transport = transport
	return httpClient
}

type apiTransport struct {
	base        http.RoundTripper
	config      Config
	headers     http.Header
	rateLimiter *middleware.RateLimiter
	retryer     *middleware.Retryer
}

func (transport *apiTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if transport.rateLimiter != nil {
		if err := transport.rateLimiter.Wait(request.Context()); err != nil {
			return nil, err
		}
	}

	perform := func() (*http.Response, error) {
		attempt := request.Clone(request.Context())
		if request.GetBody != nil {
			body, err := request.GetBody()
			if err != nil {
				return nil, err
			}
			attempt.Body = body
		}
		transport.applyHeaders(attempt)

		response, err := transport.base.RoundTrip(attempt)
		if err != nil {
			return nil, err
		}
		if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
			return response, nil
		}

		body, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		return nil, &APIError{
			StatusCode: response.StatusCode,
			Body:       body,
			Header:     response.Header.Clone(),
			RequestID:  response.Header.Get("X-Request-Id"),
		}
	}

	if transport.retryer != nil {
		return middleware.DoWithResult(request.Context(), transport.retryer, perform)
	}
	return perform()
}

func (transport *apiTransport) applyHeaders(request *http.Request) {
	for name, values := range transport.headers {
		for _, value := range values {
			request.Header.Add(name, value)
		}
	}
	request.Header.Set("X-App-Id", strconv.Itoa(transport.config.AppID))
	request.Header.Set("X-App-Key", transport.config.AppKey)
	if transport.config.HasUserAuth() {
		request.Header.Set("X-User-Id", strconv.Itoa(transport.config.UserID))
		request.Header.Set("X-User-Token", transport.config.Token)
	}
}
