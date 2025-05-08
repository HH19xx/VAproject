package main

import (
	"log"
	"net/http"

	"server/main/app/usecases"
	"server/main/infra/db"
	"server/main/infra/repository"
	"server/main/interface/handler"
	"server/main/interface/middleware"
)

// CORSミドルウェア（暫定）
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// JWT認証ミドルウェア。除外パスを設定し、JWT認証を適用する。
func jwtProtectedMux(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 除外パスの設定
		publicPaths := []string{
			"/login",
			"/auth/google/login",
			"/auth/google/callback",
		}
		for _, path := range publicPaths {
			if r.URL.Path == path {
				mux.ServeHTTP(w, r)
				return
			}
		}

		// 認証ミドルウェアを適用
		middleware.JWTMiddleware(mux).ServeHTTP(w, r)
	})
}

func main() {
	// データベース接続
	conn, err := db.NewPostgresConnection()
	if err != nil {
		log.Fatalf("データベース接続失敗: %v", err)
	}
	defer conn.Close()

	// マイグレーション実行
	if err := db.RunMigrations(conn); err != nil {
		log.Fatalf("マイグレーション失敗: %v", err)
	}

	// ユースケースおよびハンドラの構築
	userRepo := repository.NewUserRepository(conn)
	userUsecase := &usecases.UserUsecase{Repo: userRepo}
	userHandler := &handler.UserHandler{Usecase: userUsecase}

	// HTTPルーティング設定
	mux := http.NewServeMux()
	mux.HandleFunc("/login", userHandler.LoginHandler())
	mux.HandleFunc("/me", userHandler.MeHandler())
	mux.HandleFunc("/message", handler.Handler)

	// Google OAuthのハンドラを設定
	oauthHandler := &handler.OAuthHandler{Usecase: userUsecase}
	mux.HandleFunc("/auth/google/login", oauthHandler.GoogleLoginHandler())
	mux.HandleFunc("/auth/google/callback", oauthHandler.GoogleCallbackHandler())

	// JWT認証とCORSミドルウェアを順に適用
	handlerWithAuth := jwtProtectedMux(mux)
	// CORSミドルウェアを適用
	handlerWithCORS := corsMiddleware(handlerWithAuth)

	log.Println("サーバーがポート8080で起動しました")
	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		log.Fatalf("HTTPサーバー起動失敗: %v", err)
	}
}
