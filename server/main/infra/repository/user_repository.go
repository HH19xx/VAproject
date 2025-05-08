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
		return nil, err
	}
	return &user, nil
}

// FindByEmail はOAuthで使うemail（nameフィールドに格納）からユーザーを検索します。
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create は指定されたユーザー（主にOAuthユーザー）を登録します。
func (r *userRepository) Create(ctx context.Context, user *domain.User) (int, error) {
	err := r.db.QueryRowContext(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id", user.Email).
		Scan(&user.ID)
	return user.ID, err
}
