package tradera

import (
	"context"

	"github.com/pristabell/tradera-api-client/generated/restricted"
)

// RestrictedClient provides access to the Tradera Restricted API.
// Requires user authentication (UserID and Token in config).
type RestrictedClient struct {
	client  *Client
	service restricted.RestrictedServiceSoap
}

func newRestrictedClient(c *Client) *RestrictedClient {
	soapClient := c.createSOAPClient(RestrictedServiceURL)
	return &RestrictedClient{
		client:  c,
		service: restricted.NewRestrictedServiceSoap(soapClient),
	}
}

// SellerTransaction represents a seller transaction.
type SellerTransaction struct {
	ID                      int32
	Date                    string
	Amount                  int32
	LastUpdatedDate         string
	IsMarkedAsPaidConfirmed bool
	IsMarkedAsShipped       bool
	IsShippingBooked        bool
	BuyerID                 int32
	BuyerAlias              string
	ItemID                  int32
	ItemTitle               string
}

// GetSellerTransactions retrieves transactions for the authenticated seller.
func (c *RestrictedClient) GetSellerTransactions(ctx context.Context) ([]*SellerTransaction, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return nil, err
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*restricted.GetSellerTransactionsResponse, error) {
		return c.service.GetSellerTransactionsContext(ctx, &restricted.GetSellerTransactions{})
	})
	if err != nil {
		return nil, err
	}

	if result.GetSellerTransactionsResult == nil || result.GetSellerTransactionsResult.Transaction == nil {
		return nil, nil
	}

	transactions := make([]*SellerTransaction, len(result.GetSellerTransactionsResult.Transaction))
	for i, t := range result.GetSellerTransactionsResult.Transaction {
		transactions[i] = &SellerTransaction{
			ID:                      t.Id,
			Date:                    t.Date.ToGoTime().String(),
			Amount:                  t.Amount,
			LastUpdatedDate:         t.LastUpdatedDate.ToGoTime().String(),
			IsMarkedAsPaidConfirmed: t.IsMarkedAsPaidConfirmed,
			IsMarkedAsShipped:       t.IsMarkedAsShipped,
			IsShippingBooked:        t.IsShippingBooked,
		}
		if t.Buyer != nil {
			transactions[i].BuyerID = t.Buyer.Id
			transactions[i].BuyerAlias = t.Buyer.Alias
		}
		if t.Item != nil {
			transactions[i].ItemID = t.Item.Id
			transactions[i].ItemTitle = t.Item.Title
		}
	}

	return transactions, nil
}

// UserInfo represents user information.
type UserInfo struct {
	ID               int32
	Alias            string
	FirstName        string
	LastName         string
	Email            string
	PhoneNumber      string
	Address          string
	ZipCode          string
	City             string
	CountryName      string
	PersonalNumber   string
	CurrencyCode     string
	LanguageCodeIso2 string
}

// GetUserInfo retrieves information about the authenticated user.
func (c *RestrictedClient) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return nil, err
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*restricted.GetUserInfoResponse, error) {
		return c.service.GetUserInfoContext(ctx, &restricted.GetUserInfo{})
	})
	if err != nil {
		return nil, err
	}

	if result.GetUserInfoResult == nil {
		return nil, nil
	}

	u := result.GetUserInfoResult
	return &UserInfo{
		ID:               u.Id,
		Alias:            u.Alias,
		FirstName:        u.FirstName,
		LastName:         u.LastName,
		Email:            u.Email,
		PhoneNumber:      u.PhoneNumber,
		Address:          u.Address,
		ZipCode:          u.ZipCode,
		City:             u.City,
		CountryName:      u.CountryName,
		PersonalNumber:   u.PersonalNumber,
		CurrencyCode:     u.CurrencyCode,
		LanguageCodeIso2: u.LanguageCodeIso2,
	}, nil
}

// ShopSettings represents shop settings.
type ShopSettings struct {
	CompanyInformation     string
	PurchaseTerms          string
	ShowGalleryMode        *bool
	ShowAuctionView        *bool
	BannerColor            string
	IsTemporaryClosed      *bool
	TemporaryClosedMessage string
	ContactInformation     string
	LogoImageUrl           string
	MaxActiveItems         int32
	MaxInventoryItems      int32
}

// GetShopSettings retrieves the shop settings for the authenticated user.
func (c *RestrictedClient) GetShopSettings(ctx context.Context) (*ShopSettings, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return nil, err
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*restricted.GetShopSettingsResponse, error) {
		return c.service.GetShopSettingsContext(ctx, &restricted.GetShopSettings{})
	})
	if err != nil {
		return nil, err
	}

	if result.GetShopSettingsResult == nil {
		return nil, nil
	}

	s := result.GetShopSettingsResult
	return &ShopSettings{
		CompanyInformation:     s.CompanyInformation,
		PurchaseTerms:          s.PurchaseTerms,
		ShowGalleryMode:        s.ShowGalleryMode,
		ShowAuctionView:        s.ShowAuctionView,
		BannerColor:            s.BannerColor,
		IsTemporaryClosed:      s.IsTemporaryClosed,
		TemporaryClosedMessage: s.TemporaryClosedMessage,
		ContactInformation:     s.ContactInformation,
		LogoImageUrl:           s.LogoImageUrl,
		MaxActiveItems:         s.MaxActiveItems,
		MaxInventoryItems:      s.MaxInventoryItems,
	}, nil
}

// EndItem ends an active item.
func (c *RestrictedClient) EndItem(ctx context.Context, itemID int32) error {
	if err := RequireUserAuth(c.client.config); err != nil {
		return err
	}

	return c.client.executeWithMiddleware(ctx, func() error {
		_, err := c.service.EndItemContext(ctx, &restricted.EndItem{
			ItemId: itemID,
		})
		return err
	})
}
