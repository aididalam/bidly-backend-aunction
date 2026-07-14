package service

import (
	"strings"
	"testing"
	"time"

	"auction/auction/internal/model"
)

func TestValidProductDescriptionLimit(t *testing.T) {
	seller := "11111111-1111-4111-8111-111111111111"
	input := ProductInput{
		Title:         "Desk lamp",
		Description:   strings.Repeat("a", maxDescriptionRunes),
		ImageKey:      "products/" + seller + "/image.jpg",
		Currency:      "usd",
		StartingPrice: model.Money(100),
		AuctionEndAt:  time.Now().Add(time.Hour),
	}
	if !validProduct(seller, &input, time.Now()) {
		t.Fatal("expected description at the limit to be valid")
	}
	input.Description += "a"
	if validProduct(seller, &input, time.Now()) {
		t.Fatal("expected description above the limit to be invalid")
	}
	input.Description = "Short description"
	input.Currency = "US"
	if validProduct(seller, &input, time.Now()) {
		t.Fatal("expected invalid currency to be rejected")
	}
}
