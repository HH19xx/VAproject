package handler

import (
	"encoding/json"
	"net/http"
)

// MessageはレスポンスのJSONの構造体
type Message struct {
	Message string `json:"message"`
}

// Handlerは/messageエンドポイントの処理を行う
func Handler(w http.ResponseWriter, r *http.Request) {
	response := Message{Message: "こんにちは、みなさん"}

	// レスポンスのJSONを返す
	w.Header().Set("Content-Type", "application/json")

	// レスポンスのステータスコードを設定
	w.WriteHeader(http.StatusOK)

	// エラーが発生した場合はエラーメッセージを返す
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
