package main

import (
	"log"
	"net/http"

	"server/apps/registration/handler"
	"server/infra/db"
)

// CORS対応ミドルウェア
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

func main() {
	// DB接続
	conn, err := db.NewPostgresConnection()
	if err != nil {
		log.Fatalf("データベース接続失敗: %v", err)
	}
	defer conn.Close()

	// マイグレーション実行（RunMigrationsを作っている前提）
	if err := db.RunMigrations(conn); err != nil {
		log.Fatalf("マイグレーション失敗: %v", err)
	}

	// HTTPサーバ起動
	mux := http.NewServeMux()
	mux.HandleFunc("/message", handler.Handler)

	wrappedMux := corsMiddleware(mux)

	log.Println("サーバーがポート8080で起動しました")

	if err := http.ListenAndServe(":8080", wrappedMux); err != nil {
		panic(err)
	}
}
