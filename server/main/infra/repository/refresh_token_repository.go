package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"server/main/domain"
)

// リフレッシュトークン操作の具象実装
type RefreshTokenRepository struct {
	db *sql.DB
}

// リフレッシュトークン用リポジトリを生成。
func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// 新規リフレッシュトークンを保存
func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	const query = `
        INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked, created_at, create_user, updated_at, update_user)
        VALUES ($1, $2, $3, $4, NOW(), $5, NOW(), $6)
        RETURNING id, created_at, updated_at
    `

	return r.db.QueryRowContext(ctx, query,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.Revoked,
		token.CreateUser,
		token.UpdateUser,
	).Scan(&token.ID, &token.CreatedAt, &token.UpdatedAt)
}

// ハッシュ値で検索
func (r *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	const query = `
        SELECT id, user_id, token_hash, expires_at, revoked, created_at, create_user, updated_at, update_user
        FROM refresh_tokens
        WHERE token_hash = $1
        LIMIT 1
    `

	token := &domain.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.Revoked,
		&token.CreatedAt,
		&token.CreateUser,
		&token.UpdatedAt,
		&token.UpdateUser,
	)

	if err != nil {
		return nil, err
	}
	return token, nil
}

// 指定IDのリフレッシュトークンを失効
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id int, updateUser string) error {
	const query = `
        UPDATE refresh_tokens
        SET revoked = TRUE,
            updated_at = NOW(),
            update_user = $1
        WHERE id = $2
    `

	res, err := r.db.ExecContext(ctx, query, updateUser, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("refresh token not found")
	}
	return nil
}

// 期限切れトークンを削除（メンテ用）
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context, threshold time.Time) (int64, error) {
	const query = `
        DELETE FROM refresh_tokens
        WHERE expires_at < $1
    `

	res, err := r.db.ExecContext(ctx, query, threshold)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
