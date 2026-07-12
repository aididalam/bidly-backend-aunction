package integration_test

import (
	"auction/auction/internal/handler"
	"auction/auction/internal/middleware"
	"auction/auction/internal/model"
	"auction/auction/internal/repository"
	"auction/auction/internal/service"
	"auction/auction/internal/token"
	"auction/auction/internal/upload"
	"auction/auction/migrations"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeUpload struct{}

func (fakeUpload) Generate(_ context.Context, user, file, kind string) (upload.Result, error) {
	if kind != "image/jpeg" {
		return upload.Result{}, fmt.Errorf("invalid")
	}
	return upload.Result{UploadURL: "https://upload", ImageKey: "products/" + user + "/x.jpg", ImageURL: "https://images/products/" + user + "/x.jpg"}, nil
}
func TestAuctionFlowMySQL(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	if err = migrations.Up(ctx, db); err != nil {
		t.Fatal(err)
	}
	secret := strings.Repeat("s", 32)
	svc := service.New(repository.NewMySQL(db), fakeUpload{})
	server := httptest.NewServer(handler.New(svc, middleware.New(token.New(secret))))
	defer server.Close()
	seller := uuid.NewString()
	bidder := uuid.NewString()
	other := uuid.NewString()
	sellerToken := issue(t, secret, seller)
	bidderToken := issue(t, secret, bidder)
	otherToken := issue(t, secret, other)
	var ids []string
	defer func() {
		for _, id := range ids {
			_, _ = db.Exec(`DELETE FROM sales WHERE product_id=?`, id)
			_, _ = db.Exec(`DELETE FROM products WHERE id=?`, id)
		}
	}()
	create := func(title string) productResponse {
		body := map[string]any{"title": title, "description": "Description", "image_key": "products/" + seller + "/image.jpg", "starting_price": 10.00, "auction_end_at": time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)}
		res := call(t, server.Client(), "POST", server.URL+"/api/products", body, sellerToken)
		if res.StatusCode != 201 {
			t.Fatalf("create %d: %s", res.StatusCode, res.body)
		}
		var p productResponse
		decode(t, res.body, &p)
		ids = append(ids, p.ID)
		return p
	}
	p := create("Product")
	if p.CurrentPrice != 1000 || p.Status != "ACTIVE" {
		t.Fatalf("product: %+v", p)
	}
	if res := call(t, server.Client(), "GET", server.URL+"/api/products", nil, ""); res.StatusCode != 200 {
		t.Fatalf("list %d", res.StatusCode)
	}
	update := map[string]any{"title": "Updated", "description": "Description", "image_key": "products/" + seller + "/image.jpg", "starting_price": 12.00, "auction_end_at": time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339Nano)}
	if res := call(t, server.Client(), "PUT", server.URL+"/api/products/"+p.ID, update, sellerToken); res.StatusCode != 200 {
		t.Fatalf("update %d: %s", res.StatusCode, res.body)
	}
	if res := call(t, server.Client(), "POST", server.URL+"/api/products/"+p.ID+"/bids", map[string]any{"amount": 20.00}, sellerToken); res.StatusCode != 403 {
		t.Fatalf("seller bid %d", res.StatusCode)
	}
	if res := call(t, server.Client(), "POST", server.URL+"/api/products/"+p.ID+"/bids", map[string]any{"amount": 20.00}, bidderToken); res.StatusCode != 201 {
		t.Fatalf("bid %d: %s", res.StatusCode, res.body)
	}
	if res := call(t, server.Client(), "PUT", server.URL+"/api/products/"+p.ID, update, sellerToken); res.StatusCode != 409 {
		t.Fatalf("update after bid %d", res.StatusCode)
	}
	if res := call(t, server.Client(), "GET", server.URL+"/api/products/"+p.ID+"/bids", nil, otherToken); res.StatusCode != 403 {
		t.Fatalf("foreign bids %d", res.StatusCode)
	}
	if _, err = db.Exec(`UPDATE products SET auction_end_at=? WHERE id=?`, time.Now().Add(-time.Minute), p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	if res := call(t, server.Client(), "POST", server.URL+"/api/products/"+p.ID+"/sell", nil, sellerToken); res.StatusCode != 200 {
		t.Fatalf("sell %d: %s", res.StatusCode, res.body)
	}
	if res := call(t, server.Client(), "POST", server.URL+"/api/products/"+p.ID+"/sell", nil, sellerToken); res.StatusCode != 409 {
		t.Fatalf("second sell %d", res.StatusCode)
	}
	cancelProduct := create("Cancel")
	if res := call(t, server.Client(), "DELETE", server.URL+"/api/products/"+cancelProduct.ID, nil, sellerToken); res.StatusCode != 204 {
		t.Fatalf("cancel %d: %s", res.StatusCode, res.body)
	}
	concurrent := create("Concurrent")
	amounts := []float64{30, 40}
	codes := make([]int, 2)
	var wg sync.WaitGroup
	for i := range amounts {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			codes[i] = call(t, server.Client(), "POST", server.URL+"/api/products/"+concurrent.ID+"/bids", map[string]any{"amount": amounts[i]}, []string{bidderToken, otherToken}[i]).StatusCode
		}(i)
	}
	wg.Wait()
	for _, code := range codes {
		if code != 201 && code != 409 {
			t.Fatalf("concurrent codes: %v", codes)
		}
	}
	var current productResponse
	res := call(t, server.Client(), "GET", server.URL+"/api/products/"+concurrent.ID, nil, "")
	decode(t, res.body, &current)
	if current.CurrentPrice != 4000 {
		t.Fatalf("current price %d, codes %v", current.CurrentPrice, codes)
	}
	if res := call(t, server.Client(), "POST", server.URL+"/api/uploads/presigned-url", map[string]any{"filename": "x.jpg", "content_type": "image/jpeg"}, sellerToken); res.StatusCode != 200 {
		t.Fatalf("upload %d: %s", res.StatusCode, res.body)
	}
}

type response struct {
	StatusCode int
	body       []byte
}
type productResponse struct {
	ID           string      `json:"id"`
	CurrentPrice model.Money `json:"current_price"`
	Status       string      `json:"status"`
}

func call(t *testing.T, c *http.Client, method, url string, body any, bearer string) response {
	t.Helper()
	var b bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&b).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, url, &b)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	res, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out bytes.Buffer
	_, _ = out.ReadFrom(res.Body)
	return response{res.StatusCode, out.Bytes()}
}
func decode(t *testing.T, data []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("decode %s: %v", data, err)
	}
}
func issue(t *testing.T, secret, user string) string {
	t.Helper()
	now := time.Now()
	claims := token.Claims{Email: user + "@example.com", RegisteredClaims: jwt.RegisteredClaims{Subject: user, Issuer: "auth-service", IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}}
	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return value
}
