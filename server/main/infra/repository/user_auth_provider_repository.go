package repository

import (
	"context"
	"database/sql"
	"server/main/domain"
)

// UserAuthProviderRepositoryインターフェースの具体実装
type userAuthProviderRepository struct {
	db *sql.DB
}

// userAuthProviderRepositoryの生成関数
func NewUserAuthProviderRepository(db *sql.DB) *userAuthProviderRepository {
	return &userAuthProviderRepository{db: db}
}

// ユーザーIDで認証プロバイダー一覧を取得（論理削除済みを除外）
func (r *userAuthProviderRepository) FindByUserID(ctx context.Context, userID int) ([]*domain.UserAuthProvider, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, provider, COALESCE(provider_user_id, '')
		FROM user_auth_providers
		WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*domain.UserAuthProvider
	for rows.Next() {
		var p domain.UserAuthProvider
		if err := rows.Scan(&p.ID, &p.UserID, &p.Provider, &p.ProviderUserID); err != nil {
			return nil, err
		}
		providers = append(providers, &p)
	}
	return providers, nil
}

// 指定ユーザーが特定の認証プロバイダーを持つかチェック（論理削除済みを除外）
func (r *userAuthProviderRepository) HasProvider(ctx context.Context, userID int, provider string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_auth_providers
			WHERE user_id = $1 AND provider = $2 AND deleted_at IS NULL
		)`, userID, provider).Scan(&exists)
	return exists, err
}

// 認証プロバイダーを追加
func (r *userAuthProviderRepository) Create(ctx context.Context, authProvider *domain.UserAuthProvider) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_auth_providers (user_id, provider, provider_user_id, create_user)
		VALUES ($1, $2, $3, 'system')`,
		authProvider.UserID, authProvider.Provider, nullIfEmpty(authProvider.ProviderUserID))
	return err
}

// 認証プロバイダーを削除
func (r *userAuthProviderRepository) Delete(ctx context.Context, userID int, provider string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_auth_providers 
		WHERE user_id = $1 AND provider = $2`,
		userID, provider)
	return err
}

// 空文字の場合はNULLを返すヘルパー関数
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
