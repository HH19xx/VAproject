package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"server/main/app/usecases"
)

// ユーザー関連エンドポイントに必要な依存をまとめた構造体。
type UserHandler struct {
	Usecase *usecases.UserUsecase
}

// 新規ユーザー登録用のHTTPハンドラ。
func (h *UserHandler) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("RegisterHandler: リクエストのデコード失敗: %v", err)
			JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "入力エラー", nil)
			return
		}

		log.Printf("RegisterHandler: 登録試行 - name=%s", req.Name)

		ctx := r.Context()
		userID, err := h.Usecase.Register(ctx, req.Name, req.Password)
		if err != nil {
			log.Printf("RegisterHandler: 登録失敗 - name=%s, error=%v", req.Name, err)
			JSONError(w, http.StatusBadRequest, "REGISTRATION_ERROR", err.Error(), nil)
			return
		}

		log.Printf("RegisterHandler: 登録成功 - userID=%d", userID)

		// 登録後、自動的にログイン処理を行いトークンを発行します。
		tokens, err := h.Usecase.IssueTokens(ctx, userID, "auth")
		if err != nil {
			log.Printf("RegisterHandler: トークン生成失敗 - userID=%d, error=%v", userID, err)
			JSONError(w, http.StatusInternalServerError, "TOKEN_ERROR", "トークン生成失敗", nil)
			return
		}

		log.Printf("RegisterHandler: 登録完了 - userID=%d", userID)

		JSONSuccess(w, http.StatusCreated, map[string]interface{}{
			"user_id":                 userID,
			"access_token":            tokens.AccessToken,
			"refresh_token":           tokens.RefreshToken,
			"access_token_expires_at": tokens.AccessTokenExpiresAt.UTC().Format(time.RFC3339),
		}, "ユーザー登録が完了しました")
	}
}

// ログイン用のHTTPハンドラ。
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

// ログイン済みユーザーの情報を返すエンドポイント
func (h *UserHandler) MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"id": userID,
		}, "ユーザー情報")
	}
}

// 認証済みユーザーのプロフィール情報を取得するハンドラ
func (h *UserHandler) GetProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 認証ミドルウェアからユーザーIDを取得する
		userID := r.Context().Value("userID").(int)

		log.Printf("GetProfileHandler: プロフィール取得試行 - userID=%d", userID)

		ctx := r.Context()
		profile, err := h.Usecase.GetProfile(ctx, userID)
		if err != nil {
			log.Printf("GetProfileHandler: プロフィール取得失敗 - userID=%d, error=%v", userID, err)
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "ユーザーが見つかりません", nil)
			return
		}

		log.Printf("GetProfileHandler: プロフィール取得成功 - userID=%d, name=%s", userID, profile.User.Name)

		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"id":             profile.User.ID,
			"name":           profile.User.Name,
			"email":          profile.User.Email,
			"auth_providers": profile.AuthProviders,
		}, "プロフィール情報を取得しました")
	}
}

// 認証済みユーザーのプロフィール情報を更新するハンドラ
func (h *UserHandler) UpdateProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 認証ミドルウェアからユーザーIDを取得する
		userID := r.Context().Value("userID").(int)

		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("UpdateProfileHandler: リクエストのデコード失敗: %v", err)
			JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "入力エラー", nil)
			return
		}

		log.Printf("UpdateProfileHandler: プロフィール更新試行 - userID=%d, newName=%s", userID, req.Name)

		ctx := r.Context()
		if err := h.Usecase.UpdateProfile(ctx, userID, req.Name); err != nil {
			log.Printf("UpdateProfileHandler: プロフィール更新失敗 - userID=%d, error=%v", userID, err)
			JSONError(w, http.StatusBadRequest, "UPDATE_ERROR", err.Error(), nil)
			return
		}

		log.Printf("UpdateProfileHandler: プロフィール更新成功 - userID=%d", userID)

		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"user_id": userID,
		}, "プロフィールを更新しました")
	}
}

// リフレッシュトークンを受け取りアクセストークンを再発行するハンドラ
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
