package tradera

import (
	"context"

	"github.com/SebbeJohansson/tradera-go-client/generated/order"
)

// OrderClient provides access to the Tradera Order API.
// Requires user authentication (UserID and Token in config).
type OrderClient struct {
	client  *Client
	service order.OrderServiceSoap
}

func newOrderClient(c *Client) *OrderClient {
	soapClient := c.createSOAPClient(OrderServiceURL)
	return &OrderClient{
		client:  c,
		service: order.NewOrderServiceSoap(soapClient),
	}
}

// SellerOrder represents an order for a seller.
type SellerOrder struct {
	ID             int32
	BuyerID        int32
	BuyerAlias     string
	TotalAmount    int32
	ShippingAmount int32
	Status         string
	CreatedDate    string
	IsPaid         bool
	IsShipped      bool
}

// GetSellerOrders retrieves orders for the authenticated seller.
func (c *OrderClient) GetSellerOrders(ctx context.Context) ([]*SellerOrder, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return nil, err
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*order.GetSellerOrdersResponse, error) {
		return c.service.GetSellerOrdersContext(ctx, &order.GetSellerOrders{})
	})
	if err != nil {
		return nil, err
	}

	if result.GetSellerOrdersResult == nil || result.GetSellerOrdersResult.SellerOrders == nil {
		return nil, nil
	}

	orders := make([]*SellerOrder, len(result.GetSellerOrdersResult.SellerOrders.SellerOrder))
	for i, o := range result.GetSellerOrdersResult.SellerOrders.SellerOrder {
		orders[i] = &SellerOrder{
			ID: o.OrderId,
		}
		if o.Buyer != nil {
			orders[i].BuyerID = o.Buyer.UserId
			orders[i].BuyerAlias = o.Buyer.Alias
		}
	}

	return orders, nil
}

// SetSellerOrderAsShipped marks an order as shipped.
func (c *OrderClient) SetSellerOrderAsShipped(ctx context.Context, orderID int32) error {
	if err := RequireUserAuth(c.client.config); err != nil {
		return err
	}

	return c.client.executeWithMiddleware(ctx, func() error {
		_, err := c.service.SetSellerOrderAsShippedContext(ctx, &order.SetSellerOrderAsShipped{
			Request: &order.SetSellerOrderAsShippedRequest{
				OrderId: orderID,
			},
		})
		return err
	})
}

// SetSellerOrderAsPaid marks an order as paid.
func (c *OrderClient) SetSellerOrderAsPaid(ctx context.Context, orderID int32) error {
	if err := RequireUserAuth(c.client.config); err != nil {
		return err
	}

	return c.client.executeWithMiddleware(ctx, func() error {
		_, err := c.service.SetSellerOrderAsPaidContext(ctx, &order.SetSellerOrderAsPaid{
			Request: &order.SetSellerOrderAsPaidRequest{
				OrderId: orderID,
			},
		})
		return err
	})
}
