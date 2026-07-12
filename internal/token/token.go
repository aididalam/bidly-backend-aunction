package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
type Verifier interface{ Verify(string) (Claims, error) }
type Manager struct{ secret []byte }

func New(secret string) *Manager { return &Manager{secret: []byte(secret)} }
func (m *Manager) Verify(value string) (Claims, error) {
	claims := Claims{}
	parsed, err := jwt.ParseWithClaims(value, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing algorithm")
		}
		return m.secret, nil
	}, jwt.WithIssuer("auth-service"), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !parsed.Valid || claims.Subject == "" || claims.Email == "" {
		return Claims{}, errors.New("invalid token")
	}
	return claims, nil
}
