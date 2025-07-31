package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	// テスト用のHTTPリクエストを作成
	req, err := http.NewRequest("GET", "/message", nil)
	if err != nil {
		t.Fatal(err)
	}

	// レスポンスレコーダーを作成
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Handler)

	// ハンドラーを呼び出し
	handler.ServeHTTP(rr, req)

	// ステータスコードの確認
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Content-Typeの確認
	expected := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expected {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, expected)
	}

	// レスポンスボディの確認
	var response Message
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response body: %v", err)
	}

	expectedMessage := "こんにちは、みなさん"
	if response.Message != expectedMessage {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response.Message, expectedMessage)
	}
}