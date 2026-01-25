package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"server/main/domain"
)

type ActionLogRepository interface {
	FindAll(ctx context.Context, targetID int, limit, offset int) ([]*domain.ActionLog, error)
	Count(ctx context.Context, targetID int) (int, error)
	FindByID(ctx context.Context, id int) (*domain.ActionLog, error)
	Create(ctx context.Context, log *domain.ActionLog) (int, error)
	Update(ctx context.Context, log *domain.ActionLog) error
	Delete(ctx context.Context, id int) error
}

type actionLogRepository struct {
	db *sql.DB
}

func NewActionLogRepository(db *sql.DB) ActionLogRepository {
	return &actionLogRepository{db: db}
}

func (r *actionLogRepository) FindAll(ctx context.Context, targetID int, limit, offset int) ([]*domain.ActionLog, error) {
	query := `
		SELECT id, target_id, action_type, timestamp, notes,
		       created_at, create_user, updated_at, update_user, deleted_at
		FROM action_logs
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}
	paramIndex := 1
	if targetID > 0 {
		query += fmt.Sprintf(" AND target_id = $%d", paramIndex)
		args = append(args, targetID)
		paramIndex++
	}
	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", paramIndex, paramIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.ActionLog
	for rows.Next() {
		var l domain.ActionLog
		var target sql.NullInt64
		if err := rows.Scan(
			&l.ID, &target, &l.ActionType, &l.Timestamp, &l.Notes,
			&l.CreatedAt, &l.CreateUser, &l.UpdatedAt, &l.UpdateUser, &l.DeletedAt,
		); err != nil {
			return nil, err
		}
		if target.Valid {
			val := int(target.Int64)
			l.TargetID = &val
		} else {
			l.TargetID = nil
		}
		logs = append(logs, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *actionLogRepository) Count(ctx context.Context, targetID int) (int, error) {
	query := `SELECT COUNT(*) FROM action_logs WHERE deleted_at IS NULL`
	args := []interface{}{}
	if targetID > 0 {
		query += " AND target_id = $1"
		args = append(args, targetID)
	}
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *actionLogRepository) FindByID(ctx context.Context, id int) (*domain.ActionLog, error) {
	query := `
		SELECT id, target_id, action_type, timestamp, notes,
		       created_at, create_user, updated_at, update_user, deleted_at
		FROM action_logs
		WHERE id = $1 AND deleted_at IS NULL
	`
	var l domain.ActionLog
	var target sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&l.ID, &target, &l.ActionType, &l.Timestamp, &l.Notes,
		&l.CreatedAt, &l.CreateUser, &l.UpdatedAt, &l.UpdateUser, &l.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	if target.Valid {
		val := int(target.Int64)
		l.TargetID = &val
	} else {
		l.TargetID = nil
	}
	return &l, nil
}

func (r *actionLogRepository) Create(ctx context.Context, log *domain.ActionLog) (int, error) {
	var target sql.NullInt64
	if log.TargetID != nil {
		target = sql.NullInt64{Int64: int64(*log.TargetID), Valid: true}
	}
	query := `
		INSERT INTO action_logs (target_id, action_type, timestamp, notes, create_user, update_user)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query,
		target, log.ActionType, log.Timestamp, log.Notes, log.CreateUser, log.UpdateUser,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *actionLogRepository) Update(ctx context.Context, log *domain.ActionLog) error {
	var target sql.NullInt64
	if log.TargetID != nil {
		target = sql.NullInt64{Int64: int64(*log.TargetID), Valid: true}
	}
	query := `
		UPDATE action_logs
		SET target_id = $1,
		    action_type = $2,
		    timestamp = $3,
		    notes = $4,
		    update_user = $5,
		    updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query,
		target, log.ActionType, log.Timestamp, log.Notes, log.UpdateUser, time.Now().UTC(), log.ID,
	)
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

func (r *actionLogRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE action_logs
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, id)
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
