package tradera

import (
	"context"
	"time"

	"github.com/hooklift/gowsdl/soap"
	"github.com/SebbeJohansson/tradera-go-client/generated/buyer"
)

// BuyerClient provides access to the Tradera Buyer API.
// Requires user authentication (UserID and Token in config).
type BuyerClient struct {
	client  *Client
	service buyer.BuyerServiceSoap
}

func newBuyerClient(c *Client) *BuyerClient {
	soapClient := c.createSOAPClient(BuyerServiceURL)
	return &BuyerClient{
		client:  c,
		service: buyer.NewBuyerServiceSoap(soapClient),
	}
}

// BuyResult represents the result of a buy operation.
type BuyResult struct {
	NextBid int32
	Status  string
}

// Buy purchases an item (Buy It Now).
func (c *BuyerClient) Buy(ctx context.Context, itemID int32, buyAmount int32) (*BuyResult, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return nil, err
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*buyer.BuyResponse, error) {
		return c.service.BuyContext(ctx, &buyer.Buy{
			ItemId:    itemID,
			BuyAmount: buyAmount,
		})
	})
	if err != nil {
		return nil, err
	}

	if result.BuyResult == nil {
		return nil, nil
	}

	status := ""
	if result.BuyResult.Status != nil {
		status = string(*result.BuyResult.Status)
	}

	return &BuyResult{
		NextBid: result.BuyResult.NextBid,
		Status:  status,
	}, nil
}

// MemorylistItem represents an item in the user's memory list (watchlist).
type MemorylistItem struct {
	ID                int32
	Title             string
	EndDate           time.Time
	CurrentPrice      int32
	BuyItNowPrice     *int32
	ThumbnailLink     string
	IsEnded           bool
	TotalBids         int32
	RemainingQuantity int32
}

// GetMemorylistItems retrieves items from the user's memory list (watchlist).
func (c *BuyerClient) GetMemorylistItems(ctx context.Context, filterActive *string, minEndDate, maxEndDate *time.Time) ([]*MemorylistItem, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return nil, err
	}

	req := &buyer.GetMemorylistItems{}

	if filterActive != nil {
		filter := buyer.ActiveFilter(*filterActive)
		req.FilterActive = &filter
	}

	if minEndDate != nil {
		dt := soap.CreateXsdDateTime(*minEndDate, true)
		req.MinEndDate = &dt
	}

	if maxEndDate != nil {
		dt := soap.CreateXsdDateTime(*maxEndDate, true)
		req.MaxEndDate = &dt
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*buyer.GetMemorylistItemsResponse, error) {
		return c.service.GetMemorylistItemsContext(ctx, req)
	})
	if err != nil {
		return nil, err
	}

	if result.GetMemorylistItemsResult == nil || result.GetMemorylistItemsResult.Item == nil {
		return nil, nil
	}

	items := make([]*MemorylistItem, len(result.GetMemorylistItemsResult.Item))
	for i, item := range result.GetMemorylistItemsResult.Item {
		isEnded := false
		if item.Status != nil {
			isEnded = item.Status.Ended
		}

		items[i] = &MemorylistItem{
			ID:                item.Id,
			Title:             item.ShortDescription,
			EndDate:           item.EndDate.ToGoTime(),
			CurrentPrice:      item.NextBid,
			BuyItNowPrice:     item.BuyItNowPrice,
			ThumbnailLink:     item.ThumbnailLink,
			IsEnded:           isEnded,
			TotalBids:         item.TotalBids,
			RemainingQuantity: item.RemainingQuantity,
		}
	}

	return items, nil
}

// AddToMemorylist adds items to the user's memory list (watchlist).
func (c *BuyerClient) AddToMemorylist(ctx context.Context, itemIDs []int32) error {
	if err := RequireUserAuth(c.client.config); err != nil {
		return err
	}

	return c.client.executeWithMiddleware(ctx, func() error {
		_, err := c.service.AddToMemorylistContext(ctx, &buyer.AddToMemorylist{
			ItemIds: &buyer.ArrayOfInt{},
		})
		return err
	})
}

// RemoveFromMemorylist removes items from the user's memory list (watchlist).
func (c *BuyerClient) RemoveFromMemorylist(ctx context.Context, itemIDs []int32) error {
	if err := RequireUserAuth(c.client.config); err != nil {
		return err
	}

	return c.client.executeWithMiddleware(ctx, func() error {
		_, err := c.service.RemoveFromMemorylistContext(ctx, &buyer.RemoveFromMemorylist{
			ItemIds: &buyer.ArrayOfInt{},
		})
		return err
	})
}

