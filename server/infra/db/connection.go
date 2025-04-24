package db

import (
    "database/sql"
    "fmt"

    _"github.com/lib/pq"
)

func NewPostgresConnection() (*sql.DB, error) {
    host := "db"
    port := 5432
    user := "admin"
    password := "secret"
    dbname := "master"

    dsn :=
}
