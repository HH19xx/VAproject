package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// このファイルではAPIレスポンスの統一フォーマットを提供
// - 成功時は success=true, data, message, timestamp を返す
// - 失敗時は success=false, error{code,message,details}, timestamp を返す
// これによりフロントエンドが安定したパースを行える
type successEnvelope struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type errorEnvelope struct {
	Success   bool        `json:"success"`
	Error     interface{} `json:"error"`
	Timestamp string      `json:"timestamp"`
}

type apiError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	// レスポンスヘッダにJSONであることを明示し指定のステータスコードでボディを書き込む
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func JSONSuccess(w http.ResponseWriter, status int, data interface{}, message string) {
	// 成功時の共通レスポンス
	writeJSON(w, status, successEnvelope{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func JSONError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	// 失敗時の共通レスポンス
	writeJSON(w, status, errorEnvelope{
		Success: false,
		Error: apiError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
