package tradera

import (
	"context"

	"github.com/hooklift/gowsdl/soap"
	"github.com/SebbeJohansson/tradera-go-client/generated/search"
)

// SearchClient provides access to the Tradera Search API.
type SearchClient struct {
	client  *Client
	service search.SearchServiceSoap
}

func newSearchClient(c *Client) *SearchClient {
	soapClient := c.createSOAPClient(SearchServiceURL)
	return &SearchClient{
		client:  c,
		service: search.NewSearchServiceSoap(soapClient),
	}
}

// SearchRequest contains parameters for a basic search.
type SearchRequest struct {
	Query      string
	CategoryID int32
	PageNumber int32
	OrderBy    string
}

// SearchResult represents the result of a search operation.
type SearchResult struct {
	TotalNumberOfItems int32
	TotalNumberOfPages int32
	Items              []*SearchItem
	Errors             []*SearchError
}

// SearchItem represents an item in search results.
type SearchItem struct {
	ID               int32
	ShortDescription string
	LongDescription  string
	BuyItNowPrice    *int32
	SellerID         int32
	SellerAlias      string
	MaxBid           *int32
	ThumbnailLink    string
	SellerDsrAverage float64
	EndDate          soap.XSDDateTime
	NextBid          *int32
	HasBids          bool
	IsEnded          bool
	ItemType         string
	ItemURL          string
	CategoryID       int32
	BidCount         int32
	ImageLinks       []ImageLink
}

// ImageLink represents an image URL with format information.
type ImageLink struct {
	URL    string
	Format string
}

// SearchError represents an error from the search API.
type SearchError struct {
	Code    string
	Message string
}

// Search performs a basic item search.
func (c *SearchClient) Search(ctx context.Context, query string, categoryID int32) (*SearchResult, error) {
	return c.SearchWithOptions(ctx, SearchRequest{
		Query:      query,
		CategoryID: categoryID,
		PageNumber: 1,
	})
}

// SearchWithOptions performs a search with custom options.
func (c *SearchClient) SearchWithOptions(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*search.SearchResponse, error) {
		return c.service.SearchContext(ctx, &search.Search{
			Query:      req.Query,
			CategoryId: req.CategoryID,
			PageNumber: req.PageNumber,
			OrderBy:    req.OrderBy,
		})
	})
	if err != nil {
		return nil, err
	}

	return convertSearchResult(result.SearchResult), nil
}

// SearchAdvancedRequest contains parameters for an advanced search.
type SearchAdvancedRequest struct {
	SearchWords            string
	CategoryID             int32
	SearchInDescription    bool
	Mode                   string // "AllWords" or "AnyWords"
	PriceMinimum           *int32
	PriceMaximum           *int32
	BidsMinimum            *int32
	BidsMaximum            *int32
	ZipCode                string
	CountyID               int32
	Alias                  string
	OrderBy                string
	ItemStatus             string // "Active" or "Ended"
	ItemType               string // "All", "Auction", "FixedPrice"
	OnlyAuctionsWithBuyNow bool
	OnlyItemsWithThumbnail bool
	ItemsPerPage           int32
	PageNumber             int32
	ItemCondition          string // "All", "OnlyNew", "OnlySecondHand"
	SellerType             string // "All", "OnlyPrivate", "OnlyBusiness"
	Brands                 []string
}

