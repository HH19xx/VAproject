package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func NewPostgresConnection() (*sql.DB, error) {
	// 環境変数からDB接続情報を取得
	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "5432")
	// ポート番号は文字列で取得されるので、整数に変換
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("ポート番号の変換に失敗しました: %w", err)
	}
	user := getEnv("DB_USER", "admin")
	password := getEnv("DB_PASSWORD", "secret")
	dbname := getEnv("DB_NAME", "master")

	// dsn（Data Source Name）を作成
	// 例: "host=localhost port=5432 user=admin password=secret dbname=master sslmode=disable"
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// ヘルパー関数: 環境変数を取得し、なければデフォルト値を返す
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
