package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"auction/auction/internal/model"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
	ErrConflict  = errors.New("state conflict")
)

type Repository interface {
	CreateProduct(context.Context, model.Product) (model.Product, error)
	ListActive(context.Context, time.Time) ([]model.Product, error)
	GetProduct(context.Context, string) (model.Product, error)
	ListSeller(context.Context, string) ([]model.Product, error)
	UpdateProduct(context.Context, string, model.Product, time.Time) (model.Product, error)
	CancelProduct(context.Context, string, string, time.Time) error
	PlaceBid(context.Context, string, string, model.Money, time.Time) (model.BidResult, error)
	ListBids(context.Context, string, string) ([]model.Bid, error)
	Sell(context.Context, string, string, time.Time) (model.SaleResult, error)
	Expire(context.Context, time.Time) (int64, error)
}

type MySQL struct{ db *sql.DB }

func NewMySQL(db *sql.DB) *MySQL { return &MySQL{db: db} }

func (r *MySQL) CreateProduct(ctx context.Context, p model.Product) (model.Product, error) {
	p.ID = uuid.NewString()
	_, err := r.db.ExecContext(ctx, `INSERT INTO products (id,seller_id,title,description,image_key,currency,starting_price,current_price,auction_end_at,status) VALUES (?,?,?,?,?,?,?,?,?, 'ACTIVE')`, p.ID, p.SellerID, p.Title, p.Description, p.ImageKey, p.Currency, p.StartingPrice.String(), p.StartingPrice.String(), p.AuctionEndAt.UTC())
	if err != nil {
		return model.Product{}, err
	}
	return r.GetProduct(ctx, p.ID)
}

const productColumns = `id,seller_id,title,description,image_key,currency,starting_price,current_price,highest_bidder_id,auction_end_at,status,created_at,updated_at`

func (r *MySQL) ListActive(ctx context.Context, now time.Time) ([]model.Product, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+productColumns+` FROM products WHERE status='ACTIVE' AND auction_end_at > ? ORDER BY created_at DESC`, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}
func (r *MySQL) GetProduct(ctx context.Context, id string) (model.Product, error) {
	return scanProduct(r.db.QueryRowContext(ctx, `SELECT `+productColumns+` FROM products WHERE id=?`, id))
}
func (r *MySQL) ListSeller(ctx context.Context, seller string) ([]model.Product, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+productColumns+` FROM products WHERE seller_id=? ORDER BY created_at DESC`, seller)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (r *MySQL) UpdateProduct(ctx context.Context, seller string, next model.Product, now time.Time) (model.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Product{}, err
	}
	defer tx.Rollback()
	current, err := lockedProduct(ctx, tx, next.ID)
	if err != nil {
		return model.Product{}, err
	}
	if current.SellerID != seller {
		return model.Product{}, ErrForbidden
	}
	if current.Status != "ACTIVE" || !now.Before(current.AuctionEndAt) {
		return model.Product{}, ErrConflict
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM bids WHERE product_id=?`, next.ID).Scan(&count); err != nil {
		return model.Product{}, err
	}
	if count != 0 {
		return model.Product{}, ErrConflict
	}
	_, err = tx.ExecContext(ctx, `UPDATE products SET title=?,description=?,image_key=?,currency=?,starting_price=?,current_price=?,auction_end_at=? WHERE id=?`, next.Title, next.Description, next.ImageKey, next.Currency, next.StartingPrice.String(), next.StartingPrice.String(), next.AuctionEndAt.UTC(), next.ID)
	if err != nil {
		return model.Product{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.Product{}, err
	}
	return r.GetProduct(ctx, next.ID)
}

func (r *MySQL) CancelProduct(ctx context.Context, id, seller string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	p, err := lockedProduct(ctx, tx, id)
	if err != nil {
		return err
	}
	if p.SellerID != seller {
		return ErrForbidden
	}
	if p.Status != "ACTIVE" || !now.Before(p.AuctionEndAt) {
		return ErrConflict
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM bids WHERE product_id=?`, id).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return ErrConflict
	}
	if _, err = tx.ExecContext(ctx, `UPDATE products SET status='CANCELLED' WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *MySQL) PlaceBid(ctx context.Context, productID, bidder string, amount model.Money, now time.Time) (model.BidResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.BidResult{}, err
	}
	defer tx.Rollback()
	p, err := lockedProduct(ctx, tx, productID)
	if err != nil {
		return model.BidResult{}, err
	}
	if p.SellerID == bidder {
		return model.BidResult{}, ErrForbidden
	}
	if p.Status != "ACTIVE" || !now.Before(p.AuctionEndAt) || amount <= p.CurrentPrice {
		return model.BidResult{}, ErrConflict
	}
	id := uuid.NewString()
	if _, err = tx.ExecContext(ctx, `INSERT INTO bids (id,product_id,bidder_id,amount) VALUES (?,?,?,?)`, id, productID, bidder, amount.String()); err != nil {
		return model.BidResult{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE products SET current_price=?,highest_bidder_id=? WHERE id=?`, amount.String(), bidder, productID); err != nil {
		return model.BidResult{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.BidResult{}, err
	}
	return model.BidResult{BidID: id, ProductID: productID, Amount: amount, CurrentPrice: amount, HighestBidderID: bidder}, nil
}

