package db

import (
	"testing"
)

func TestPostgresConnection(t *testing.T) {
	db, err := NewPostgresConnection()
	if err != nil {
		t.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		t.Fatalf("データベースPingに失敗しました: %v", err)
	}
}
