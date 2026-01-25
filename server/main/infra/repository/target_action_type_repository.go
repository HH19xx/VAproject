package repository

import (
	"context"
	"database/sql"

	"server/main/domain"
)

type TargetActionTypeRepository struct {
	db *sql.DB
}

func NewTargetActionTypeRepository(db *sql.DB) *TargetActionTypeRepository {
	return &TargetActionTypeRepository{db: db}
}

// FindByTargetID は指定された対象に紐づく行動種別一覧を取得
func (r *TargetActionTypeRepository) FindByTargetID(ctx context.Context, targetID int) ([]*domain.TargetActionType, error) {
	query := `
		SELECT id, target_id, action_type_id, created_at, create_user
		FROM target_action_types
		WHERE target_id = $1
		ORDER BY id
	`
	rows, err := r.db.QueryContext(ctx, query, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.TargetActionType
	for rows.Next() {
		var t domain.TargetActionType
		if err := rows.Scan(&t.ID, &t.TargetID, &t.ActionTypeID, &t.CreatedAt, &t.CreateUser); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, rows.Err()
}

// FindByActionTypeID は指定された行動種別に紐づく対象一覧を取得
func (r *TargetActionTypeRepository) FindByActionTypeID(ctx context.Context, actionTypeID int) ([]*domain.TargetActionType, error) {
	query := `
		SELECT id, target_id, action_type_id, created_at, create_user
		FROM target_action_types
		WHERE action_type_id = $1
		ORDER BY id
	`
	rows, err := r.db.QueryContext(ctx, query, actionTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.TargetActionType
	for rows.Next() {
		var t domain.TargetActionType
		if err := rows.Scan(&t.ID, &t.TargetID, &t.ActionTypeID, &t.CreatedAt, &t.CreateUser); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, rows.Err()
}

// Create は新しい紐づけを作成
func (r *TargetActionTypeRepository) Create(ctx context.Context, tat *domain.TargetActionType) (int, error) {
	query := `
		INSERT INTO target_action_types (target_id, action_type_id, create_user)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query, tat.TargetID, tat.ActionTypeID, tat.CreateUser).Scan(&id)
	return id, err
}

// Delete は紐づけを削除
func (r *TargetActionTypeRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM target_action_types WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DeleteByTargetAndActionType は対象IDと行動種別IDの組み合わせで削除
func (r *TargetActionTypeRepository) DeleteByTargetAndActionType(ctx context.Context, targetID, actionTypeID int) error {
	query := `DELETE FROM target_action_types WHERE target_id = $1 AND action_type_id = $2`
	_, err := r.db.ExecContext(ctx, query, targetID, actionTypeID)
	return err
}

// Exists は紐づけが存在するか確認
func (r *TargetActionTypeRepository) Exists(ctx context.Context, targetID, actionTypeID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM target_action_types WHERE target_id = $1 AND action_type_id = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, targetID, actionTypeID).Scan(&exists)
	return exists, err
}
