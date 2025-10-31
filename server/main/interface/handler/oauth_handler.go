package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"server/main/app/usecases"
	"server/main/infra/oauth"

	"google.golang.org/api/oauth2/v2"
)

// OAuthHandler はGoogle OAuthログインを扱います。
type OAuthHandler struct {
	Usecase *usecases.UserUsecase
}

// Googleログイン画面にリダイレクトするハンドラ
func (h *OAuthHandler) GoogleLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config := oauth.GetGoogleOAuthConfig()
		log.Printf("OAuth設定 - ClientID: %s, RedirectURL: %s", config.ClientID, config.RedirectURL)
		// CSRF対策: 認可リクエストごとにstateを暗号学的乱数で生成し、
		// SameSite=LaxなHTTPOnlyクッキーに10分間だけ保持します。
		// これによりコールバックで一致検証ができ、不正リクエストを拒否できます。
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			JSONError(w, http.StatusInternalServerError, "OAUTH_STATE_ERROR", "state生成に失敗しました", nil)
			return
		}
		state := base64.RawURLEncoding.EncodeToString(b)
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Secure:   false, // 本番ではtrue
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(10 * time.Minute),
		})

		// 生成したstateを付与してGoogleの認可URLを作成します。
		url := config.AuthCodeURL(state)
		log.Printf("リダイレクトURL: %s", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// Google OAuthのコールバックを処理するハンドラ
func (h *OAuthHandler) GoogleCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "codeパラメータが存在しません", nil)
			return
		}
		// state検証: クエリのstateとCookieのstateを比較し不一致なら拒否します。
		state := r.URL.Query().Get("state")
		c, err := r.Cookie("oauth_state")
		if err != nil || c.Value == "" || c.Value != state {
			JSONError(w, http.StatusBadRequest, "INVALID_STATE", "stateが不正です", nil)
			return
		}

		config := oauth.GetGoogleOAuthConfig()

		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("トークン取得エラー: %v", err)
			JSONError(w, http.StatusBadRequest, "TOKEN_EXCHANGE_ERROR", "トークンの取得に失敗しました", nil)
			return
		}

		client := config.Client(context.Background(), token)
		srv, err := oauth2.New(client)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "OAUTH_CLIENT_ERROR", "ユーザー情報サービスの初期化に失敗しました", nil)
			return
		}

		userinfo, err := srv.Userinfo.Get().Do()
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "USERINFO_ERROR", "ユーザー情報の取得に失敗しました", nil)
			return
		}

		log.Printf("Googleユーザー: %s", userinfo.Email)

		userID, err := h.Usecase.FindOrCreateUserByEmail(r.Context(), userinfo.Email)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "USER_UPSERT_ERROR", "ユーザーの認証・登録に失敗しました", nil)
			return
		}

		tokens, err := h.Usecase.IssueTokens(r.Context(), userID, "oauth")
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "TOKEN_ERROR", "トークン生成に失敗しました", nil)
			return
		}

		// stateクッキーは利用済みなので削除し再利用を防ぎます。
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Unix(0, 0),
		})

		// フロントにアクセストークンとリフレッシュトークンを渡します。
		query := url.Values{}
		query.Set("token", tokens.AccessToken)
		query.Set("refresh", tokens.RefreshToken)
		query.Set("expires", tokens.AccessTokenExpiresAt.UTC().Format(time.RFC3339))
		frontendURL := fmt.Sprintf("http://localhost:3000/?%s", query.Encode())
		http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
	}
}
