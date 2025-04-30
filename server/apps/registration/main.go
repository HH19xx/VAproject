package main

import (
	"log"
	"net/http"

	"server/apps/registration/handler"
	"server/infra/db"
)

// CORSを無効化する
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// main関数で全体を制御する
func main() {
	// DBに接続する
	conn, err := db.NewPostgresConnection()
	if err != nil {
		log.Fatalf("データベース接続失敗: %v", err)
	}
	defer conn.Close()

	// マイグレーションを実行する
	if err := db.RunMigrations(conn); err != nil {
		log.Fatalf("マイグレーション失敗: %v", err)
	}

	// HTTPサーバを起動する
	mux := http.NewServeMux()
	mux.HandleFunc("/message", handler.Handler)

	wrappedMux := corsMiddleware(mux)

	log.Println("サーバーがポート8080で起動しました")

	if err := http.ListenAndServe(":8080", wrappedMux); err != nil {
		panic(err)
	}
}
