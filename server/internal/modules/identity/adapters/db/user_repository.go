package db

import (
	"context"
	"database/sql"

	identitydomain "server/internal/modules/identity/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByName(ctx context.Context, name string) (*identitydomain.User, error) {
	var user identitydomain.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email, password_hash FROM users WHERE username = $1 AND deleted_at IS NULL", name).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*identitydomain.User, error) {
	var user identitydomain.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email FROM users WHERE email = $1 AND deleted_at IS NULL", email).
		Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *identitydomain.User) (int, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password_hash, create_user, update_user)
		VALUES ($1, $2, $3, 'system', 'system')
		RETURNING id`,
		user.Name, user.Email, user.Password).Scan(&user.ID)
	return user.ID, err
}

func (r *UserRepository) FindByID(ctx context.Context, userID int) (*identitydomain.User, error) {
	var user identitydomain.User
	err := r.db.QueryRowContext(ctx, "SELECT id, username, email FROM users WHERE id = $1 AND deleted_at IS NULL", userID).
		Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateName(ctx context.Context, userID int, newName string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET username = $1, update_user = 'system', updated_at = CURRENT_TIMESTAMP
		WHERE id = $2`,
		newName, userID)
	return err
}
