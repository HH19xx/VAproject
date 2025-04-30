package main

import (
	"log"
	"net/http"

	"server/apps/registration/handler"
	"server/infra/db"
)

// 簡易CORSミドルウェア（後にinterface/middlewareに移行予定）
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// main関数：全体の初期化と起動
func main() {
	conn, err := db.NewPostgresConnection()
	if err != nil {
		log.Fatalf("データベース接続失敗: %v", err)
	}
	defer conn.Close()

	if err := db.RunMigrations(conn); err != nil {
		log.Fatalf("マイグレーション失敗: %v", err)
	}

	// ルーティング設定（本来は apps/registration/routes/ に切り出す予定）
	mux := http.NewServeMux()
	mux.HandleFunc("/message", handler.Handler)

	// ミドルウェアを適用
	wrapped := corsMiddleware(mux)

	log.Println("サーバーがポート8080で起動しました")
	if err := http.ListenAndServe(":8080", wrapped); err != nil {
		log.Fatalf("HTTPサーバー起動失敗: %v", err)
	}
}
