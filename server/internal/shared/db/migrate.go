package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func RunMigrations(db *sql.DB) error {
	migrationFiles := append(coreMigrationFiles(), deferenceMigrationFiles()...)
	return runMigrationFiles(db, migrationFiles)
}

func RunCoreMigrations(db *sql.DB) error {
	return runMigrationFiles(db, coreMigrationFiles())
}

func RunDeferenceMigrations(db *sql.DB) error {
	return runMigrationFiles(db, deferenceMigrationFiles())
}

func coreMigrationFiles() []string {
	return []string{
		"0001_create_users.up.sql",
		"0003_create_refresh_tokens.up.sql",
		"0004_create_user_auth_providers.up.sql",
		"0001_create_targets.up.sql",
		"0005_add_and_search_schema.up.sql",
		"0006_create_prototypes.up.sql",
		"0001_create_action_logs.up.sql",
		"0010_create_world_signals.up.sql",
	}
}

func deferenceMigrationFiles() []string {
	return []string{
		"0009_create_analysis_snapshots.up.sql",
	}
}

func runMigrationFiles(db *sql.DB, migrationFiles []string) error {
	basePath := "internal/shared/db/migrations"

	for _, file := range migrationFiles {
		fullPath := filepath.Join(basePath, file)
		sqlBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file (%s): %w", fullPath, err)
		}

		log.Printf("running migration: %s\n", file)
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("failed to execute migration SQL (%s): %w", file, err)
		}
	}

	log.Println("all migrations completed successfully")
	return nil
}

func RunTestDataMigrations(db *sql.DB) error {
	testDataPath := "internal/shared/db/migrations/test_data"

	files, err := os.ReadDir(testDataPath)
	if err != nil {
		log.Println("test_data directory not found. skipping test-data migrations")
		return nil
	}

	if len(files) == 0 {
		log.Println("test_data directory is empty. skipping test-data migrations")
		return nil
	}

	log.Println("starting test-data migrations")

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		fullPath := filepath.Join(testDataPath, file.Name())
		sqlBytes, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read test-data file (%s): %w", fullPath, err)
		}

		log.Printf("running test-data migration: %s\n", file.Name())
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("failed to execute test-data SQL (%s): %w", file.Name(), err)
		}
	}

	log.Println("all test-data migrations completed successfully")
	return nil
}
