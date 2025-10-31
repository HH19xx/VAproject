package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// NewPostgresConnection establishes a PostgreSQL connection with retry logic.
func NewPostgresConnection() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "admin")
	password := getEnv("DB_PASSWORD", "secret")
	dbname := getEnv("DB_NAME", "master")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var db *sql.DB
	for i := 1; i <= 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			fmt.Printf("[DB attempt %d] sql.Open error: %v\n", i, err)
		} else if pingErr := db.Ping(); pingErr != nil {
			fmt.Printf("[DB attempt %d] ping error: %v\n", i, pingErr)
		} else {
			fmt.Println("Connected to PostgreSQL.")
			return db, nil
		}

		fmt.Printf("Waiting for PostgreSQL... (%d/10)\n", i)
		time.Sleep(3 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("failed to get working directory: %v", err)
		return
	}

	for i := 0; i < 6; i++ {
		candidate := filepath.Join(cwd, ".env")
		if _, err := os.Stat(candidate); err == nil {
			if loadErr := godotenv.Overload(candidate); loadErr == nil {
				log.Printf("loaded .env from %s", candidate)
				promoteEnv("DB_USER", "POSTGRES_USER")
				promoteEnv("DB_PASSWORD", "POSTGRES_PASSWORD")
				promoteEnv("DB_NAME", "POSTGRES_DB")
				return
			}
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}

	log.Println(".env file not found (expected in local development).")
}

func promoteEnv(target, source string) {
	if os.Getenv(target) == "" {
		if v := os.Getenv(source); v != "" {
			os.Setenv(target, v)
		}
	}
}
