package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

// RunMigrations は、マイグレーション用のSQLファイルを順に読み込み、DBに適用します。
func RunMigrations(db *sql.DB) error {
	// 適用するマイグレーションファイル一覧（順序が大切）を配列に入れる
	migrationFiles := []string{
		"0001_create_users.up.sql",
		"0001_create_targets.up.sql",
		"0001_create_action_types.sql",
		"0001_create_custom_actions.up.sql",
	}

	// migrationsディレクトリの基底パスを指定
	basePath := "main/infra/db/migrations"

	// マイグレーションファイルを順に適用
	for _, file := range migrationFiles {
		fullPath := filepath.Join(basePath, file)

		// マイグレーションファイルを読み込む
		sqlBytes, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("マイグレーションファイルの読み込みに失敗しました (%s): %w", fullPath, err)
		}

		// 適用しているSQLをログに出力
		log.Printf("マイグレーション実行中: %s\n", file)

		// SQLを実行
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("マイグレーションSQL実行に失敗しました (%s): %w", file, err)
		}
	}

	log.Println("すべてのマイグレーションが正常に適用されました。")
	return nil
}