// BuyerTransaction represents a transaction from the buyer's perspective.
type BuyerTransaction struct {
	ID                      int32
	Date                    time.Time
	Amount                  int32
	LastUpdatedDate         time.Time
	IsMarkedAsPaid          bool
	IsMarkedAsPaidConfirmed bool
	IsMarkedAsShipped       bool
	IsFeedbackLeftBySeller  bool
	IsFeedbackLeftByBuyer   bool
	ItemID                  int32
	ItemTitle               string
	SellerID                int32
	SellerAlias             string
}

// GetBuyerTransactions retrieves transactions for the authenticated buyer.
func (c *BuyerClient) GetBuyerTransactions(ctx context.Context, minDate, maxDate *time.Time) ([]*BuyerTransaction, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return nil, err
	}

	req := &buyer.GetBuyerTransactions{
		Request: &buyer.GetBuyerTransactionsRequest{},
	}

	if minDate != nil {
		dt := soap.CreateXsdDateTime(*minDate, true)
		req.Request.MinTransactionDate = &dt
	}

	if maxDate != nil {
		dt := soap.CreateXsdDateTime(*maxDate, true)
		req.Request.MaxTransactionDate = &dt
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*buyer.GetBuyerTransactionsResponse, error) {
		return c.service.GetBuyerTransactionsContext(ctx, req)
	})
	if err != nil {
		return nil, err
	}

	if result.GetBuyerTransactionsResult == nil || result.GetBuyerTransactionsResult.Transaction == nil {
		return nil, nil
	}

	transactions := make([]*BuyerTransaction, len(result.GetBuyerTransactionsResult.Transaction))
	for i, t := range result.GetBuyerTransactionsResult.Transaction {
		tx := &BuyerTransaction{
			ID:                      t.Id,
			Date:                    t.Date.ToGoTime(),
			Amount:                  t.Amount,
			LastUpdatedDate:         t.LastUpdatedDate.ToGoTime(),
			IsMarkedAsPaid:          t.IsMarkedAsPaid,
			IsMarkedAsPaidConfirmed: t.IsMarkedAsPaidConfirmed,
			IsMarkedAsShipped:       t.IsMarkedAsShipped,
			IsFeedbackLeftBySeller:  t.IsFeedbackLeftBySeller,
			IsFeedbackLeftByBuyer:   t.IsFeedbackLeftByBuyer,
		}

		if t.Item != nil {
			tx.ItemID = t.Item.Id
			tx.ItemTitle = t.Item.Title
		}

		if t.Seller != nil {
			tx.SellerID = t.Seller.Id
			tx.SellerAlias = t.Seller.Alias
		}

		transactions[i] = tx
	}

	return transactions, nil
}

// AuctionBiddingInfo represents bidding information for an auction.
type AuctionBiddingInfo struct {
	ID                  int32
	ShortDescription    string
	SellerID            int32
	EndDate             time.Time
	ThumbnailLink       string
	ReservePriceReached bool
	HasBuyItNowOption   bool
	BuyItNowPrice       int32
	IsEnded             bool
	NextBid             int32
	MaxBid              int32
	MaxBidderID         int32
	TotalBids           int32
	MaxAutoBid          int32
}

// GetBiddingInfo retrieves bidding information for items the user has bid on.
func (c *BuyerClient) GetBiddingInfo(ctx context.Context, minDate, maxDate *time.Time, filterActive, filterLeading *string, includeHidden *bool) ([]*AuctionBiddingInfo, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return nil, err
	}

	req := &buyer.GetBiddingInfo{
		Request: &buyer.GetBiddingInfoRequest{},
	}

	if minDate != nil {
		dt := soap.CreateXsdDateTime(*minDate, true)
		req.Request.MinDate = &dt
	}

	if maxDate != nil {
		dt := soap.CreateXsdDateTime(*maxDate, true)
		req.Request.MaxDate = &dt
	}

	if filterActive != nil {
		filter := buyer.ActiveFilter(*filterActive)
		req.Request.FilterActive = &filter
	}

	if filterLeading != nil {
		filter := buyer.LeadingFilter(*filterLeading)
		req.Request.FilterLeading = &filter
	}

	req.Request.IncludeHidden = includeHidden

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*buyer.GetBiddingInfoResponse, error) {
		return c.service.GetBiddingInfoContext(ctx, req)
	})
	if err != nil {
		return nil, err
	}

	if result.GetBiddingInfoResult == nil || result.GetBiddingInfoResult.AuctionBiddingInfo == nil {
		return nil, nil
	}

	infos := make([]*AuctionBiddingInfo, len(result.GetBiddingInfoResult.AuctionBiddingInfo))
	for i, info := range result.GetBiddingInfoResult.AuctionBiddingInfo {
		infos[i] = &AuctionBiddingInfo{
			ID:                  info.Id,
			ShortDescription:    info.ShortDescription,
			SellerID:            info.SellerId,
			EndDate:             info.EndDate.ToGoTime(),
			ThumbnailLink:       info.ThumbnailLink,
			ReservePriceReached: info.ReservePriceReached,
			HasBuyItNowOption:   info.HasBuyItNowOption,
			BuyItNowPrice:       info.BuyItNowPrice,
			IsEnded:             info.IsEnded,
			NextBid:             info.NextBid,
			MaxBid:              info.MaxBid,
			MaxBidderID:         info.MaxBidderId,
			TotalBids:           info.TotalBids,
			MaxAutoBid:          info.MaxAutoBid,
		}
	}

	return infos, nil
}

