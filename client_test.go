package tradera_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	tradera "github.com/SebbeJohansson/tradera-go-client/v4"
	"github.com/SebbeJohansson/tradera-go-client/v4/generated/rest/search"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func TestPublicClientAddsAuthenticationAndMapsPath(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/v4/items/42" {
			t.Fatalf("unexpected path: %s", request.URL.Path)
		}
		assertHeader(t, request.Header, "X-App-Id", "123")
		assertHeader(t, request.Header, "X-App-Key", "app-key")
		assertHeader(t, request.Header, "X-User-Id", "456")
		assertHeader(t, request.Header, "X-User-Token", "user-token")
		assertHeader(t, request.Header, "X-Custom", "custom-value")

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"id":42,"shortDescription":"Test item"}`)),
			Request:    request,
		}, nil
	})

	config := tradera.DefaultConfig(123, "app-key").WithUserAuth(456, "user-token")
	config = config.WithBaseURL("https://example.test").WithHTTPClient(&http.Client{Transport: transport})
	config = config.WithHeaders(http.Header{"X-Custom": []string{"custom-value"}})
	client, err := tradera.NewPublicClient(config)
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.GetItemWithResponse(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if response.JSON200 == nil || response.JSON200.Id == nil || *response.JSON200.Id != 42 {
		t.Fatalf("unexpected response: %#v", response.JSON200)
	}
}

func TestGeneratedClientReturnsStructuredAPIError(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"X-Request-Id": []string{"request-123"},
			},
			Body:    io.NopCloser(strings.NewReader(`{"message":"invalid item"}`)),
			Request: request,
		}, nil
	})

	config := tradera.DefaultConfig(123, "app-key").WithBaseURL("https://example.test")
	config = config.WithHTTPClient(&http.Client{Transport: transport})
	client, err := tradera.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetItemWithResponse(context.Background(), 42)
	var apiError *tradera.APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiError.StatusCode != http.StatusBadRequest || apiError.RequestID != "request-123" {
		t.Fatalf("unexpected API error: %#v", apiError)
	}
	if string(apiError.Body) != `{"message":"invalid item"}` {
		t.Fatalf("unexpected API error body: %s", apiError.Body)
	}
	if client.Public() == nil {
		t.Fatal("expected aggregate client to expose the Public service")
	}
}

func TestGeneratedSearchClientSerializesQueryParameters(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		query := request.URL.Query()
		if query.Get("query") != "vintage camera" || query.Get("categoryId") != "12" || query.Get("pageNumber") != "3" {
			t.Fatalf("unexpected query parameters: %s", request.URL.RawQuery)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{}`)),
			Request:    request,
		}, nil
	})

	config := tradera.DefaultConfig(123, "app-key").WithBaseURL("https://example.test")
	config = config.WithHTTPClient(&http.Client{Transport: transport})
	client, err := tradera.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}

	query := "vintage camera"
	categoryID := int32(12)
	pageNumber := int32(3)
	response, err := client.Search().SearchWithResponse(context.Background(), &search.SearchParams{
		Query:      &query,
		CategoryId: &categoryID,
		PageNumber: &pageNumber,
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.JSON200 == nil {
		t.Fatal("expected a decoded JSON response")
	}
}

func TestConfigRejectsPartialUserAuthentication(t *testing.T) {
	config := tradera.DefaultConfig(123, "app-key")
	config.UserID = 456
	_, err := tradera.NewClient(config)
	if !errors.Is(err, tradera.ErrIncompleteUserAuth) {
		t.Fatalf("expected ErrIncompleteUserAuth, got %v", err)
	}
}

func assertHeader(t *testing.T, header http.Header, name, expected string) {
	t.Helper()
	if actual := header.Get(name); actual != expected {
		t.Fatalf("unexpected %s header: %q", name, actual)
	}
}
