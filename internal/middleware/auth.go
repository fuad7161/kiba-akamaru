package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/fuad71/job-circular-api/internal/service"
	"github.com/fuad71/job-circular-api/pkg/response"
)

type contextKey string

const ClaimsKey contextKey = "claims"

// AuthRequired validates JWT and injects claims into the request context.
func AuthRequired(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "missing or malformed token")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwt.ParseWithClaims(tokenStr, &service.Claims{},
				func(t *jwt.Token) (interface{}, error) {
					return []byte(authSvc.GetJWTSecret()), nil
				})
			if err != nil || !claims.Valid {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			claimsData, ok := claims.Claims.(*service.Claims)
			if !ok {
				response.Error(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claimsData)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims extracts JWT claims from the request context.
func GetClaims(r *http.Request) *service.Claims {
	claims, ok := r.Context().Value(ClaimsKey).(*service.Claims)
	if !ok {
		return nil
	}
	return claims
}
