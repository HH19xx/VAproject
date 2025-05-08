package handler

import (
	"encoding/json"
	"net/http"
	"server/main/app/usecases"
	"server/main/infra/jwt"
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
			http.Error(w, "入力エラー", http.StatusBadRequest)
			return
		}

		userID, err := h.Usecase.Login(r.Context(), req.Name, req.Password)
		if err != nil {
			http.Error(w, "認証失敗", http.StatusUnauthorized)
			return
		}

		token, err := jwt.GenerateJWT(userID)
		if err != nil {
			http.Error(w, "トークン生成失敗", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

// MeHandler はログイン済みユーザーの情報を返すエンドポイントです。
func (h *UserHandler) MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "ログイン中のユーザー情報",
			"user_id": userID,
		})
	}
}
