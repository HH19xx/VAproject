package main

import (
	"log"
	"net/http"
	"strings"

	"server/main/infra/db"
	"server/main/infra/repository"
	"server/main/interface/handler"
	"server/main/interface/middleware"
	"server/main/usecases"
)

// CORSミドルウェア（暫定）
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// 開発環境では localhost からのリクエストを許可
		if origin == "http://localhost:5173" || origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true") // Cookie送信を許可
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// JWT認証ミドルウェア。除外パスを設定しJWT認証を適用
func jwtProtectedMux(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 除外パスの設定: 未認証でもアクセス可能なエンドポイントを列挙
		publicPaths := []string{
			"/",
			"/message",
			"/api/v1/auth/register",
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

	// テストデータのマイグレーション実行
	if err := db.RunTestDataMigrations(conn); err != nil {
		log.Fatalf("テストデータのマイグレーション失敗: %v", err)
	}

	// ユースケースおよびハンドラの構築
	userRepo := repository.NewUserRepository(conn)
	authProviderRepo := repository.NewUserAuthProviderRepository(conn)
	refreshRepo := repository.NewRefreshTokenRepository(conn)
	userUsecase := &usecases.UserUsecase{
		Repo:             userRepo,
		AuthProviderRepo: authProviderRepo,
		RefreshRepo:      refreshRepo,
	}
	userHandler := &handler.UserHandler{Usecase: userUsecase}

	// 観察対象のユースケースおよびハンドラを構築
	targetRepo := repository.NewTargetRepository(conn)
	targetUsecase := &usecases.TargetUsecase{Repo: targetRepo}
	targetHandler := &handler.TargetHandler{Usecase: targetUsecase}

	actionLogRepo := repository.NewActionLogRepository(conn)
	actionLogUsecase := &usecases.ActionLogUsecase{Repo: actionLogRepo}
	actionLogHandler := &handler.ActionLogHandler{Usecase: actionLogUsecase}

	actionTypeRepo := repository.NewActionTypeRepository(conn)
	actionTypeUsecase := &usecases.ActionTypeUsecase{Repo: actionTypeRepo}
	actionTypeHandler := &handler.ActionTypeHandler{Usecase: actionTypeUsecase}

	// 対象と行動種別の紐づけユースケースおよびハンドラを構築
	targetActionTypeRepo := repository.NewTargetActionTypeRepository(conn)
	targetActionTypeUsecase := &usecases.TargetActionTypeUsecase{Repo: targetActionTypeRepo}
	targetActionTypeHandler := &handler.TargetActionTypeHandler{Usecase: targetActionTypeUsecase}

	// HTTPルーティング設定
	mux := http.NewServeMux()

	// ルートパス用のシンプルなハンドラ
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "VAproject API Server", "oauth_login": "/api/v1/auth/oauth/google/login"}`))
	})
	// v1 auth endpoints
	mux.HandleFunc("/api/v1/auth/register", userHandler.RegisterHandler())
	mux.HandleFunc("/api/v1/auth/login", userHandler.LoginHandler())
	mux.HandleFunc("/api/v1/auth/me", userHandler.MeHandler())
	mux.HandleFunc("/api/v1/auth/refresh", userHandler.RefreshHandler())
	mux.HandleFunc("/api/v1/auth/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			userHandler.GetProfileHandler()(w, r)
		} else if r.Method == http.MethodPut {
			userHandler.UpdateProfileHandler()(w, r)
		} else {
			handler.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "このエンドポイントはGETまたはPUTメソッドのみ対応", nil)
		}
	})
	mux.HandleFunc("/message", handler.Handler)

	// Google OAuthのハンドラを設定
	oauthHandler := &handler.OAuthHandler{Usecase: userUsecase}
	mux.HandleFunc("/api/v1/auth/oauth/google/login", oauthHandler.GoogleLoginHandler())
	mux.HandleFunc("/api/v1/auth/oauth/google/callback", oauthHandler.GoogleCallbackHandler())

	// 観察対象のエンドポイントを設定（すべて認証必須）
	mux.HandleFunc("/api/v1/targets", func(w http.ResponseWriter, r *http.Request) {
		// リクエストメソッドで処理を分岐
		if r.Method == http.MethodGet {
			targetHandler.ListTargetsHandler()(w, r)
		} else if r.Method == http.MethodPost {
			targetHandler.CreateTargetHandler()(w, r)
		} else {
			handler.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "このエンドポイントはGETまたはPOSTメソッドのみ対応しています", nil)
		}
	})
	mux.HandleFunc("/api/v1/targets/", func(w http.ResponseWriter, r *http.Request) {
		// 対象と行動種別の紐づけ操作を優先的に処理
		if strings.Contains(r.URL.Path, "/action-types") {
			if r.Method == http.MethodGet {
				targetActionTypeHandler.ListByTargetHandler()(w, r)
			} else if r.Method == http.MethodPost {
				targetActionTypeHandler.LinkHandler()(w, r)
			} else if r.Method == http.MethodDelete {
				targetActionTypeHandler.UnlinkHandler()(w, r)
			} else {
				handler.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "このエンドポイントはGET/POST/DELETEメソッドのみ対応しています", nil)
			}
			return
		}

		targetHandler.TargetDetailHandler()(w, r)
	})

	// 行動記録のエンドポイント（認証必須）
	mux.HandleFunc("/api/v1/action_logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			actionLogHandler.ListHandler()(w, r)
		} else if r.Method == http.MethodPost {
			actionLogHandler.CreateHandler()(w, r)
		} else {
			handler.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "このエンドポイントはGETまたはPOSTメソッドのみ対応しています", nil)
		}
	})
	mux.HandleFunc("/api/v1/action_logs/", actionLogHandler.DetailHandler())

	// 行動種別のエンドポイント（認証必須）
	mux.HandleFunc("/api/v1/action_types", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			actionTypeHandler.ListHandler()(w, r)
		} else if r.Method == http.MethodPost {
			actionTypeHandler.CreateHandler()(w, r)
		} else {
			handler.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "このエンドポイントはGETまたはPOSTメソッドのみ対応しています", nil)
		}
	})
	mux.HandleFunc("/api/v1/action_types/", actionTypeHandler.DetailHandler())

	// JWT認証とCORSミドルウェアを順に適用
	handlerWithAuth := jwtProtectedMux(mux)
	// CORSミドルウェアを適用
	handlerWithCORS := corsMiddleware(handlerWithAuth)

	log.Println("サーバーがポート8080で起動しました")
	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		log.Fatalf("HTTPサーバー起動失敗: %v", err)
	}
}
