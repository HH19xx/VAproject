package db

import (
	"context"
	"database/sql"

	identitydomain "server/internal/modules/identity/domain"
)

type UserAuthProviderRepository struct {
	db *sql.DB
}

func NewUserAuthProviderRepository(db *sql.DB) *UserAuthProviderRepository {
	return &UserAuthProviderRepository{db: db}
}

func (r *UserAuthProviderRepository) FindByUserID(ctx context.Context, userID int) ([]*identitydomain.UserAuthProvider, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, provider, COALESCE(provider_user_id, '')
		FROM user_auth_providers
		WHERE user_id = $1 AND deleted_at IS NULL`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*identitydomain.UserAuthProvider
	for rows.Next() {
		var p identitydomain.UserAuthProvider
		if err := rows.Scan(&p.ID, &p.UserID, &p.Provider, &p.ProviderUserID); err != nil {
			return nil, err
		}
		providers = append(providers, &p)
	}
	return providers, nil
}

func (r *UserAuthProviderRepository) HasProvider(ctx context.Context, userID int, provider string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_auth_providers
			WHERE user_id = $1 AND provider = $2 AND deleted_at IS NULL
		)`, userID, provider).Scan(&exists)
	return exists, err
}

func (r *UserAuthProviderRepository) Create(ctx context.Context, authProvider *identitydomain.UserAuthProvider) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_auth_providers (user_id, provider, provider_user_id, create_user)
		VALUES ($1, $2, $3, 'system')`,
		authProvider.UserID, authProvider.Provider, nullIfEmpty(authProvider.ProviderUserID))
	return err
}

func (r *UserAuthProviderRepository) Delete(ctx context.Context, userID int, provider string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_auth_providers
		WHERE user_id = $1 AND provider = $2`,
		userID, provider)
	return err
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
