package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Money int64

func ParseMoney(value string) (Money, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ".")
	if len(parts) > 2 || len(parts) == 0 || parts[0] == "" {
		return 0, errors.New("invalid money")
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, errors.New("invalid money")
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if len(fraction) > 2 {
		return 0, errors.New("money supports at most two decimal places")
	}
	fraction += strings.Repeat("0", 2-len(fraction))
	cents := int64(0)
	if fraction != "" {
		cents, err = strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return 0, errors.New("invalid money")
		}
	}
	if whole > (1<<63-1-cents)/100 {
		return 0, errors.New("money is too large")
	}
	return Money(whole*100 + cents), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) || data[0] == '"' {
		return errors.New("money must be a JSON number")
	}
	value, err := ParseMoney(string(data))
	if err != nil {
		return err
	}
	*m = value
	return nil
}

func (m Money) MarshalJSON() ([]byte, error) { return []byte(m.String()), nil }
func (m Money) String() string               { return fmt.Sprintf("%d.%02d", int64(m)/100, int64(m)%100) }

type Product struct {
	ID              string    `json:"id"`
	SellerID        string    `json:"seller_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	ImageKey        string    `json:"image_key"`
	StartingPrice   Money     `json:"starting_price"`
	CurrentPrice    Money     `json:"current_price"`
	HighestBidderID *string   `json:"highest_bidder_id"`
	AuctionEndAt    time.Time `json:"auction_end_at"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Bid struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	BidderID  string    `json:"bidder_id"`
	Amount    Money     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type BidResult struct {
	BidID           string `json:"bid_id"`
	ProductID       string `json:"product_id"`
	Amount          Money  `json:"amount"`
	CurrentPrice    Money  `json:"current_price"`
	HighestBidderID string `json:"highest_bidder_id"`
}

type SaleResult struct {
	ProductID  string    `json:"product_id"`
	Status     string    `json:"status"`
	BuyerID    string    `json:"buyer_id"`
	FinalPrice Money     `json:"final_price"`
	SoldAt     time.Time `json:"sold_at"`
}

var _ json.Marshaler = Money(0)
