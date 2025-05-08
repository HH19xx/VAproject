package middleware

import (
	"context"
	"net/http"
	"strings"

	utils "server/main/infra/jwt"
)

// JWTMiddleware はJWTトークンを検証するミドルウェアです。
// リクエストヘッダにAuthorizationが含まれていることを確認し、
// トークンを検証して、ユーザーIDをコンテキストに追加します。
func JWTMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "認証ヘッダが必要です", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ValidateJWT(tokenStr)
		if err != nil {
			http.Error(w, "無効なトークン", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
