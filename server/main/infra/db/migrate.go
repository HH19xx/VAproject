package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// マイグレーション用のSQLファイルを順に読み込みDBに適用
func RunMigrations(db *sql.DB) error {
	// 適用するマイグレーションファイル一覧を配列に入れる
	migrationFiles := []string{
		"0001_create_users.up.sql",
		"0003_create_refresh_tokens.up.sql",
		"0004_create_user_auth_providers.up.sql",
		"0001_create_targets.up.sql",
		"0001_create_action_types.sql",
		"0002_create_target_action_types.up.sql",
		"0005_add_and_search_schema.up.sql",
		"0006_create_prototypes.up.sql",
		"0001_create_action_logs.up.sql",
		"0001_create_custom_actions.up.sql",
		"0009_create_analysis_snapshots.up.sql",
		"0010_create_world_signals.up.sql",
	}

	// migrationsディレクトリの基底パスを指定
	basePath := "main/infra/db/migrations"

	// マイグレーションファイルを順に適用
	for _, file := range migrationFiles {
		fullPath := filepath.Join(basePath, file)

		// マイグレーションファイルを読み込む
		sqlBytes, err := os.ReadFile(fullPath)
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

// test_dataディレクトリ内のSQLファイルを順に読み込みテストデータをDBに挿入
// test_dataディレクトリが存在しないまたは空の場合は何も実行せず本番環境とみなす
func RunTestDataMigrations(db *sql.DB) error {
	// test_dataディレクトリのパスを指定
	testDataPath := "main/infra/db/migrations/test_data"

	// test_dataディレクトリ内のファイル一覧を取得
	files, err := os.ReadDir(testDataPath)
	if err != nil {
		// ディレクトリが存在しない場合は本番環境とみなし、ログを出力して正常終了
		log.Println("test_dataディレクトリが存在しないため、テストデータのマイグレーションをスキップします（本番環境）")
		return nil
	}

	// ディレクトリが空の場合も本番環境とみなす
	if len(files) == 0 {
		log.Println("test_dataディレクトリが空のため、テストデータのマイグレーションをスキップします（本番環境）")
		return nil
	}

	log.Println("テストデータのマイグレーションを開始します（開発/テスト環境）")

	// test_dataディレクトリ内のSQLファイルを順に適用
	for _, file := range files {
		// ディレクトリはスキップ
		if file.IsDir() {
			continue
		}

		// .sqlファイルのみ処理
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		fullPath := filepath.Join(testDataPath, file.Name())

		// SQLファイルを読み込む
		sqlBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("テストデータファイルの読み込みに失敗しました (%s): %w", fullPath, err)
		}

		// 適用しているSQLをログに出力
		log.Printf("テストデータ実行中: %s\n", file.Name())

		// SQLを実行
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("テストデータSQL実行に失敗しました (%s): %w", file.Name(), err)
		}
	}

	log.Println("すべてのテストデータが正常に適用されました。")
	return nil
}
