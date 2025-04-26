package repository

import (
	"VAproject/server/core/domain"
	"database/sql"
	"fmt"
)

// UserRepository はユーザーデータへのアクセスを抽象化するインターフェースです。
type UserRepository interface {
	Save(user *domain.User) error
	// 他に必要なメソッド（例: FindByID, FindByEmailなど）があればここに追加します。
}

// userRepository はUserRepositoryインターフェースの実装です。
type userRepository struct {
	db *sql.DB
}

// NewUserRepository は新しいuserRepositoryのインスタンスを作成します。
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Save はユーザー情報をデータベースに保存します。
func (r *userRepository) Save(user *domain.User) error {
	// ここでデータベースへの保存処理を実装します。
	// 今回はadminテーブルに保存することを想定します。
	query := `INSERT INTO admin (name, email) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRow(query, user.Name, user.Email).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("ユーザー情報の保存に失敗しました: %w", err)
	}
	return nil
}