// SearchAdvanced performs an advanced search with filters.
func (c *SearchClient) SearchAdvanced(ctx context.Context, req SearchAdvancedRequest) (*SearchResult, error) {
	advReq := &search.SearchAdvancedRequest{
		SearchWords:            req.SearchWords,
		CategoryId:             req.CategoryID,
		SearchInDescription:    req.SearchInDescription,
		Mode:                   req.Mode,
		PriceMinimum:           req.PriceMinimum,
		PriceMaximum:           req.PriceMaximum,
		BidsMinimum:            req.BidsMinimum,
		BidsMaximum:            req.BidsMaximum,
		ZipCode:                req.ZipCode,
		CountyId:               req.CountyID,
		Alias:                  req.Alias,
		OrderBy:                req.OrderBy,
		ItemStatus:             req.ItemStatus,
		ItemType:               req.ItemType,
		OnlyAuctionsWithBuyNow: req.OnlyAuctionsWithBuyNow,
		OnlyItemsWithThumbnail: req.OnlyItemsWithThumbnail,
		ItemsPerPage:           req.ItemsPerPage,
		PageNumber:             req.PageNumber,
		ItemCondition:          req.ItemCondition,
		SellerType:             req.SellerType,
	}

	if len(req.Brands) > 0 {
		brands := make([]*string, len(req.Brands))
		for i := range req.Brands {
			brands[i] = &req.Brands[i]
		}
		advReq.Brands = &search.ArrayOfString{Astring: brands}
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*search.SearchAdvancedResponse, error) {
		return c.service.SearchAdvancedContext(ctx, &search.SearchAdvanced{
			Request: advReq,
		})
	})
	if err != nil {
		return nil, err
	}

	return convertSearchResult(result.SearchAdvancedResult), nil
}

// CategoryCountRequest contains parameters for a category count search.
type CategoryCountRequest struct {
	CategoryID             int32
	SearchWords            string
	Alias                  string
	CountyID               int32
	SearchInDescription    bool
	ItemCondition          string
	ZipCode                string
	OnlyItemsWithThumbnail bool
	OnlyAuctionsWithBuyNow bool
	Mode                   string
	PriceMinimum           *int32
	PriceMaximum           *int32
	BidsMinimum            *int32
	BidsMaximum            *int32
	ItemStatus             string
	ItemType               string
	SellerType             string
}

// CategoryCountResult represents the result of a category count search.
type CategoryCountResult struct {
	Categories []*SearchCategory
	Errors     []*SearchError
}

// SearchCategory represents a category with item counts.
type SearchCategory struct {
	ID                                   int32
	Name                                 string
	NoOfItemsInCategory                  int32
	NoOfItemsInCategoryIncludingChildren int32
	ChildCategories                      []*SearchCategory
}

// SearchCategoryCount gets item counts per category.
func (c *SearchClient) SearchCategoryCount(ctx context.Context, req CategoryCountRequest) (*CategoryCountResult, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*search.SearchCategoryCountResponse, error) {
		return c.service.SearchCategoryCountContext(ctx, &search.SearchCategoryCount{
			Request: &search.CategoryCountRequest{
				CategoryId:             req.CategoryID,
				SearchWords:            req.SearchWords,
				Alias:                  req.Alias,
				CountyId:               req.CountyID,
				SearchInDescription:    req.SearchInDescription,
				ItemCondition:          req.ItemCondition,
				ZipCode:                req.ZipCode,
				OnlyItemsWithThumbnail: req.OnlyItemsWithThumbnail,
				OnlyAuctionsWithBuyNow: req.OnlyAuctionsWithBuyNow,
				Mode:                   req.Mode,
				PriceMinimum:           req.PriceMinimum,
				PriceMaximum:           req.PriceMaximum,
				BidsMinimum:            req.BidsMinimum,
				BidsMaximum:            req.BidsMaximum,
				ItemStatus:             req.ItemStatus,
				ItemType:               req.ItemType,
				SellerType:             req.SellerType,
			},
		})
	})
	if err != nil {
		return nil, err
	}

	return convertCategoryCountResult(result.SearchCategoryCountResult), nil
}

// SearchByZipCode searches items by zip code.
func (c *SearchClient) SearchByZipCode(ctx context.Context, zipCode string, pageNumber int32, orderBy string) (*SearchResult, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*search.SearchByZipCodeResponse, error) {
		return c.service.SearchByZipCodeContext(ctx, &search.SearchByZipCode{
			Request: &search.SearchByZipCodeRequest{
				ZipCode:    zipCode,
				PageNumber: pageNumber,
				OrderBy:    orderBy,
			},
		})
	})
	if err != nil {
		return nil, err
	}

	return convertSearchResult(result.SearchByZipCodeResult), nil
}

