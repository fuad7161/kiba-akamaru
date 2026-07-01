package middleware

import (
	"net/http"

	"github.com/fuad71/job-circular-api/pkg/response"
)

// AdminOnly checks that the authenticated user has role "admin".
// Must be used after AuthRequired middleware.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r)
		if claims == nil || claims.Role != "admin" {
			response.Error(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
