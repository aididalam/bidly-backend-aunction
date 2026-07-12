package token

import (
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"testing"
	"time"
)

func TestVerifyAuthToken(t *testing.T) {
	secret := strings.Repeat("s", 32)
	now := time.Now()
	claims := Claims{Email: "u@example.com", RegisteredClaims: jwt.RegisteredClaims{Subject: "id", Issuer: "auth-service", IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}}
	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	got, err := New(secret).Verify(value)
	if err != nil || got.Subject != "id" {
		t.Fatalf("verify: %+v %v", got, err)
	}
	if _, err = New("wrong-secret-wrong-secret-wrong-secret").Verify(value); err == nil {
		t.Fatal("accepted wrong secret")
	}
}
