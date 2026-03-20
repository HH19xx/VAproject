package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"server/main/domain"
)

type PrototypeRepository interface {
	FindAll(ctx context.Context, userID int) ([]*domain.Prototype, error)
	FindByID(ctx context.Context, id, userID int) (*domain.Prototype, error)
	Create(ctx context.Context, prototype *domain.Prototype) (int, error)
	Update(ctx context.Context, prototype *domain.Prototype) error
	Delete(ctx context.Context, id, userID int) error
}

type prototypeRepository struct {
	db *sql.DB
}

func NewPrototypeRepository(db *sql.DB) PrototypeRepository {
	return &prototypeRepository{db: db}
}

func (r *prototypeRepository) FindAll(ctx context.Context, userID int) ([]*domain.Prototype, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, description, parent_prototype_id, created_at, create_user, updated_at, update_user, deleted_at
		FROM prototypes
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prototypes []*domain.Prototype
	for rows.Next() {
		var p domain.Prototype
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Name, &p.Description, &p.ParentPrototypeID,
			&p.CreatedAt, &p.CreateUser, &p.UpdatedAt, &p.UpdateUser, &p.DeletedAt,
		); err != nil {
			return nil, err
		}
		tagIDs, err := r.findTagIDsByPrototypeID(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.TagIDs = tagIDs
		prototypes = append(prototypes, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return prototypes, nil
}

func (r *prototypeRepository) FindByID(ctx context.Context, id, userID int) (*domain.Prototype, error) {
	var p domain.Prototype
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, description, parent_prototype_id, created_at, create_user, updated_at, update_user, deleted_at
		FROM prototypes
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.Description, &p.ParentPrototypeID,
		&p.CreatedAt, &p.CreateUser, &p.UpdatedAt, &p.UpdateUser, &p.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	tagIDs, err := r.findTagIDsByPrototypeID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.TagIDs = tagIDs
	return &p, nil
}

func (r *prototypeRepository) Create(ctx context.Context, prototype *domain.Prototype) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO prototypes (user_id, name, description, parent_prototype_id, create_user, update_user)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, prototype.UserID, prototype.Name, prototype.Description, prototype.ParentPrototypeID, prototype.CreateUser, prototype.UpdateUser).Scan(&id)
	if err != nil {
		return 0, err
	}

	if err := replacePrototypeTags(ctx, tx, id, prototype.TagIDs, prototype.CreateUser); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *prototypeRepository) Update(ctx context.Context, prototype *domain.Prototype) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		UPDATE prototypes
		SET name = $1, description = $2, parent_prototype_id = $3, update_user = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND user_id = $6 AND deleted_at IS NULL
	`, prototype.Name, prototype.Description, prototype.ParentPrototypeID, prototype.UpdateUser, prototype.ID, prototype.UserID)
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

	if err := replacePrototypeTags(ctx, tx, prototype.ID, prototype.TagIDs, prototype.UpdateUser); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *prototypeRepository) Delete(ctx context.Context, id, userID int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE prototypes
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

func (r *prototypeRepository) findTagIDsByPrototypeID(ctx context.Context, prototypeID int) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tag_id
		FROM prototype_tags
		WHERE prototype_id = $1
		ORDER BY tag_id ASC
	`, prototypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tagIDs []int
	for rows.Next() {
		var tagID int
		if err := rows.Scan(&tagID); err != nil {
			return nil, err
		}
		tagIDs = append(tagIDs, tagID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tagIDs, nil
}

func replacePrototypeTags(ctx context.Context, tx *sql.Tx, prototypeID int, tagIDs []int, user string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM prototype_tags WHERE prototype_id = $1`, prototypeID); err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO prototype_tags (prototype_id, tag_id, create_user)
		SELECT $1, tag_id, $2
		FROM unnest($3::int[]) AS tag_id
		ON CONFLICT (prototype_id, tag_id) DO NOTHING
	`, prototypeID, user, pq.Array(uniqueIntIDs(tagIDs)))
	return err
}
