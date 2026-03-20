package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"server/main/domain"
	"server/main/usecases"
)

type ActionLogRepository interface {
	FindAll(ctx context.Context, userID int, tagIDs []int, limit, offset int, options usecases.ActionLogListOptions) ([]*domain.ActionLog, error)
	Count(ctx context.Context, userID int, tagIDs []int, options usecases.ActionLogListOptions) (int, error)
	FindByID(ctx context.Context, id, userID int) (*domain.ActionLog, error)
	ListAttributeKeys(ctx context.Context, userID int) ([]string, error)
	Create(ctx context.Context, log *domain.ActionLog) (int, error)
	Update(ctx context.Context, log *domain.ActionLog) error
	Delete(ctx context.Context, id, userID int) error
}

type actionLogRepository struct {
	db *sql.DB
}

func NewActionLogRepository(db *sql.DB) ActionLogRepository {
	return &actionLogRepository{db: db}
}

func (r *actionLogRepository) FindAll(
	ctx context.Context,
	userID int,
	tagIDs []int,
	limit, offset int,
	options usecases.ActionLogListOptions,
) ([]*domain.ActionLog, error) {
	sortExpr := mapSortExpr(options.Sort)
	orderExpr := mapOrderExpr(options.Order)

	query := `
		SELECT al.id, al.user_id, al.title, al.prototype_id, al.occurred_at, al.notes,
		       al.created_at, al.create_user, al.updated_at, al.update_user, al.deleted_at
		FROM action_logs al
		WHERE al.deleted_at IS NULL
		  AND al.user_id = $1
	`
	args := []interface{}{userID}
	paramIndex := 2

	if options.From != nil {
		query += fmt.Sprintf(" AND al.occurred_at >= $%d", paramIndex)
		args = append(args, *options.From)
		paramIndex++
	}
	if options.To != nil {
		query += fmt.Sprintf(" AND al.occurred_at <= $%d", paramIndex)
		args = append(args, *options.To)
		paramIndex++
	}
	if len(tagIDs) > 0 {
		query += fmt.Sprintf(`
		  AND NOT EXISTS (
		    SELECT 1
		    FROM unnest($%d::int[]) AS q(tag_id)
		    WHERE NOT EXISTS (
		      SELECT 1
		      FROM action_log_tags alt
		      WHERE alt.action_log_id = al.id
		        AND alt.tag_id = q.tag_id
		    )
		  )
		`, paramIndex)
		args = append(args, pq.Array(tagIDs))
		paramIndex++
	}
	if len(options.AnyTagIDs) > 0 {
		query += fmt.Sprintf(`
		  AND EXISTS (
		    SELECT 1
		    FROM action_log_tags alt
		    WHERE alt.action_log_id = al.id
		      AND alt.tag_id = ANY($%d::int[])
		  )
		`, paramIndex)
		args = append(args, pq.Array(options.AnyTagIDs))
		paramIndex++
	}
	if len(options.AnyTagGroups) > 0 {
		query += " AND ("
		for i, group := range options.AnyTagGroups {
			if i > 0 {
				query += " OR "
			}
			query += fmt.Sprintf(`
			  NOT EXISTS (
			    SELECT 1
			    FROM unnest($%d::int[]) AS q(tag_id)
			    WHERE NOT EXISTS (
			      SELECT 1
			      FROM action_log_tags alt
			      WHERE alt.action_log_id = al.id
			        AND alt.tag_id = q.tag_id
			    )
			  )
			`, paramIndex)
			args = append(args, pq.Array(group))
			paramIndex++
		}
		query += " )"
	}
	if len(options.ExcludeTagIDs) > 0 {
		query += fmt.Sprintf(`
		  AND NOT EXISTS (
		    SELECT 1
		    FROM action_log_tags alt
		    WHERE alt.action_log_id = al.id
		      AND alt.tag_id = ANY($%d::int[])
		  )
		`, paramIndex)
		args = append(args, pq.Array(options.ExcludeTagIDs))
		paramIndex++
	}
	query += fmt.Sprintf(" ORDER BY %s %s, al.id DESC LIMIT $%d OFFSET $%d", sortExpr, orderExpr, paramIndex, paramIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.ActionLog
	for rows.Next() {
		var l domain.ActionLog
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.Title, &l.PrototypeID, &l.OccurredAt, &l.Notes,
			&l.CreatedAt, &l.CreateUser, &l.UpdatedAt, &l.UpdateUser, &l.DeletedAt,
		); err != nil {
			return nil, err
		}
		tagIDs, err := r.findTagIDsByActionLogID(ctx, l.ID)
		if err != nil {
			return nil, err
		}
		l.TagIDs = tagIDs
		attributes, err := r.findAttributesByActionLogID(ctx, l.ID)
		if err != nil {
			return nil, err
		}
		l.Attributes = attributes
		logs = append(logs, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *actionLogRepository) Count(
	ctx context.Context,
	userID int,
	tagIDs []int,
	options usecases.ActionLogListOptions,
) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM action_logs al
		WHERE al.deleted_at IS NULL
		  AND al.user_id = $1
	`
	args := []interface{}{userID}
	paramIndex := 2

	if options.From != nil {
		query += fmt.Sprintf(" AND al.occurred_at >= $%d", paramIndex)
		args = append(args, *options.From)
		paramIndex++
	}
	if options.To != nil {
		query += fmt.Sprintf(" AND al.occurred_at <= $%d", paramIndex)
		args = append(args, *options.To)
		paramIndex++
	}
	if len(tagIDs) > 0 {
		query += fmt.Sprintf(`
		  AND NOT EXISTS (
		    SELECT 1
		    FROM unnest($%d::int[]) AS q(tag_id)
		    WHERE NOT EXISTS (
		      SELECT 1
		      FROM action_log_tags alt
		      WHERE alt.action_log_id = al.id
		        AND alt.tag_id = q.tag_id
		    )
		  )
		`, paramIndex)
		args = append(args, pq.Array(tagIDs))
		paramIndex++
	}
	if len(options.AnyTagIDs) > 0 {
		query += fmt.Sprintf(`
		  AND EXISTS (
		    SELECT 1
		    FROM action_log_tags alt
		    WHERE alt.action_log_id = al.id
		      AND alt.tag_id = ANY($%d::int[])
		  )
		`, paramIndex)
		args = append(args, pq.Array(options.AnyTagIDs))
		paramIndex++
	}
	if len(options.AnyTagGroups) > 0 {
		query += " AND ("
		for i, group := range options.AnyTagGroups {
			if i > 0 {
				query += " OR "
			}
			query += fmt.Sprintf(`
			  NOT EXISTS (
			    SELECT 1
			    FROM unnest($%d::int[]) AS q(tag_id)
			    WHERE NOT EXISTS (
			      SELECT 1
			      FROM action_log_tags alt
			      WHERE alt.action_log_id = al.id
			        AND alt.tag_id = q.tag_id
			    )
			  )
			`, paramIndex)
			args = append(args, pq.Array(group))
			paramIndex++
		}
		query += " )"
	}
	if len(options.ExcludeTagIDs) > 0 {
		query += fmt.Sprintf(`
		  AND NOT EXISTS (
		    SELECT 1
		    FROM action_log_tags alt
		    WHERE alt.action_log_id = al.id
		      AND alt.tag_id = ANY($%d::int[])
		  )
		`, paramIndex)
		args = append(args, pq.Array(options.ExcludeTagIDs))
	}

	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *actionLogRepository) FindByID(ctx context.Context, id, userID int) (*domain.ActionLog, error) {
	query := `
		SELECT id, user_id, title, prototype_id, occurred_at, notes,
		       created_at, create_user, updated_at, update_user, deleted_at
		FROM action_logs
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
	`
	var l domain.ActionLog
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&l.ID, &l.UserID, &l.Title, &l.PrototypeID, &l.OccurredAt, &l.Notes,
		&l.CreatedAt, &l.CreateUser, &l.UpdatedAt, &l.UpdateUser, &l.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	tagIDs, err := r.findTagIDsByActionLogID(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	l.TagIDs = tagIDs
	attributes, err := r.findAttributesByActionLogID(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	l.Attributes = attributes
	return &l, nil
}

func (r *actionLogRepository) Create(ctx context.Context, log *domain.ActionLog) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO action_logs (user_id, title, prototype_id, occurred_at, "timestamp", notes, create_user, update_user)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	var id int
	if err := tx.QueryRowContext(ctx, query,
		log.UserID, log.Title, log.PrototypeID, log.OccurredAt, log.OccurredAt, log.Notes, log.CreateUser, log.UpdateUser,
	).Scan(&id); err != nil {
		return 0, err
	}

	if err := replaceActionLogTags(ctx, tx, id, log.TagIDs, log.CreateUser); err != nil {
		return 0, err
	}
	if err := replaceActionLogAttributes(ctx, tx, id, log.Attributes, log.CreateUser, log.UpdateUser); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *actionLogRepository) Update(ctx context.Context, log *domain.ActionLog) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE action_logs
		SET title = $1,
		    prototype_id = $2,
		    occurred_at = $3,
		    "timestamp" = $4,
		    notes = $5,
		    update_user = $6,
		    updated_at = $7
		WHERE id = $8
		  AND user_id = $9
		  AND deleted_at IS NULL
	`
	res, err := tx.ExecContext(ctx, query,
		log.Title, log.PrototypeID, log.OccurredAt, log.OccurredAt, log.Notes, log.UpdateUser, time.Now().UTC(), log.ID, log.UserID,
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

	if err := replaceActionLogTags(ctx, tx, log.ID, log.TagIDs, log.UpdateUser); err != nil {
		return err
	}
	if err := replaceActionLogAttributes(ctx, tx, log.ID, log.Attributes, log.UpdateUser, log.UpdateUser); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *actionLogRepository) Delete(ctx context.Context, id, userID int) error {
	query := `
		UPDATE action_logs
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, id, userID)
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

func (r *actionLogRepository) ListAttributeKeys(ctx context.Context, userID int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT ala.attr_key
		FROM action_log_attributes ala
		INNER JOIN action_logs al ON al.id = ala.action_log_id
		WHERE al.user_id = $1
		  AND al.deleted_at IS NULL
		ORDER BY ala.attr_key ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]string, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *actionLogRepository) findTagIDsByActionLogID(ctx context.Context, actionLogID int) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tag_id
		FROM action_log_tags
		WHERE action_log_id = $1
		ORDER BY tag_id ASC
	`, actionLogID)
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

func (r *actionLogRepository) findAttributesByActionLogID(ctx context.Context, actionLogID int) ([]domain.ActionLogAttribute, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT attr_key, value_number
		FROM action_log_attributes
		WHERE action_log_id = $1
		ORDER BY attr_key ASC
	`, actionLogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attributes := make([]domain.ActionLogAttribute, 0)
	for rows.Next() {
		var attr domain.ActionLogAttribute
		if err := rows.Scan(&attr.Key, &attr.ValueNumber); err != nil {
			return nil, err
		}
		attributes = append(attributes, attr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return attributes, nil
}

func replaceActionLogTags(ctx context.Context, tx *sql.Tx, actionLogID int, tagIDs []int, user string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM action_log_tags WHERE action_log_id = $1`, actionLogID); err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO action_log_tags (action_log_id, tag_id, create_user)
		SELECT $1, tag_id, $2
		FROM unnest($3::int[]) AS tag_id
		ON CONFLICT (action_log_id, tag_id) DO NOTHING
	`, actionLogID, user, pq.Array(tagIDs))
	return err
}

func replaceActionLogAttributes(
	ctx context.Context,
	tx *sql.Tx,
	actionLogID int,
	attributes []domain.ActionLogAttribute,
	createUser, updateUser string,
) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM action_log_attributes WHERE action_log_id = $1`, actionLogID); err != nil {
		return err
	}
	if len(attributes) == 0 {
		return nil
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO action_log_attributes (
			action_log_id, attr_key, value_number, create_user, update_user
		)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, attr := range attributes {
		if _, err := stmt.ExecContext(ctx, actionLogID, attr.Key, attr.ValueNumber, createUser, updateUser); err != nil {
			return err
		}
	}
	return nil
}

func mapSortExpr(sort string) string {
	switch sort {
	case "created_at":
		return "al.created_at"
	case "updated_at":
		return "al.updated_at"
	case "title":
		return "al.title"
	case "tag_count":
		return "(SELECT COUNT(*) FROM action_log_tags alt WHERE alt.action_log_id = al.id)"
	default:
		return "al.occurred_at"
	}
}

func mapOrderExpr(order string) string {
	if order == "asc" {
		return "ASC"
	}
	return "DESC"
}
