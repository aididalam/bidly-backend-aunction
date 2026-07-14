package demo

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"
)

//go:embed assets/*.jpg
var assets embed.FS

type ObjectStore interface {
	PutDemoObject(context.Context, string, []byte) error
}

type productSeed struct {
	id, sellerID, title, description, asset, currency, startingPrice, currentPrice, highestBidderID string
	days                                                                                            int
}

type bidSeed struct{ id, productID, bidderID, amount string }

const (
	amina  = "11111111-1111-4111-8111-111111111111"
	nabil  = "22222222-2222-4222-8222-222222222222"
	farah  = "33333333-3333-4333-8333-333333333333"
	tariq  = "44444444-4444-4444-8444-444444444444"
	samira = "55555555-5555-4555-8555-555555555555"
)

var products = []productSeed{
	{"a1000000-0000-4000-8000-000000000001", amina, "Vintage brass desk lamp", "A warm brass desk lamp with a classic silhouette and an inviting pool of light. A handsome piece for a study, bedside table, or reading corner.", "listing-01.jpg", "USD", "280.00", "340.00", farah, 1},
	{"a1000000-0000-4000-8000-000000000002", nabil, "Monochrome retro desk lamp", "A sculptural retro lamp in a striking monochrome finish. Its focused shade makes a quiet desk or shelf feel instantly more considered.", "listing-02.jpg", "BDT", "1600.00", "2050.00", tariq, 2},
	{"a1000000-0000-4000-8000-000000000003", farah, "Mid-century reading lamp", "A refined reading lamp with a softly diffused glow and clean mid-century lines. The compact footprint works beautifully on a writing desk.", "listing-03.jpg", "GBP", "39.00", "46.50", amina, 3},
	{"a1000000-0000-4000-8000-000000000004", tariq, "Industrial Edison desk lamp", "An industrial-style desk lamp featuring an exposed Edison bulb and aged metal finish. It brings warmth and character to a workspace.", "listing-04.jpg", "BDT", "2200.00", "2800.00", samira, 4},
	{"a1000000-0000-4000-8000-000000000005", samira, "Teak workspace chair", "A wooden workspace chair with tactile natural grain and a relaxed, architectural profile. It is equally at home beside a desk or dining table.", "listing-05.jpg", "EUR", "125.00", "170.00", nabil, 5},
	{"a1000000-0000-4000-8000-000000000006", amina, "Carved wooden armchair", "A handsome carved wooden armchair with an ornate frame and quietly traditional proportions. A distinctive occasional chair with real presence.", "listing-06.jpg", "USD", "180.00", "180.00", "", 6},
	{"a1000000-0000-4000-8000-000000000007", nabil, "Botanical art display", "A framed botanical artwork with soft, natural tones and a handmade feel. It offers an easy way to bring greenery and calm to a wall.", "listing-07.jpg", "AUD", "95.00", "95.00", "", 7},
	{"a1000000-0000-4000-8000-000000000008", farah, "Framed botanical study", "A minimalist botanical study in a slim frame. The dried leaf composition and textured paper make it a subtle, collected wall piece.", "listing-08.jpg", "INR", "4200.00", "4200.00", "", 8},
	{"a1000000-0000-4000-8000-000000000009", tariq, "Dried flower wall frame", "A delicate arrangement of dried flowers, neatly framed against a warm neutral backdrop. Ideal for a quiet bedroom, hall, or reading nook.", "listing-09.jpg", "CAD", "130.00", "130.00", "", 9},
	{"a1000000-0000-4000-8000-000000000010", samira, "Framed plant print", "A plant study in a simple frame with gentle natural texture. It layers beautifully with other small artworks or works alone above a desk.", "listing-10.jpg", "USD", "310.00", "310.00", "", 10},
	{"a1000000-0000-4000-8000-000000000011", amina, "Wooden botanical frame", "A warm wood frame paired with a dried palm leaf arrangement. It brings a restrained natural accent to a wall or shelf.", "listing-11.jpg", "EUR", "260.00", "260.00", "", 11},
	{"a1000000-0000-4000-8000-000000000012", nabil, "Walnut cheese serving board", "A generous wooden serving board styled for cheese, fruit, and relaxed gatherings. The natural grain makes every table setting feel more inviting.", "listing-12.jpg", "GBP", "240.00", "240.00", "", 12},
	{"a1000000-0000-4000-8000-000000000013", farah, "Walnut snack bowl", "A simple walnut bowl with a warm natural finish, perfect for fruit, nuts, or small everyday objects. A useful piece of understated tabletop craft.", "listing-13.jpg", "USD", "110.00", "110.00", "", 13},
	{"a1000000-0000-4000-8000-000000000014", tariq, "Brass candle snuffer", "An elegant brass candle snuffer with a slender handle and traditional silhouette. A small, ritualistic object for anyone who loves candlelight.", "listing-14.jpg", "JPY", "15000.00", "15000.00", "", 14},
	{"a1000000-0000-4000-8000-000000000015", samira, "Metal candleholder", "A refined metal candleholder with a softly aged finish. It holds a dinner candle securely and creates a warm focal point on a table or mantel.", "listing-15.jpg", "USD", "85.00", "85.00", "", 15},
	{"a1000000-0000-4000-8000-000000000016", amina, "Leather travel notebook", "A leather-bound notebook with a tactile cover and generous pages for notes, sketches, and small plans. It is made to become more personal with use.", "listing-16.jpg", "EUR", "360.00", "360.00", "", 16},
	{"a1000000-0000-4000-8000-000000000017", nabil, "Leather-bound sketchbook", "A compact leather notebook on a warm wood surface, ready for ideas, lists, and field notes. Its simple cover and sturdy binding are made for daily use.", "listing-17.jpg", "BDT", "1750.00", "1750.00", "", 17},
	{"a1000000-0000-4000-8000-000000000018", farah, "Textured ceramic vase", "A tall handmade ceramic vase with tactile surface detail and a quiet sculptural shape. It stands beautifully alone or filled with a few branches.", "listing-18.jpg", "CAD", "120.00", "120.00", "", 18},
	{"a1000000-0000-4000-8000-000000000019", tariq, "Black iron candlesticks", "A set of black iron candlesticks with classic proportions and a strong silhouette. They add warmth and a little ceremony to a meal or mantel.", "listing-19.jpg", "AUD", "145.00", "145.00", "", 19},
	{"a1000000-0000-4000-8000-000000000020", samira, "Minimalist candleholder pair", "A clean pair of candleholders with balanced proportions and warm, quiet presence. They work beautifully together on a cabinet, table, or shelf.", "listing-20.jpg", "USD", "210.00", "210.00", "", 20},
}

