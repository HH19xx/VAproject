package db

import (
	"context"
	"database/sql"
	"time"

	inferencedomain "server/internal/modules/inference/domain"
)

type legacyTagStore struct {
	db *sql.DB
}

func newLegacyTagStore(db *sql.DB) tagStore {
	return &legacyTagStore{db: db}
}

func (s *legacyTagStore) FindAll(ctx context.Context, userID int) ([]*inferencedomain.Tag, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, name, group_name, description, created_at, create_user, updated_at, update_user, deleted_at
		FROM tags
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*inferencedomain.Tag
	for rows.Next() {
		var tag inferencedomain.Tag
		if err := rows.Scan(
			&tag.ID, &tag.UserID, &tag.Name, &tag.GroupName, &tag.Description,
			&tag.CreatedAt, &tag.CreateUser, &tag.UpdatedAt, &tag.UpdateUser, &tag.DeletedAt,
		); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}

func (s *legacyTagStore) FindByID(ctx context.Context, id, userID int) (*inferencedomain.Tag, error) {
	var tag inferencedomain.Tag
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, group_name, description, created_at, create_user, updated_at, update_user, deleted_at
		FROM tags
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, userID).Scan(
		&tag.ID, &tag.UserID, &tag.Name, &tag.GroupName, &tag.Description,
		&tag.CreatedAt, &tag.CreateUser, &tag.UpdatedAt, &tag.UpdateUser, &tag.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (s *legacyTagStore) Create(ctx context.Context, tag *inferencedomain.Tag) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO tags (user_id, name, group_name, description, create_user, update_user)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, tag.UserID, tag.Name, tag.GroupName, tag.Description, tag.CreateUser, tag.UpdateUser).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *legacyTagStore) Update(ctx context.Context, tag *inferencedomain.Tag) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE tags
		SET name = $1, group_name = $2, description = $3, update_user = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`, tag.Name, tag.GroupName, tag.Description, tag.UpdateUser, time.Now().UTC(), tag.ID, tag.UserID)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *legacyTagStore) Delete(ctx context.Context, id, userID int) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE tags
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, userID)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}
