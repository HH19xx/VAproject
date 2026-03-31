package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	inferenceapp "server/internal/modules/inference/app"
	inferencedomain "server/internal/modules/inference/domain"
)

type legacyActionLogStore struct {
	db *sql.DB
}

func newLegacyActionLogStore(db *sql.DB) actionLogStore {
	return &legacyActionLogStore{db: db}
}

func (s *legacyActionLogStore) FindAll(
	ctx context.Context,
	userID int,
	tagIDs []int,
	limit, offset int,
	options inferenceapp.ActionLogListOptions,
) ([]*inferencedomain.ActionLog, error) {
	sortExpr := mapActionLogSortExpr(options.Sort)
	orderExpr := mapActionLogOrderExpr(options.Order)

	query := `
		SELECT al.id, al.user_id, al.title, al.prototype_id, al.occurred_at, al.notes,
		       al.created_at, al.create_user, al.updated_at, al.update_user, al.deleted_at
		FROM action_logs al
		WHERE al.deleted_at IS NULL
		  AND al.user_id = $1
	`
	args := []interface{}{userID}
	paramIndex := 2

	query, args, paramIndex = appendActionLogWindowFilters(query, args, paramIndex, options)
	query, args, paramIndex = appendActionLogTagFilters(query, args, paramIndex, tagIDs, options)

	query += fmt.Sprintf(" ORDER BY %s %s, al.id DESC LIMIT $%d OFFSET $%d", sortExpr, orderExpr, paramIndex, paramIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*inferencedomain.ActionLog, 0)
	for rows.Next() {
		var log inferencedomain.ActionLog
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.Title, &log.PrototypeID, &log.OccurredAt, &log.Notes,
			&log.CreatedAt, &log.CreateUser, &log.UpdatedAt, &log.UpdateUser, &log.DeletedAt,
		); err != nil {
			return nil, err
		}

		log.TagIDs, err = s.findTagIDsByActionLogID(ctx, log.ID)
		if err != nil {
			return nil, err
		}
		log.Attributes, err = s.findAttributesByActionLogID(ctx, log.ID)
		if err != nil {
			return nil, err
		}

		logs = append(logs, &log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *legacyActionLogStore) Count(
	ctx context.Context,
	userID int,
	tagIDs []int,
	options inferenceapp.ActionLogListOptions,
) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM action_logs al
		WHERE al.deleted_at IS NULL
		  AND al.user_id = $1
	`
	args := []interface{}{userID}
	paramIndex := 2

	query, args, paramIndex = appendActionLogWindowFilters(query, args, paramIndex, options)
	query, args, _ = appendActionLogTagFilters(query, args, paramIndex, tagIDs, options)

	var count int
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *legacyActionLogStore) FindByID(ctx context.Context, id, userID int) (*inferencedomain.ActionLog, error) {
	query := `
		SELECT id, user_id, title, prototype_id, occurred_at, notes,
		       created_at, create_user, updated_at, update_user, deleted_at
		FROM action_logs
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
	`

	var log inferencedomain.ActionLog
	if err := s.db.QueryRowContext(ctx, query, id, userID).Scan(
		&log.ID, &log.UserID, &log.Title, &log.PrototypeID, &log.OccurredAt, &log.Notes,
		&log.CreatedAt, &log.CreateUser, &log.UpdatedAt, &log.UpdateUser, &log.DeletedAt,
	); err != nil {
		return nil, err
	}

	var err error
	log.TagIDs, err = s.findTagIDsByActionLogID(ctx, log.ID)
	if err != nil {
		return nil, err
	}
	log.Attributes, err = s.findAttributesByActionLogID(ctx, log.ID)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (s *legacyActionLogStore) ListAttributeKeys(ctx context.Context, userID int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
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

func (s *legacyActionLogStore) Create(ctx context.Context, log *inferencedomain.ActionLog) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
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
	if err := tx.QueryRowContext(
		ctx,
		query,
		log.UserID,
		log.Title,
		log.PrototypeID,
		log.OccurredAt,
		log.OccurredAt,
		log.Notes,
		log.CreateUser,
		log.UpdateUser,
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

func (s *legacyActionLogStore) Update(ctx context.Context, log *inferencedomain.ActionLog) error {
	tx, err := s.db.BeginTx(ctx, nil)
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

	result, err := tx.ExecContext(
		ctx,
		query,
		log.Title,
		log.PrototypeID,
		log.OccurredAt,
		log.OccurredAt,
		log.Notes,
		log.UpdateUser,
		time.Now().UTC(),
		log.ID,
		log.UserID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
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

func (s *legacyActionLogStore) Delete(ctx context.Context, id, userID int) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE action_logs
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1
		  AND user_id = $2
		  AND deleted_at IS NULL
	`, id, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *legacyActionLogStore) findTagIDsByActionLogID(ctx context.Context, actionLogID int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tag_id
		FROM action_log_tags
		WHERE action_log_id = $1
		ORDER BY tag_id ASC
	`, actionLogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagIDs := make([]int, 0)
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

func (s *legacyActionLogStore) findAttributesByActionLogID(ctx context.Context, actionLogID int) ([]inferencedomain.ActionLogAttribute, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT attr_key, value_number
		FROM action_log_attributes
		WHERE action_log_id = $1
		ORDER BY attr_key ASC
	`, actionLogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attributes := make([]inferencedomain.ActionLogAttribute, 0)
	for rows.Next() {
		var attr inferencedomain.ActionLogAttribute
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
	`, actionLogID, user, pq.Array(uniqueIntIDs(tagIDs)))
	return err
}

func replaceActionLogAttributes(
	ctx context.Context,
	tx *sql.Tx,
	actionLogID int,
	attributes []inferencedomain.ActionLogAttribute,
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

func appendActionLogWindowFilters(
	query string,
	args []interface{},
	paramIndex int,
	options inferenceapp.ActionLogListOptions,
) (string, []interface{}, int) {
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
	return query, args, paramIndex
}

func appendActionLogTagFilters(
	query string,
	args []interface{},
	paramIndex int,
	tagIDs []int,
	options inferenceapp.ActionLogListOptions,
) (string, []interface{}, int) {
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

	return query, args, paramIndex
}

func mapActionLogSortExpr(sort string) string {
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

func mapActionLogOrderExpr(order string) string {
	if order == "asc" {
		return "ASC"
	}
	return "DESC"
}
