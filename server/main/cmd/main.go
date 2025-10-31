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
		// 除外パスの設定: 未認証でもアクセス可能なエンドポイントを列挙します。
		// 互換性維持のため旧パスも残しつつ、新API仕様に沿った /api/v1 配下も許可します。
		publicPaths := []string{
			"/",
			"/message",
			// legacy
			"/login",
			"/auth/oauth/google/login",
			"/auth/oauth/google/callback",
			// v1
			"/api/v1/auth/login",
			"/api/v1/auth/refresh",
			"/api/v1/auth/oauth/google/login",
			"/api/v1/auth/oauth/google/callback",
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

	// テストデータのマイグレーション実行（test_dataディレクトリが存在する場合のみ）
	if err := db.RunTestDataMigrations(conn); err != nil {
		log.Fatalf("テストデータのマイグレーション失敗: %v", err)
	}

	// ユースケースおよびハンドラの構築
	userRepo := repository.NewUserRepository(conn)
	refreshRepo := repository.NewRefreshTokenRepository(conn)
	userUsecase := &usecases.UserUsecase{Repo: userRepo, RefreshRepo: refreshRepo}
	userHandler := &handler.UserHandler{Usecase: userUsecase}

	// HTTPルーティング設定
	mux := http.NewServeMux()

	// ルートパス用のシンプルなハンドラ
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "VAproject API Server", "oauth_login": "/auth/oauth/google/login"}`))
	})
	// v1 auth endpoints
	mux.HandleFunc("/api/v1/auth/login", userHandler.LoginHandler())
	mux.HandleFunc("/api/v1/auth/me", userHandler.MeHandler())
	mux.HandleFunc("/api/v1/auth/refresh", userHandler.RefreshHandler())
	mux.HandleFunc("/message", handler.Handler)

	// Google OAuthのハンドラを設定
	oauthHandler := &handler.OAuthHandler{Usecase: userUsecase}
	mux.HandleFunc("/api/v1/auth/oauth/google/login", oauthHandler.GoogleLoginHandler())
	mux.HandleFunc("/api/v1/auth/oauth/google/callback", oauthHandler.GoogleCallbackHandler())

	// backward-compatible legacy routes (temporary)
	mux.HandleFunc("/login", userHandler.LoginHandler())
	mux.HandleFunc("/me", userHandler.MeHandler())
	mux.HandleFunc("/auth/oauth/google/login", oauthHandler.GoogleLoginHandler())
	mux.HandleFunc("/auth/oauth/google/callback", oauthHandler.GoogleCallbackHandler())

	// JWT認証とCORSミドルウェアを順に適用
	handlerWithAuth := jwtProtectedMux(mux)
	// CORSミドルウェアを適用
	handlerWithCORS := corsMiddleware(handlerWithAuth)

	log.Println("サーバーがポート8080で起動しました")
	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		log.Fatalf("HTTPサーバー起動失敗: %v", err)
	}
}
