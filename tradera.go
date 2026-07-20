// Package tradera provides a generated Go client for Tradera's REST API v4.
package tradera

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/SebbeJohansson/tradera-go-client/generated/rest"
	"github.com/SebbeJohansson/tradera-go-client/generated/rest/buyer"
	"github.com/SebbeJohansson/tradera-go-client/generated/rest/listing"
	"github.com/SebbeJohansson/tradera-go-client/generated/rest/order"
	"github.com/SebbeJohansson/tradera-go-client/generated/rest/public"
	"github.com/SebbeJohansson/tradera-go-client/generated/rest/restricted"
	"github.com/SebbeJohansson/tradera-go-client/generated/rest/search"
	"github.com/SebbeJohansson/tradera-go-client/middleware"
)

// Client provides aggregate and service-scoped generated REST clients.
type Client struct {
	config Config

	raw        *rest.ClientWithResponses
	search     *search.ClientWithResponses
	public     *public.ClientWithResponses
	listing    *listing.ClientWithResponses
	restricted *restricted.ClientWithResponses
	order      *order.ClientWithResponses
	buyer      *buyer.ClientWithResponses
}

// NewClient creates a Tradera REST API v4 client.
func NewClient(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}

	httpClient := configuredHTTPClient(config)

	rawClient, err := rest.NewClientWithResponses(config.BaseURL, rest.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create aggregate client: %w", err)
	}
	searchClient, err := search.NewClientWithResponses(config.BaseURL, search.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create search client: %w", err)
	}
	publicClient, err := public.NewClientWithResponses(config.BaseURL, public.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create public client: %w", err)
	}
	listingClient, err := listing.NewClientWithResponses(config.BaseURL, listing.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create listing client: %w", err)
	}
	restrictedClient, err := restricted.NewClientWithResponses(config.BaseURL, restricted.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create restricted client: %w", err)
	}
	orderClient, err := order.NewClientWithResponses(config.BaseURL, order.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create order client: %w", err)
	}
	buyerClient, err := buyer.NewClientWithResponses(config.BaseURL, buyer.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create buyer client: %w", err)
	}

	return &Client{
		config:     config,
		raw:        rawClient,
		search:     searchClient,
		public:     publicClient,
		listing:    listingClient,
		restricted: restrictedClient,
		order:      orderClient,
		buyer:      buyerClient,
	}, nil
}

// Raw returns the aggregate generated client containing every REST v4 operation.
func (client *Client) Raw() *rest.ClientWithResponses { return client.raw }

// Search returns the generated Search service client.
func (client *Client) Search() *search.ClientWithResponses { return client.search }

// Public returns the generated Public service client.
func (client *Client) Public() *public.ClientWithResponses { return client.public }

// Listing returns the generated Listing service client.
func (client *Client) Listing() *listing.ClientWithResponses { return client.listing }

// Restricted returns the generated Restricted service client.
func (client *Client) Restricted() *restricted.ClientWithResponses { return client.restricted }

// Order returns the generated Order service client.
func (client *Client) Order() *order.ClientWithResponses { return client.order }

// Buyer returns the generated Buyer service client.
func (client *Client) Buyer() *buyer.ClientWithResponses { return client.buyer }

// Config returns the effective client configuration.
func (client *Client) Config() Config { return client.config }

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
