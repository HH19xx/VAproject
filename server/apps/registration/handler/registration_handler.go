package handler

import (
	"encoding/json"
	"net/http"
	"server/core/domain"
	"server/core/usecases"
)

// RegistrationHandler は登録関連のHTTPハンドラです。
type RegistrationHandler struct {
	userUseCase *usecases.UserUseCase
}

// NewRegistrationHandler は新しいRegistrationHandlerのインスタンスを作成します。
func NewRegistrationHandler(userUseCase *usecases.UserUseCase) *RegistrationHandler {
	return &RegistrationHandler{userUseCase: userUseCase}
}

// RegisterUserHandler はユーザー登録のリクエストを処理します。
func (h *RegistrationHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "許可されていないメソッドです", http.StatusMethodNotAllowed)
		return
	}

	var user domain.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "リクエストボディの解析に失敗しました: "+err.Error(), http.StatusBadRequest)
		return
	}

	// ユースケースを呼び出してユーザーを登録します。
	err = h.userUseCase.RegisterUser(&user)
	if err != nil {
		http.Error(w, "ユーザー登録に失敗しました: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "ユーザーが正常に登録されました"})
}
