package tradera

import (
	"context"

	"github.com/pristabell/tradera-api-client/generated/listing"
)

// ListingClient provides access to the Tradera Listing API.
type ListingClient struct {
	client  *Client
	service listing.ListingServiceSoap
}

func newListingClient(c *Client) *ListingClient {
	soapClient := c.createSOAPClient(ListingServiceURL)
	return &ListingClient{
		client:  c,
		service: listing.NewListingServiceSoap(soapClient),
	}
}

// ItemRestarts represents information about item restarts.
type ItemRestarts struct {
	LastRestartedItemID int32
	AncestorItemID      int32
	RestartedItems      []*RestartedItem
}

// RestartedItem represents a single restart event.
type RestartedItem struct {
	RestartedItemID   int32
	RestartedAsItemID int32
	RestartedDate     string
}

// GetItemRestarts retrieves item restart information.
func (c *ListingClient) GetItemRestarts(ctx context.Context, itemID int32) (*ItemRestarts, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*listing.GetItemRestartsResponse, error) {
		return c.service.GetItemRestartsContext(ctx, &listing.GetItemRestarts{
			ItemId: itemID,
		})
	})
	if err != nil {
		return nil, err
	}

	if result.GetItemRestartsResult == nil {
		return nil, nil
	}

	r := result.GetItemRestartsResult
	restarts := &ItemRestarts{
		LastRestartedItemID: r.LastRestartedItemId,
		AncestorItemID:      r.AncestorItemId,
	}

	if r.RestartedItems != nil && r.RestartedItems.RestartedItem != nil {
		restarts.RestartedItems = make([]*RestartedItem, len(r.RestartedItems.RestartedItem))
		for i, item := range r.RestartedItems.RestartedItem {
			restarts.RestartedItems[i] = &RestartedItem{
				RestartedItemID:   item.RestartedItemId,
				RestartedAsItemID: item.RestartedAsItemId,
				RestartedDate:     item.RestartedDate.ToGoTime().String(),
			}
		}
	}

	return restarts, nil
}
