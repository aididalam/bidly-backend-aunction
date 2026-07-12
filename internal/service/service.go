package service

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"auction/auction/internal/model"
	"auction/auction/internal/repository"
	"auction/auction/internal/upload"
	"github.com/google/uuid"
)

var ErrInvalid = errors.New("invalid input")

type ProductInput struct {
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	ImageKey      string      `json:"image_key"`
	StartingPrice model.Money `json:"starting_price"`
	AuctionEndAt  time.Time   `json:"auction_end_at"`
}
type Service struct {
	repo    repository.Repository
	uploads upload.Generator
	now     func() time.Time
}

func New(repo repository.Repository, uploads upload.Generator) *Service {
	return &Service{repo: repo, uploads: uploads, now: time.Now}
}
func (s *Service) Create(ctx context.Context, seller string, in ProductInput) (model.Product, error) {
	if !validProduct(seller, &in, s.now()) {
		return model.Product{}, ErrInvalid
	}
	return s.repo.CreateProduct(ctx, model.Product{SellerID: seller, Title: in.Title, Description: in.Description, ImageKey: in.ImageKey, StartingPrice: in.StartingPrice, AuctionEndAt: in.AuctionEndAt})
}
func (s *Service) List(ctx context.Context) ([]model.Product, error) {
	return s.repo.ListActive(ctx, s.now())
}
func (s *Service) Get(ctx context.Context, id string) (model.Product, error) {
	if uuid.Validate(id) != nil {
		return model.Product{}, repository.ErrNotFound
	}
	return s.repo.GetProduct(ctx, id)
}
func (s *Service) Mine(ctx context.Context, seller string) ([]model.Product, error) {
	return s.repo.ListSeller(ctx, seller)
}
func (s *Service) Update(ctx context.Context, id, seller string, in ProductInput) (model.Product, error) {
	if uuid.Validate(id) != nil || !validProduct(seller, &in, s.now()) {
		return model.Product{}, ErrInvalid
	}
	return s.repo.UpdateProduct(ctx, seller, model.Product{ID: id, Title: in.Title, Description: in.Description, ImageKey: in.ImageKey, StartingPrice: in.StartingPrice, AuctionEndAt: in.AuctionEndAt}, s.now())
}
func (s *Service) Cancel(ctx context.Context, id, seller string) error {
	if uuid.Validate(id) != nil {
		return repository.ErrNotFound
	}
	return s.repo.CancelProduct(ctx, id, seller, s.now())
}
func (s *Service) Bid(ctx context.Context, id, bidder string, amount model.Money) (model.BidResult, error) {
	if uuid.Validate(id) != nil {
		return model.BidResult{}, repository.ErrNotFound
	}
	if amount <= 0 || amount > 999999999999 {
		return model.BidResult{}, ErrInvalid
	}
	return s.repo.PlaceBid(ctx, id, bidder, amount, s.now())
}
func (s *Service) Bids(ctx context.Context, id, seller string) ([]model.Bid, error) {
	if uuid.Validate(id) != nil {
		return nil, repository.ErrNotFound
	}
	return s.repo.ListBids(ctx, id, seller)
}
func (s *Service) Sell(ctx context.Context, id, seller string) (model.SaleResult, error) {
	if uuid.Validate(id) != nil {
		return model.SaleResult{}, repository.ErrNotFound
	}
	return s.repo.Sell(ctx, id, seller, s.now())
}
func (s *Service) Upload(ctx context.Context, user, filename, contentType string) (upload.Result, error) {
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	extension, ok := extensions[contentType]
	provided := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	if provided == ".jpeg" {
		provided = ".jpg"
	}
	if !ok || provided != extension {
		return upload.Result{}, ErrInvalid
	}
	result, err := s.uploads.Generate(ctx, user, filename, contentType)
	if err != nil {
		return upload.Result{}, err
	}
	return result, nil
}
func (s *Service) Expire(ctx context.Context) (int64, error) { return s.repo.Expire(ctx, s.now()) }
func validProduct(seller string, in *ProductInput, now time.Time) bool {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.ImageKey = strings.TrimSpace(in.ImageKey)
	return in.Title != "" && utf8.RuneCountInString(in.Title) <= 200 && in.Description != "" && len(in.Description) <= 10000 && strings.HasPrefix(in.ImageKey, "products/"+seller+"/") && in.StartingPrice > 0 && in.StartingPrice <= 999999999999 && in.AuctionEndAt.After(now)
}