// SellerInfo represents public information about a seller.
type SellerInfo struct {
	TotalRating             int32
	PositiveFeedbackPercent *int32
	PersonalMessage         string
	IsCompany               bool
	HasShop                 bool
	DetailedRating          *DetailedSellerRating
}

// DetailedSellerRating contains detailed seller rating information.
type DetailedSellerRating struct {
	ItemAsDescribedCount           *int32
	ItemAsDescribedAverage         *float64
	CommResponsivenessCount        *int32
	CommResponsivenessAverage      *float64
	ShippingTimeCount              *int32
	ShippingTimeAverage            *float64
	ShippingHandlingChargesCount   *int32
	ShippingHandlingChargesAverage *float64
}

// GetSellerInfo retrieves public information about a seller.
func (c *BuyerClient) GetSellerInfo(ctx context.Context, userID int32) (*SellerInfo, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return nil, err
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*buyer.GetSellerInfoResponse, error) {
		return c.service.GetSellerInfoContext(ctx, &buyer.GetSellerInfo{
			UserId: userID,
		})
	})
	if err != nil {
		return nil, err
	}

	if result.GetSellerInfoResult == nil {
		return nil, nil
	}

	info := &SellerInfo{
		TotalRating:             result.GetSellerInfoResult.TotalRating,
		PositiveFeedbackPercent: result.GetSellerInfoResult.PositiveFeedbackPercent,
		PersonalMessage:         result.GetSellerInfoResult.PersonalMessage,
		IsCompany:               result.GetSellerInfoResult.IsCompany,
		HasShop:                 result.GetSellerInfoResult.HasShop,
	}

	if result.GetSellerInfoResult.DetailedSellerRating != nil {
		dsr := result.GetSellerInfoResult.DetailedSellerRating
		info.DetailedRating = &DetailedSellerRating{
			ItemAsDescribedCount:           dsr.ItemAsDescribedCount,
			ItemAsDescribedAverage:         dsr.ItemAsDescribedAverage,
			CommResponsivenessCount:        dsr.CommResponsivenessCount,
			CommResponsivenessAverage:      dsr.CommResponsivenessAverage,
			ShippingTimeCount:              dsr.ShippingTimeCount,
			ShippingTimeAverage:            dsr.ShippingTimeAverage,
			ShippingHandlingChargesCount:   dsr.ShippingHandlingChargesCount,
			ShippingHandlingChargesAverage: dsr.ShippingHandlingChargesAverage,
		}
	}

	return info, nil
}

// MarkTransactionsPaid marks transactions as paid by the buyer.
func (c *BuyerClient) MarkTransactionsPaid(ctx context.Context, transactionIDs []int32, markedAsPaid bool) error {
	if err := RequireUserAuth(c.client.config); err != nil {
		return err
	}

	requests := make([]*buyer.MarkTransactionsPaidRequest, len(transactionIDs))
	for i, id := range transactionIDs {
		requests[i] = &buyer.MarkTransactionsPaidRequest{
			TransactionId:     id,
			MarkedAsPaidValue: markedAsPaid,
		}
	}

	return c.client.executeWithMiddleware(ctx, func() error {
		_, err := c.service.MarkTransactionsPaidContext(ctx, &buyer.MarkTransactionsPaid{
			Request: &buyer.ArrayOfMarkTransactionsPaidRequest{
				MarkTransactionsPaidRequest: requests,
			},
		})
		return err
	})
}

// SendQuestionToSeller sends a question to the seller of an item.
func (c *BuyerClient) SendQuestionToSeller(ctx context.Context, itemID int32, question string, sendCopyToSender bool) (string, error) {
	if err := RequireUserAuth(c.client.config); err != nil {
		return "", err
	}

	result, err := executeWithMiddlewareResult(c.client, ctx, func() (*buyer.SendQuestionToSellerResponse, error) {
		return c.service.SendQuestionToSellerContext(ctx, &buyer.SendQuestionToSeller{
			ItemId:           itemID,
			Question:         question,
			SendCopyToSender: sendCopyToSender,
		})
	})
	if err != nil {
		return "", err
	}

	if result.SendQuestionToSellerResult == nil {
		return "", nil
	}

	return result.SendQuestionToSellerResult.Status, nil
}
