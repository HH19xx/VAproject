package repository

import (
	"context"
	"database/sql"
	"time"

	"server/main/domain"
)

// TagRepository は tags テーブルに対する永続化操作を提供する
type TagRepository interface {
	FindAll(ctx context.Context, userID int) ([]*domain.Tag, error)
	FindByID(ctx context.Context, id, userID int) (*domain.Tag, error)
	Create(ctx context.Context, tag *domain.Tag) (int, error)
	Update(ctx context.Context, tag *domain.Tag) error
	Delete(ctx context.Context, id, userID int) error
}

type tagRepository struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) FindAll(ctx context.Context, userID int) ([]*domain.Tag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, group_name, description, created_at, create_user, updated_at, update_user, deleted_at
		FROM tags
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		var tag domain.Tag
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

func (r *tagRepository) FindByID(ctx context.Context, id, userID int) (*domain.Tag, error) {
	var tag domain.Tag
	err := r.db.QueryRowContext(ctx, `
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

func (r *tagRepository) Create(ctx context.Context, tag *domain.Tag) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO tags (user_id, name, group_name, description, create_user, update_user)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, tag.UserID, tag.Name, tag.GroupName, tag.Description, tag.CreateUser, tag.UpdateUser).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *tagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	res, err := r.db.ExecContext(ctx, `
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

func (r *tagRepository) Delete(ctx context.Context, id, userID int) error {
	res, err := r.db.ExecContext(ctx, `
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
