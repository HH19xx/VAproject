package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// NewPostgresConnection は、PostgreSQLデータベースへの接続を確立します。
// 起動直後のDBがまだ準備できていない可能性に備えて、接続を最大10回までリトライします。
func NewPostgresConnection() (*sql.DB, error) {
	// 環境変数からDB接続情報を取得
	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "admin")
	password := getEnv("DB_PASSWORD", "secret")
	dbname := getEnv("DB_NAME", "master")

	// ポート番号は文字列で取得されるので、整数に変換
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("ポート番号の変換に失敗しました: %w", err)
	}

	// DSN（Data Source Name）を作成
	// 例: "host=localhost port=5432 user=admin password=secret dbname=master sslmode=disable"
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var db *sql.DB

	// 最大10回まで、3秒間隔で接続をリトライ
	for i := 1; i <= 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			fmt.Printf("[DB接続 %d回目] sql.Open エラー: %v\n", i, err)
		} else if pingErr := db.Ping(); pingErr != nil {
			fmt.Printf("[DB接続 %d回目] Ping失敗: %v\n", i, pingErr)
		} else {
			fmt.Println("PostgreSQLへの接続に成功しました。")
			return db, nil
		}

		fmt.Printf("PostgreSQLへの接続待機中...（%d回目）\n", i)
		time.Sleep(3 * time.Second)
	}

	// 最後まで接続できなかった場合、エラーを返す
	return nil, fmt.Errorf("PostgreSQLへの接続に失敗しました（%v）", err)
}

// getEnv は、環境変数を取得し、未設定の場合はデフォルト値を返します。
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func init() {
	// .env ファイルを読み込む。存在しない場合は無視。
	if err := godotenv.Load(); err != nil {
		log.Println(".env ファイルが見つかりません（Docker環境なら問題ありません）")
	}
}
