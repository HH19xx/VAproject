package repository

import (
	"context"
	"database/sql"
	"server/main/domain"
	"time"
)

type ActionTypeRepository interface {
	FindAll(ctx context.Context) ([]*domain.ActionType, error)
	FindByID(ctx context.Context, id int) (*domain.ActionType, error)
	Create(ctx context.Context, at *domain.ActionType) (int, error)
	Update(ctx context.Context, at *domain.ActionType) error
	Delete(ctx context.Context, id int) error
}

type actionTypeRepository struct {
	db *sql.DB
}

func NewActionTypeRepository(db *sql.DB) ActionTypeRepository {
	return &actionTypeRepository{db: db}
}

func (r *actionTypeRepository) FindAll(ctx context.Context) ([]*domain.ActionType, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, action_name, description, created_at, create_user, updated_at, update_user, deleted_at
		FROM action_types
		WHERE deleted_at IS NULL
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*domain.ActionType
	for rows.Next() {
		var at domain.ActionType
		if err := rows.Scan(&at.ID, &at.ActionName, &at.Description, &at.CreatedAt, &at.CreateUser, &at.UpdatedAt, &at.UpdateUser, &at.DeletedAt); err != nil {
			return nil, err
		}
		res = append(res, &at)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *actionTypeRepository) FindByID(ctx context.Context, id int) (*domain.ActionType, error) {
	var at domain.ActionType
	err := r.db.QueryRowContext(ctx, `
		SELECT id, action_name, description, created_at, create_user, updated_at, update_user, deleted_at
		FROM action_types
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&at.ID, &at.ActionName, &at.Description, &at.CreatedAt, &at.CreateUser, &at.UpdatedAt, &at.UpdateUser, &at.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &at, nil
}

func (r *actionTypeRepository) Create(ctx context.Context, at *domain.ActionType) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO action_types (action_name, description, create_user, update_user)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, at.ActionName, at.Description, at.CreateUser, at.UpdateUser).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *actionTypeRepository) Update(ctx context.Context, at *domain.ActionType) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE action_types
		SET action_name = $1, description = $2, update_user = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`, at.ActionName, at.Description, at.UpdateUser, time.Now().UTC(), at.ID)
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

func (r *actionTypeRepository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE action_types
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
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
