package tradera

import (
	"context"
	"time"

	"github.com/pristabell/tradera-api-client/generated/public"
)

// PublicClient provides access to the Tradera Public API.
type PublicClient struct {
	client  *Client
	service public.PublicServiceSoap
}

func newPublicClient(c *Client) *PublicClient {
	soapClient := c.createSOAPClient(PublicServiceURL)
	return &PublicClient{
		client:  c,
		service: public.NewPublicServiceSoap(soapClient),
	}
}

// Item represents a Tradera item with full details.
type Item struct {
	ID                int32
	ShortDescription  string
	LongDescription   string
	StartDate         time.Time
	EndDate           time.Time
	CategoryID        int32
	OpeningBid        int32
	ReservePrice      *int32
	BuyItNowPrice     *int32
	NextBid           int32
	MaxBid            int32
	TotalBids         int32
	ItemType          string
	StartQuantity     int32
	RemainingQuantity int32
	VAT               *int32
	PaymentCondition  string
	ShippingCondition string
	AcceptsPickup     bool
	Bold              bool
	Thumbnail         bool
	Highlight         bool
	FeaturedItem      bool
	ItemLink          string
	ThumbnailLink     string
	Restarts          int32
	Duration          int32
	Seller            *User
	MaxBidder         *User
	Buyers            []*User
	ImageLinks        []string
	ShippingOptions   []*ItemShipping
}

// User represents a Tradera user.
type User struct {
	ID                int32
	Alias             string
	FirstName         string
	LastName          string
	Email             string
	TotalRating       int32
	PhoneNumber       string
	MobilePhoneNumber string
	Address           string
	ZipCode           string
	City              string
	CountryName       string
}

// ItemShipping represents shipping options for an item.
type ItemShipping struct {
	ShippingOptionID   *int32
	Cost               int32
	ShippingWeight     *float64
	ShippingProductID  *int32
	ShippingProviderID *int32
}

// Category represents a Tradera category.
type Category struct {
	ID       int32
	Name     string
	Children []*Category
}

// GetItem retrieves detailed information about a specific item.
func (c *PublicClient) GetItem(ctx context.Context, itemID int32) (*Item, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*public.GetItemResponse, error) {
		return c.service.GetItemContext(ctx, &public.GetItem{
			ItemId: itemID,
		})
	})
	if err != nil {
		return nil, err
	}

	return convertPublicItem(result.GetItemResult), nil
}

// GetUserByAlias retrieves a user by their alias.
func (c *PublicClient) GetUserByAlias(ctx context.Context, alias string) (*User, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*public.GetUserByAliasResponse, error) {
		return c.service.GetUserByAliasContext(ctx, &public.GetUserByAlias{
			Alias: alias,
		})
	})
	if err != nil {
		return nil, err
	}

	return convertUser(result.GetUserByAliasResult), nil
}

// FetchToken retrieves an authorization token for a user.
// This token is required for authenticated operations.
func (c *PublicClient) FetchToken(ctx context.Context, userID int32, secretKey string) (string, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*public.FetchTokenResponse, error) {
		return c.service.FetchTokenContext(ctx, &public.FetchToken{
			UserId:    userID,
			SecretKey: secretKey,
		})
	})
	if err != nil {
		return "", err
	}

	return result.FetchTokenResult, nil
}

// GetOfficialTime retrieves the official Tradera server time.
func (c *PublicClient) GetOfficialTime(ctx context.Context) (time.Time, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*public.GetOfficalTimeResponse, error) {
		return c.service.GetOfficalTimeContext(ctx, &public.GetOfficalTime{})
	})
	if err != nil {
		return time.Time{}, err
	}

	return result.GetOfficalTimeResult.ToGoTime(), nil
}

// GetCategories retrieves the full category tree.
func (c *PublicClient) GetCategories(ctx context.Context) ([]*Category, error) {
	// Check cache first
	if c.client.cache != nil {
		if cached, ok := c.client.cache.Get("categories"); ok {
			return cached.([]*Category), nil
		}
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*public.GetCategoriesResponse, error) {
		return c.service.GetCategoriesContext(ctx, &public.GetCategories{})
	})
	if err != nil {
		return nil, err
	}

	categories := convertCategories(result.GetCategoriesResult)

	// Cache the result
	if c.client.cache != nil {
		c.client.cache.Set("categories", categories)
	}

	return categories, nil
}

// GetSellerItems retrieves items for a specific seller.
func (c *PublicClient) GetSellerItems(ctx context.Context, userID int32, categoryID int32) ([]*Item, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*public.GetSellerItemsResponse, error) {
		return c.service.GetSellerItemsContext(ctx, &public.GetSellerItems{
			UserId:     userID,
			CategoryId: categoryID,
		})
	})
	if err != nil {
		return nil, err
	}

	return convertPublicItems(result.GetSellerItemsResult), nil
}

