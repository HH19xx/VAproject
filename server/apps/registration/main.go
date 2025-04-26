package main

import (
	"net/http"

	"server/apps/registration/handler"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/message", handler.Handler)

	// CORSミドルウェアを適用
	wrappedMux := corsMiddleware(mux)

	if err := http.ListenAndServe(":8080", wrappedMux); err != nil {
		panic(err)
	}
}
