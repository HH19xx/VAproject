package repository

import (
	"context"
	"database/sql"
	"server/main/domain"
)

// UserRepositoryインターフェースの具体実装
type userRepository struct {
	db *sql.DB
}

// userRepositoryの生成関数。
func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

// ユーザー名で検索しUserエンティティを返す（論理削除済みを除外）
func (r *userRepository) FindByName(ctx context.Context, name string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email, password_hash FROM users WHERE username = $1 AND deleted_at IS NULL", name).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// emailで検索しユーザーを返す（OAuth用、論理削除済みを除外）
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email FROM users WHERE email = $1 AND deleted_at IS NULL", email).
		Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ユーザーを登録
func (r *userRepository) Create(ctx context.Context, user *domain.User) (int, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password_hash, create_user, update_user)
		VALUES ($1, $2, $3, 'system', 'system')
		RETURNING id`,
		user.Name, user.Email, user.Password).Scan(&user.ID)
	return user.ID, err
}

// ユーザーIDで情報を取得（プロフィール編集用、論理削除済みを除外）
func (r *userRepository) FindByID(ctx context.Context, userID int) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email FROM users WHERE id = $1 AND deleted_at IS NULL", userID).
		Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ユーザー名を更新（プロフィール編集用）
func (r *userRepository) UpdateName(ctx context.Context, userID int, newName string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET username = $1, update_user = 'system', updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`,
		newName, userID)
	return err
}
