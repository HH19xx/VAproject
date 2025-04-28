package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
)

func RunMigrations(db *sql.DB) error {
	migrationFile := "./infra/db/migrations/0001_create_action_types.sql"

	sqlBytes, err := ioutil.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("マイグレーションファイルの読み込みに失敗しました: %w", err)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("マイグレーション実行に失敗しました: %w", err)
	}

	return nil
}
