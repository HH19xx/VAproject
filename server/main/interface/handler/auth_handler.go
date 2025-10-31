package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"server/main/app/usecases"
)

// UserHandler はユーザー関連エンドポイントに必要な依存をまとめた構造体です。
type UserHandler struct {
	Usecase *usecases.UserUsecase
}

// LoginHandler はログイン用のHTTPハンドラです。
func (h *UserHandler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("LoginHandler: リクエストのデコード失敗: %v", err)
			JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "入力エラー", nil)
			return
		}

		log.Printf("LoginHandler: ログイン試行 - name=%s", req.Name)

		ctx := r.Context()
		userID, err := h.Usecase.Login(ctx, req.Name, req.Password)
		if err != nil {
			log.Printf("LoginHandler: 認証失敗 - name=%s, error=%v", req.Name, err)
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証失敗", nil)
			return
		}

		log.Printf("LoginHandler: 認証成功 - userID=%d", userID)

		tokens, err := h.Usecase.IssueTokens(ctx, userID, "auth")
		if err != nil {
			log.Printf("LoginHandler: トークン生成失敗 - userID=%d, error=%v", userID, err)
			JSONError(w, http.StatusInternalServerError, "TOKEN_ERROR", "トークン生成失敗", nil)
			return
		}

		log.Printf("LoginHandler: ログイン完了 - userID=%d", userID)

		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"access_token":            tokens.AccessToken,
			"refresh_token":           tokens.RefreshToken,
			"access_token_expires_at": tokens.AccessTokenExpiresAt.UTC().Format(time.RFC3339),
		}, "ログインしました")
	}
}

// MeHandler はログイン済みユーザーの情報を返すエンドポイントです。
func (h *UserHandler) MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"id": userID,
		}, "ユーザー情報")
	}
}

// RefreshHandler はリフレッシュトークンを受け取りアクセストークンを再発行します。
func (h *UserHandler) RefreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
			JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "refresh_tokenが必要です", nil)
			return
		}

		tokens, err := h.Usecase.RefreshTokens(r.Context(), req.RefreshToken)
		if err != nil {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "リフレッシュトークンが不正です", nil)
			return
		}

		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"access_token":            tokens.AccessToken,
			"refresh_token":           tokens.RefreshToken,
			"access_token_expires_at": tokens.AccessTokenExpiresAt.UTC().Format(time.RFC3339),
		}, "アクセストークンを再発行しました")
	}
}