// SearchByFixedCriteria searches items by predefined criteria.
func (c *SearchClient) SearchByFixedCriteria(ctx context.Context, name string, pageNumber int32, itemType string, orderBy string) (*SearchResult, error) {
	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*search.SearchByFixedCriteriaResponse, error) {
		return c.service.SearchByFixedCriteriaContext(ctx, &search.SearchByFixedCriteria{
			Request: &search.SearchByFixedCriteriaRequest{
				Name:       name,
				PageNumber: pageNumber,
				ItemType:   itemType,
				OrderBy:    orderBy,
			},
		})
	})
	if err != nil {
		return nil, err
	}

	return convertSearchResult(result.SearchByFixedCriteriaResult), nil
}

// Helper functions to convert generated types to our types

func convertSearchResult(r *search.SearchResult) *SearchResult {
	if r == nil {
		return nil
	}

	result := &SearchResult{
		TotalNumberOfItems: r.TotalNumberOfItems,
		TotalNumberOfPages: r.TotalNumberOfPages,
	}

	if r.Items != nil {
		result.Items = make([]*SearchItem, len(r.Items))
		for i, item := range r.Items {
			result.Items[i] = convertSearchItem(item)
		}
	}

	if r.Errors != nil {
		result.Errors = make([]*SearchError, len(r.Errors))
		for i, err := range r.Errors {
			result.Errors[i] = &SearchError{
				Code:    err.Code,
				Message: err.Message,
			}
		}
	}

	return result
}

func convertSearchItem(item *search.SearchItem) *SearchItem {
	if item == nil {
		return nil
	}

	si := &SearchItem{
		ID:               item.Id,
		ShortDescription: item.ShortDescription,
		LongDescription:  item.LongDescription,
		BuyItNowPrice:    item.BuyItNowPrice,
		SellerID:         item.SellerId,
		SellerAlias:      item.SellerAlias,
		MaxBid:           item.MaxBid,
		ThumbnailLink:    item.ThumbnailLink,
		SellerDsrAverage: item.SellerDsrAverage,
		EndDate:          item.EndDate,
		NextBid:          item.NextBid,
		HasBids:          item.HasBids,
		IsEnded:          item.IsEnded,
		ItemType:         item.ItemType,
		ItemURL:          item.ItemUrl,
		CategoryID:       item.CategoryId,
		BidCount:         item.BidCount,
	}

	if item.ImageLinks != nil && item.ImageLinks.ImageLink != nil {
		si.ImageLinks = make([]ImageLink, len(item.ImageLinks.ImageLink))
		for i, link := range item.ImageLinks.ImageLink {
			si.ImageLinks[i] = ImageLink{
				URL:    link.Url,
				Format: link.Format,
			}
		}
	}

	return si
}

func convertCategoryCountResult(r *search.CategoryCountResult) *CategoryCountResult {
	if r == nil {
		return nil
	}

	result := &CategoryCountResult{}

	if r.Categories != nil {
		result.Categories = make([]*SearchCategory, len(r.Categories))
		for i, cat := range r.Categories {
			result.Categories[i] = convertSearchCategory(cat)
		}
	}

	if r.Errors != nil {
		result.Errors = make([]*SearchError, len(r.Errors))
		for i, err := range r.Errors {
			result.Errors[i] = &SearchError{
				Code:    err.Code,
				Message: err.Message,
			}
		}
	}

	return result
}

func convertSearchCategory(c *search.SearchCategory) *SearchCategory {
	if c == nil {
		return nil
	}

	cat := &SearchCategory{
		ID:                                   c.Id,
		Name:                                 c.Name,
		NoOfItemsInCategory:                  c.NoOfItemsInCategory,
		NoOfItemsInCategoryIncludingChildren: c.NoOfItemsInCategoryIncludingChildren,
	}

	if c.ChildCategories != nil {
		cat.ChildCategories = make([]*SearchCategory, len(c.ChildCategories))
		for i, child := range c.ChildCategories {
			cat.ChildCategories[i] = convertSearchCategory(child)
		}
	}

	return cat
}