// GetCounties retrieves the list of Swedish counties.
func (c *PublicClient) GetCounties(ctx context.Context) ([]*IdDescriptionPair, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*public.GetCountiesResponse, error) {
		return c.service.GetCountiesContext(ctx, &public.GetCounties{})
	})
	if err != nil {
		return nil, err
	}

	return convertIdDescriptionPairs(result.GetCountiesResult), nil
}

// IdDescriptionPair represents a simple ID-description pair.
type IdDescriptionPair struct {
	ID          int32
	Description string
	Value       string
}

// Conversion helpers

func convertPublicItem(item *public.Item) *Item {
	if item == nil {
		return nil
	}

	i := &Item{
		ID:                item.Id,
		ShortDescription:  item.ShortDescription,
		LongDescription:   item.LongDescription,
		StartDate:         item.StartDate.ToGoTime(),
		EndDate:           item.EndDate.ToGoTime(),
		CategoryID:        item.CategoryId,
		OpeningBid:        item.OpeningBid,
		ReservePrice:      item.ReservePrice,
		BuyItNowPrice:     item.BuyItNowPrice,
		NextBid:           item.NextBid,
		MaxBid:            item.MaxBid,
		TotalBids:         item.TotalBids,
		StartQuantity:     item.StartQuantity,
		RemainingQuantity: item.RemainingQuantity,
		VAT:               item.VAT,
		PaymentCondition:  item.PaymentCondition,
		ShippingCondition: item.ShippingCondition,
		AcceptsPickup:     item.AcceptsPickup,
		Bold:              item.Bold,
		Thumbnail:         item.Thumbnail,
		Highlight:         item.Highlight,
		FeaturedItem:      item.FeaturedItem,
		ItemLink:          item.ItemLink,
		ThumbnailLink:     item.ThumbnailLink,
		Restarts:          item.Restarts,
		Duration:          item.Duration,
	}

	if item.ItemType != nil {
		i.ItemType = string(*item.ItemType)
	}

	if item.Seller != nil {
		i.Seller = convertUser(item.Seller)
	}

	if item.MaxBidder != nil {
		i.MaxBidder = convertUser(item.MaxBidder)
	}

	if item.Buyers != nil {
		i.Buyers = make([]*User, len(item.Buyers))
		for idx, buyer := range item.Buyers {
			i.Buyers[idx] = convertUser(buyer)
		}
	}

	if item.ImageLinks != nil && item.ImageLinks.Astring != nil {
		i.ImageLinks = make([]string, len(item.ImageLinks.Astring))
		for idx, link := range item.ImageLinks.Astring {
			if link != nil {
				i.ImageLinks[idx] = *link
			}
		}
	}

	if item.ShippingOptions != nil {
		i.ShippingOptions = make([]*ItemShipping, len(item.ShippingOptions))
		for idx, opt := range item.ShippingOptions {
			i.ShippingOptions[idx] = &ItemShipping{
				ShippingOptionID:   opt.ShippingOptionId,
				Cost:               opt.Cost,
				ShippingWeight:     opt.ShippingWeight,
				ShippingProductID:  opt.ShippingProductId,
				ShippingProviderID: opt.ShippingProviderId,
			}
		}
	}

	return i
}

func convertPublicItems(items *public.ArrayOfItem) []*Item {
	if items == nil || items.Item == nil {
		return nil
	}

	result := make([]*Item, len(items.Item))
	for i, item := range items.Item {
		result[i] = convertPublicItem(item)
	}
	return result
}

func convertUser(user *public.User) *User {
	if user == nil {
		return nil
	}

	return &User{
		ID:                user.Id,
		Alias:             user.Alias,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Email:             user.Email,
		TotalRating:       user.TotalRating,
		PhoneNumber:       user.PhoneNumber,
		MobilePhoneNumber: user.MobilePhoneNumber,
		Address:           user.Address,
		ZipCode:           user.ZipCode,
		City:              user.City,
		CountryName:       user.CountryName,
	}
}

func convertCategories(cats *public.ArrayOfCategory) []*Category {
	if cats == nil || cats.Category == nil {
		return nil
	}

	result := make([]*Category, len(cats.Category))
	for i, cat := range cats.Category {
		result[i] = convertCategory(cat)
	}
	return result
}

func convertCategory(cat *public.Category) *Category {
	if cat == nil {
		return nil
	}

	c := &Category{
		ID:   cat.Id,
		Name: cat.Name,
	}

	if cat.Category != nil {
		c.Children = make([]*Category, len(cat.Category))
		for i, child := range cat.Category {
			c.Children[i] = convertCategory(child)
		}
	}

	return c
}

func convertIdDescriptionPairs(pairs *public.ArrayOfIdDescriptionPair) []*IdDescriptionPair {
	if pairs == nil || pairs.IdDescriptionPair == nil {
		return nil
	}

	result := make([]*IdDescriptionPair, len(pairs.IdDescriptionPair))
	for i, pair := range pairs.IdDescriptionPair {
		result[i] = &IdDescriptionPair{
			ID:          pair.Id,
			Description: pair.Description,
			Value:       pair.Value,
		}
	}
	return result
}