func (r *MySQL) ListBids(ctx context.Context, productID, seller string) ([]model.Bid, error) {
	p, err := r.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if p.SellerID != seller {
		return nil, ErrForbidden
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,product_id,bidder_id,amount,created_at FROM bids WHERE product_id=? ORDER BY amount DESC,created_at DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.Bid{}
	for rows.Next() {
		var b model.Bid
		var amount string
		if err := rows.Scan(&b.ID, &b.ProductID, &b.BidderID, &amount, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.Amount, err = model.ParseMoney(amount)
		if err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (r *MySQL) Sell(ctx context.Context, productID, seller string, now time.Time) (model.SaleResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.SaleResult{}, err
	}
	defer tx.Rollback()
	p, err := lockedProduct(ctx, tx, productID)
	if err != nil {
		return model.SaleResult{}, err
	}
	if p.SellerID != seller {
		return model.SaleResult{}, ErrForbidden
	}
	if now.Before(p.AuctionEndAt) || (p.Status != "ACTIVE" && p.Status != "EXPIRED") || p.HighestBidderID == nil {
		return model.SaleResult{}, ErrConflict
	}
	id := uuid.NewString()
	soldAt := now.UTC()
	if _, err = tx.ExecContext(ctx, `INSERT INTO sales (id,product_id,seller_id,buyer_id,final_price,sold_at) VALUES (?,?,?,?,?,?)`, id, productID, seller, *p.HighestBidderID, p.CurrentPrice.String(), soldAt); err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return model.SaleResult{}, ErrConflict
		}
		return model.SaleResult{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE products SET status='SOLD' WHERE id=?`, productID); err != nil {
		return model.SaleResult{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.SaleResult{}, err
	}
	return model.SaleResult{ProductID: productID, Status: "SOLD", BuyerID: *p.HighestBidderID, FinalPrice: p.CurrentPrice, SoldAt: soldAt}, nil
}
func (r *MySQL) Expire(ctx context.Context, now time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE products SET status='EXPIRED' WHERE status='ACTIVE' AND auction_end_at<=?`, now.UTC())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func lockedProduct(ctx context.Context, tx *sql.Tx, id string) (model.Product, error) {
	return scanProduct(tx.QueryRowContext(ctx, `SELECT `+productColumns+` FROM products WHERE id=? FOR UPDATE`, id))
}

type scanner interface{ Scan(...any) error }

func scanProduct(row scanner) (model.Product, error) {
	var p model.Product
	var start, current string
	var highest sql.NullString
	err := row.Scan(&p.ID, &p.SellerID, &p.Title, &p.Description, &p.ImageKey, &p.Currency, &start, &current, &highest, &p.AuctionEndAt, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	if err != nil {
		return p, err
	}
	p.StartingPrice, err = model.ParseMoney(start)
	if err != nil {
		return p, err
	}
	p.CurrentPrice, err = model.ParseMoney(current)
	if highest.Valid {
		p.HighestBidderID = &highest.String
	}
	return p, err
}
func scanProducts(rows *sql.Rows) ([]model.Product, error) {
	result := []model.Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}
