package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // PostgreSQLドライバー
)

// NewDBConnection は新しいデータベース接続を確立します。
func NewDBConnection() (*sql.DB, error) {
	// 後ほど環境変数に変更。
	// Docker Composeの設定に合わせて接続情報を記述します。
	// ホストOSからDockerコンテナに接続するため、hostをlocalhostに設定します。
	connStr := "user=admin password=secret dbname=master host=localhost sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("データベース接続のオープンに失敗しました: %w", err)
	}

	// データベースへの接続を確認します。
	err = db.Ping()
	if err != nil {
		db.Close() // Pingに失敗した場合は接続を閉じます
		return nil, fmt.Errorf("データベースへのPingに失敗しました: %w", err)
	}

	return db, nil
}
