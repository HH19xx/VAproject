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

// OAuthHandler はGoogle OAuthログインを扱う
type OAuthHandler struct {
	Usecase *usecases.UserUsecase
}

// Googleログイン画面にリダイレクト
func (h *OAuthHandler) GoogleLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GoogleLoginHandler呼び出し: Method=%s, URL=%s, Origin=%s", r.Method, r.URL.String(), r.Header.Get("Origin"))
		config := oauth.GetGoogleOAuthConfig()
		log.Printf("OAuth設定 - ClientID: %s, RedirectURL: %s", config.ClientID, config.RedirectURL)
		// CSRF対策: 認可リクエストごとにstateを暗号学的乱数で生成し
		// SameSite=LaxなHTTPOnlyクッキーに10分間だけ保持
		// コールバックで一致検証をし不正リクエストを拒否
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			JSONError(w, http.StatusInternalServerError, "OAUTH_STATE_ERROR", "state生成に失敗しました", nil)
			return
		}
		state := base64.RawURLEncoding.EncodeToString(b)

		// Cookie設定
		cookie := &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(10 * time.Minute),
			// Domainは設定しない（デフォルトで現在のホスト:ポートに設定される）
		}

		cookie.Secure = true
		cookie.SameSite = http.SameSiteLaxMode

		http.SetCookie(w, cookie)
		log.Printf("Cookie設定: Name=%s, Value=%s, Secure=%v, SameSite=%v, Path=%s",
			cookie.Name, cookie.Value, cookie.Secure, cookie.SameSite, cookie.Path)

		// 生成したstateを付与してGoogleの認可URLを作成
		url := config.AuthCodeURL(state)
		log.Printf("リダイレクトURL: %s", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// Google OAuthのコールバックを処理
func (h *OAuthHandler) GoogleCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		log.Printf("Callback受信: code=%s, state=%s, Cookieヘッダ=%s", code, state, r.Header.Get("Cookie"))

		if code == "" {
			JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "codeパラメータが存在しません", nil)
			return
		}
		// state検証: クエリのstateとCookieのstateを比較し不一致なら拒否
		c, err := r.Cookie("oauth_state")
		if err != nil {
			log.Printf("state cookie 取得失敗: err=%v, query_state=%s", err, state)
			JSONError(w, http.StatusBadRequest, "INVALID_STATE", "stateが不正です", nil)
			return
		}
		if c.Value == "" || c.Value != state {
			log.Printf("state 不一致: cookie=%s, query=%s", c.Value, state)
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

		// stateクッキーは利用済みなので削除し再利用を防ぐ
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Unix(0, 0),
		})

		// フロントにアクセストークンとリフレッシュトークンを渡す
		query := url.Values{}
		query.Set("token", tokens.AccessToken)
		query.Set("refresh", tokens.RefreshToken)
		query.Set("expires", tokens.AccessTokenExpiresAt.UTC().Format(time.RFC3339))
		frontendURL := fmt.Sprintf("http://localhost:5173/?%s", query.Encode())
		http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
	}
}