var bids = []bidSeed{
	{"b1000000-0000-4000-8000-000000000001", "a1000000-0000-4000-8000-000000000001", nabil, "310.00"},
	{"b1000000-0000-4000-8000-000000000002", "a1000000-0000-4000-8000-000000000001", farah, "340.00"},
	{"b1000000-0000-4000-8000-000000000003", "a1000000-0000-4000-8000-000000000002", farah, "1850.00"},
	{"b1000000-0000-4000-8000-000000000004", "a1000000-0000-4000-8000-000000000002", tariq, "2050.00"},
	{"b1000000-0000-4000-8000-000000000005", "a1000000-0000-4000-8000-000000000003", samira, "43.00"},
	{"b1000000-0000-4000-8000-000000000006", "a1000000-0000-4000-8000-000000000003", amina, "46.50"},
	{"b1000000-0000-4000-8000-000000000007", "a1000000-0000-4000-8000-000000000004", amina, "2500.00"},
	{"b1000000-0000-4000-8000-000000000008", "a1000000-0000-4000-8000-000000000004", samira, "2800.00"},
	{"b1000000-0000-4000-8000-000000000009", "a1000000-0000-4000-8000-000000000005", farah, "150.00"},
	{"b1000000-0000-4000-8000-000000000010", "a1000000-0000-4000-8000-000000000005", nabil, "170.00"},
}

// Seed uploads the versioned 16:9 catalogue images and inserts their matching
// listings. Every write is idempotent so restarts never replace real data.
func Seed(ctx context.Context, db *sql.DB, store ObjectStore) error {
	for _, product := range products {
		image, err := assets.ReadFile("assets/" + product.asset)
		if err != nil {
			return fmt.Errorf("read demo asset %s: %w", product.asset, err)
		}
		if err := store.PutDemoObject(ctx, imageKey(product), image); err != nil {
			return fmt.Errorf("upload demo asset %s: %w", product.asset, err)
		}
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	for _, product := range products {
		_, err = tx.ExecContext(ctx, `INSERT IGNORE INTO products (id,seller_id,title,description,image_key,currency,starting_price,current_price,highest_bidder_id,auction_end_at,status) VALUES (?,?,?,?,?,?,?,?,?,?, 'ACTIVE')`, product.id, product.sellerID, product.title, product.description, imageKey(product), product.currency, product.startingPrice, product.currentPrice, nullable(product.highestBidderID), now.AddDate(0, 0, product.days))
		if err != nil {
			return fmt.Errorf("insert demo product %s: %w", product.id, err)
		}
		_, err = tx.ExecContext(ctx, `UPDATE products SET title=?,description=?,image_key=?,currency=?,starting_price=?,current_price=?,highest_bidder_id=?,auction_end_at=? WHERE id=? AND image_key=''`, product.title, product.description, imageKey(product), product.currency, product.startingPrice, product.currentPrice, nullable(product.highestBidderID), now.AddDate(0, 0, product.days), product.id)
		if err != nil {
			return fmt.Errorf("upgrade demo product %s: %w", product.id, err)
		}
	}
	for _, bid := range bids {
		if _, err = tx.ExecContext(ctx, `INSERT IGNORE INTO bids (id,product_id,bidder_id,amount) VALUES (?,?,?,?)`, bid.id, bid.productID, bid.bidderID, bid.amount); err != nil {
			return fmt.Errorf("insert demo bid %s: %w", bid.id, err)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE bids SET amount=? WHERE id=?`, bid.amount, bid.id); err != nil {
			return fmt.Errorf("upgrade demo bid %s: %w", bid.id, err)
		}
	}
	return tx.Commit()
}

func imageKey(product productSeed) string { return "products/demo/" + product.asset }

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
