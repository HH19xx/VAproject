package service

import (
	"log"

	"VAproject/infra/db"
)

func InitRegistrationService() {
	conn, err := db.NewPostgresConnection()
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer conn.Close()

	// ここでクエリ等を実行できる
}

