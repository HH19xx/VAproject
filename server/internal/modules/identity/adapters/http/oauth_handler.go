package httpadapter

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	identityapp "server/internal/modules/identity/app"
	oauthadapter "server/internal/modules/identity/adapters/oauth"
	sharedhttp "server/internal/shared/http"

	"google.golang.org/api/oauth2/v2"
)

type OAuthHandler struct {
	Usecase *identityapp.UserService
}

func (h *OAuthHandler) GoogleLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config := oauthadapter.GetGoogleOAuthConfig()

		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "OAUTH_STATE_ERROR", "failed to create oauth state", nil)
			return
		}
		state := base64.RawURLEncoding.EncodeToString(b)

		cookie := &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(10 * time.Minute),
			Secure:   requestUsesHTTPS(r),
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, config.AuthCodeURL(state), http.StatusTemporaryRedirect)
	}
}

func (h *OAuthHandler) GoogleCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code == "" {
			sharedhttp.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "code parameter is required", nil)
			return
		}

		c, err := r.Cookie("oauth_state")
		if err != nil || c.Value == "" || c.Value != state {
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_STATE", "state is invalid", nil)
			return
		}

		config := oauthadapter.GetGoogleOAuthConfig()
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "TOKEN_EXCHANGE_ERROR", "failed to exchange oauth token", nil)
			return
		}

		client := config.Client(context.Background(), token)
		srv, err := oauth2.New(client)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "OAUTH_CLIENT_ERROR", "failed to create oauth client", nil)
			return
		}

		userinfo, err := srv.Userinfo.Get().Do()
		if err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "USERINFO_ERROR", "failed to fetch user info", nil)
			return
		}

		userID, err := h.Usecase.FindOrCreateUserByEmail(r.Context(), userinfo.Email)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "USER_UPSERT_ERROR", "failed to resolve user", nil)
			return
		}

		tokens, err := h.Usecase.IssueTokens(r.Context(), userID, "oauth")
		if err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "TOKEN_ERROR", "failed to issue tokens", nil)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Unix(0, 0),
		})

		query := url.Values{}
		query.Set("token", tokens.AccessToken)
		query.Set("refresh", tokens.RefreshToken)
		query.Set("expires", tokens.AccessTokenExpiresAt.UTC().Format(time.RFC3339))
		frontendURL := fmt.Sprintf("http://localhost:5173/?%s", query.Encode())
		http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
	}
}

func requestUsesHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}

	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
