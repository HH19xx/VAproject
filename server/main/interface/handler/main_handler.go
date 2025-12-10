package handler

import (
	"encoding/json"
	"net/http"
)

// レスポンスをJSONの構造体、Messageとして定義する
type Message struct {
	Message string `json:"message"`
}

// messageエンドポイントの処理を行う
func Handler(w http.ResponseWriter, r *http.Request) {
	response := Message{Message: "こんにちは、みなさん"}

	// レスポンスのContent-TypeヘッダをJSONとして設定
	w.Header().Set("Content-Type", "application/json")

	// レスポンスのステータスコードを設定
	w.WriteHeader(http.StatusOK)

	// JSONのエンコード処理に失敗した場合、エラーメッセージを返す
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
