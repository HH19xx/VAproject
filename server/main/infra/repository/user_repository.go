package repository

import (
	"context"
	"database/sql"
	"server/main/domain"
)

// userRepository は UserRepository インターフェースの具体実装です。
type userRepository struct {
	db *sql.DB
}

// NewUserRepository は userRepository の生成関数です。
func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

// FindByName はユーザー名でユーザーを検索し、見つかれば User エンティティを返します。
func (r *userRepository) FindByName(ctx context.Context, name string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx, "SELECT id, name, password FROM users WHERE name = $1", name).
		Scan(&user.ID, &user.Name, &user.Password)

	if err != nil {
		// 上層で ErrAuthFailed に変換されます
		return nil, err
	}
	return &user, nil
}
