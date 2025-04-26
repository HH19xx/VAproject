package main

import (
	"VAproject/server/apps/registration/handler"
	"VAproject/server/apps/registration/routes"
	"VAproject/server/core/usecases"
	"VAproject/server/infra/db"
	"VAproject/server/infra/repository"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors" // CORSミドルウェアをインポート
)

func main() {
	// CORSミドルウェアの設定
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // フロントエンドのオリジンを許可
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// データベース接続の確立
	// 接続情報は環境変数などから取得するように後で修正する必要があります。
	database, err := db.NewDBConnection()
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer database.Close()

	// リポジトリ、ユースケース、ハンドラの初期化
	userRepo := repository.NewUserRepository(database)
	userUseCase := usecases.NewUserUseCase(userRepo)
	registrationHandler := handler.NewRegistrationHandler(userUseCase)

	// ルーターの設定
	router := mux.NewRouter()
	routes.SetupRegistrationRoutes(router, registrationHandler)

	// ルーターにCORSミドルウェアを適用
	corsHandler := c.Handler(router)

	// HTTPサーバーの起動
	// ポート番号は設定ファイルなどから取得するように後で修正する必要があります。
	port := ":8080"
	log.Printf("サーバーをポート %s で起動します...", port)
	// CORSミドルウェアを適用したハンドラを使用してサーバーを起動
	err = http.ListenAndServe(port, corsHandler)
	if err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
