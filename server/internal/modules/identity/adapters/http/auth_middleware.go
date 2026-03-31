package httpadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	jwtadapter "server/internal/modules/identity/adapters/jwt"
)

func JWTMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtadapter.ValidateJWT(tokenStr)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		if claims.Username != "" {
			ctx = context.WithValue(ctx, "username", claims.Username)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
