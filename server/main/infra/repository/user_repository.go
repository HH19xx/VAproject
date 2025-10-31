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
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email, password_hash FROM users WHERE username = $1", name).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail はOAuthで使うemailからユーザーを検索します。
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create は指定されたユーザー（主にOAuthユーザー）を登録します。
func (r *userRepository) Create(ctx context.Context, user *domain.User) (int, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password_hash, create_user, update_user) 
		VALUES ($1, $2, '', 'oauth', 'oauth') 
		RETURNING id`, 
		user.Name, user.Email).Scan(&user.ID)
	return user.ID, err
}
