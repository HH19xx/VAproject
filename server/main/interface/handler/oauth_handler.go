package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"server/main/app/usecases"
	"server/main/infra/jwt"
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
		url := oauth.GoogleOAuthConfig.AuthCodeURL("random_state")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// Google OAuthのコールバックを処理するハンドラ
func (h *OAuthHandler) GoogleCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "codeパラメータが存在しません", http.StatusBadRequest)
			return
		}

		token, err := oauth.GoogleOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "トークンの取得に失敗しました", http.StatusBadRequest)
			return
		}

		client := oauth.GoogleOAuthConfig.Client(context.Background(), token)
		srv, err := oauth2.New(client)
		if err != nil {
			http.Error(w, "ユーザー情報サービスの初期化に失敗しました", http.StatusInternalServerError)
			return
		}

		userinfo, err := srv.Userinfo.Get().Do()
		if err != nil {
			http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
			return
		}

		log.Printf("Googleユーザー: %s", userinfo.Email)

		userID, err := h.Usecase.FindOrCreateUserByEmail(r.Context(), userinfo.Email)
		if err != nil {
			http.Error(w, "ユーザーの認証・登録に失敗しました", http.StatusInternalServerError)
			return
		}

		jwtToken, err := jwt.GenerateJWT(userID)
		if err != nil {
			http.Error(w, "JWTトークン生成に失敗しました", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": jwtToken})
	}
}
