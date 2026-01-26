// Package tradera provides a Go client for the Tradera SOAP API.
//
// This client wraps all 6 Tradera API services:
//   - SearchClient: Item search operations
//   - PublicClient: Public data (items, categories, users)
//   - ListingClient: Listing information
//   - RestrictedClient: Seller operations (requires user auth)
//   - OrderClient: Order management (requires user auth)
//   - BuyerClient: Buyer operations (requires user auth)
//
// Features:
//   - Full context.Context support for timeouts and cancellation
//   - Optional rate limiting
//   - Optional automatic retry with exponential backoff
//   - Optional response caching
//
// Basic usage:
//
//	client, err := tradera.NewClient(tradera.DefaultConfig(1234, "your-app-key"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	result, err := client.Search().Search(ctx, "vintage camera", 0)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Found %d items\n", result.TotalNumberOfItems)
package tradera

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/hooklift/gowsdl/soap"
	"github.com/pristabell/tradera-api-client/middleware"
)

// WSDL URLs for Tradera services
const (
	SearchServiceURL     = "https://api.tradera.com/v3/SearchService.asmx"
	PublicServiceURL     = "https://api.tradera.com/v3/PublicService.asmx"
	ListingServiceURL    = "https://api.tradera.com/v3/ListingService.asmx"
	RestrictedServiceURL = "https://api.tradera.com/v3/RestrictedService.asmx"
	OrderServiceURL      = "https://api.tradera.com/v3/OrderService.asmx"
	BuyerServiceURL      = "https://api.tradera.com/v3/BuyerService.asmx"
)

// Client is the main Tradera API client.
// It provides access to all Tradera services with optional middleware support.
type Client struct {
	config Config

	// Middleware
	rateLimiter *middleware.RateLimiter
	retryer     *middleware.Retryer
	cache       *middleware.Cache

	// HTTP client
	httpClient *http.Client

	// Lazy-initialized service clients
	searchClient     *SearchClient
	publicClient     *PublicClient
	listingClient    *ListingClient
	restrictedClient *RestrictedClient
	orderClient      *OrderClient
	buyerClient      *BuyerClient

	mu sync.Mutex
}

// NewClient creates a new Tradera API client with the given configuration.
func NewClient(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	c := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	// Initialize rate limiter if configured
	if config.RateLimit > 0 {
		c.rateLimiter = middleware.NewRateLimiter(config.RateLimit)
	}

	// Initialize retryer if configured
	if config.RetryEnabled {
		retryConfig := middleware.RetryConfig{
			MaxRetries:  config.MaxRetries,
			BaseDelay:   config.RetryBaseDelay,
			MaxDelay:    30 * time.Second,
			Multiplier:  2.0,
			Jitter:      0.2,
			ShouldRetry: IsRetryable,
		}
		c.retryer = middleware.NewRetryer(retryConfig)
	}

	// Initialize cache if configured
	if config.CacheTTL > 0 {
		c.cache = middleware.NewCache(config.CacheTTL)
	}

	return c, nil
}

// Search returns the SearchClient for item search operations.
func (c *Client) Search() *SearchClient {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.searchClient == nil {
		c.searchClient = newSearchClient(c)
	}
	return c.searchClient
}

// Public returns the PublicClient for public data operations.
func (c *Client) Public() *PublicClient {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.publicClient == nil {
		c.publicClient = newPublicClient(c)
	}
	return c.publicClient
}

// Listing returns the ListingClient for listing operations.
func (c *Client) Listing() *ListingClient {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.listingClient == nil {
		c.listingClient = newListingClient(c)
	}
	return c.listingClient
}

// Restricted returns the RestrictedClient for seller operations.
// Requires user authentication (UserID and Token in config).
func (c *Client) Restricted() *RestrictedClient {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.restrictedClient == nil {
		c.restrictedClient = newRestrictedClient(c)
	}
	return c.restrictedClient
}

// Order returns the OrderClient for order management operations.
// Requires user authentication (UserID and Token in config).
func (c *Client) Order() *OrderClient {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.orderClient == nil {
		c.orderClient = newOrderClient(c)
	}
	return c.orderClient
}

// Buyer returns the BuyerClient for buyer operations.
// Requires user authentication (UserID and Token in config).
func (c *Client) Buyer() *BuyerClient {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.buyerClient == nil {
		c.buyerClient = newBuyerClient(c)
	}
	return c.buyerClient
}

// Config returns the current configuration.
func (c *Client) Config() Config {
	return c.config
}

// Close releases any resources held by the client.
func (c *Client) Close() {
	if c.cache != nil {
		c.cache.Close()
	}
}

// createSOAPClient creates a new SOAP client for the given service URL.
func (c *Client) createSOAPClient(serviceURL string) *soap.Client {
	client := soap.NewClient(serviceURL, soap.WithHTTPClient(c.httpClient))

	// Add authentication headers
	client.AddHeader(AuthenticationHeader{
		AppID:  c.config.AppID,
		AppKey: c.config.AppKey,
	})

	// Add user authorization header if configured
	if c.config.HasUserAuth() {
		client.AddHeader(AuthorizationHeader{
			UserID: c.config.UserID,
			Token:  c.config.Token,
		})
	}

	return client
}

// executeWithMiddleware executes a function with rate limiting and retry support.
func (c *Client) executeWithMiddleware(ctx context.Context, fn func() error) error {
	// Apply rate limiting
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return err
		}
	}

	// Apply retry logic
	if c.retryer != nil {
		return c.retryer.Do(ctx, fn)
	}

	return fn()
}

// executeWithMiddlewareResult executes a function that returns a result with middleware support.
func executeWithMiddlewareResult[T any](c *Client, ctx context.Context, fn func() (T, error)) (T, error) {
	var result T

	// Apply rate limiting
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return result, err
		}
	}

	// Apply retry logic
	if c.retryer != nil {
		return middleware.DoWithResult(ctx, c.retryer, fn)
	}

	return fn()
}
