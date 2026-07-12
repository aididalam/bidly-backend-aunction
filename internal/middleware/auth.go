package middleware

import (
	"auction/auction/internal/token"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type key int

const claimsKey key = 0

type Auth struct{ verifier token.Verifier }

func New(verifier token.Verifier) *Auth { return &Auth{verifier: verifier} }
func (a *Auth) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(w)
			return
		}
		claims, err := a.verifier.Verify(parts[1])
		if err != nil {
			unauthorized(w)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey, claims)))
	})
}
func Claims(ctx context.Context) (token.Claims, bool) {
	v, ok := ctx.Value(claimsKey).(token.Claims)
	return v, ok
}
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": "unauthorized", "message": "authentication is required"}})
}
