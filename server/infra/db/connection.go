package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewPostgresConnection() (*sql.DB, error) {
	host := "localhost"  // docker-composeのservice名（開発段階ではローカルなので「localhost」）
	port := 5432         // PostgreSQLポート
	user := "admin"      // POSTGRES_USER
	password := "secret" // POSTGRES_PASSWORD
	dbname := "master"   // POSTGRES_DB

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}
