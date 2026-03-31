package db

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"

	inferencedomain "server/internal/modules/inference/domain"
)

type legacyTargetStore struct {
	db *sql.DB
}

func newLegacyTargetStore(db *sql.DB) targetStore {
	return &legacyTargetStore{db: db}
}

func (s *legacyTargetStore) FindAll(ctx context.Context, userID, limit, offset int) ([]*inferencedomain.Target, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, name, description, match_mode,
		       query_text, any_tag_ids, any_tag_groups, exclude_tag_ids,
		       created_at, create_user, updated_at, update_user, deleted_at
		FROM targets
		WHERE deleted_at IS NULL AND user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := make([]*inferencedomain.Target, 0)
	for rows.Next() {
		var target inferencedomain.Target
		var anyTagGroupsRaw string
		if err := rows.Scan(
			&target.ID, &target.UserID, &target.Name, &target.Description, &target.MatchMode,
			&target.QueryText, pq.Array(&target.AnyTagIDs), &anyTagGroupsRaw, pq.Array(&target.ExcludeTagIDs),
			&target.CreatedAt, &target.CreateUser, &target.UpdatedAt, &target.UpdateUser, &target.DeletedAt,
		); err != nil {
			return nil, err
		}

		target.AnyTagGroups = decodeTargetTagGroups(anyTagGroupsRaw)
		target.TagIDs, err = s.findTagIDsByTargetID(ctx, target.ID)
		if err != nil {
			return nil, err
		}
		targets = append(targets, &target)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return targets, nil
}

func (s *legacyTargetStore) FindByID(ctx context.Context, id, userID int) (*inferencedomain.Target, error) {
	var target inferencedomain.Target
	var anyTagGroupsRaw string
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, description, match_mode,
		       query_text, any_tag_ids, any_tag_groups, exclude_tag_ids,
		       created_at, create_user, updated_at, update_user, deleted_at
		FROM targets
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, id, userID).Scan(
		&target.ID, &target.UserID, &target.Name, &target.Description, &target.MatchMode,
		&target.QueryText, pq.Array(&target.AnyTagIDs), &anyTagGroupsRaw, pq.Array(&target.ExcludeTagIDs),
		&target.CreatedAt, &target.CreateUser, &target.UpdatedAt, &target.UpdateUser, &target.DeletedAt,
	); err != nil {
		return nil, err
	}

	target.AnyTagGroups = decodeTargetTagGroups(anyTagGroupsRaw)

	var err error
	target.TagIDs, err = s.findTagIDsByTargetID(ctx, target.ID)
	if err != nil {
		return nil, err
	}
	return &target, nil
}

func (s *legacyTargetStore) Create(ctx context.Context, target *inferencedomain.Target) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO targets (
			user_id, name, description, match_mode,
			query_text, any_tag_ids, any_tag_groups, exclude_tag_ids,
			create_user, update_user
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var id int
	if err := tx.QueryRowContext(
		ctx,
		query,
		target.UserID,
		target.Name,
		target.Description,
		target.MatchMode,
		target.QueryText,
		pq.Array(normalizeTargetIntArray(target.AnyTagIDs)),
		encodeTargetTagGroups(target.AnyTagGroups),
		pq.Array(normalizeTargetIntArray(target.ExcludeTagIDs)),
		target.CreateUser,
		target.UpdateUser,
	).Scan(&id); err != nil {
		return 0, err
	}

	if err := replaceTargetTags(ctx, tx, id, target.TagIDs, target.CreateUser); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *legacyTargetStore) Update(ctx context.Context, target *inferencedomain.Target) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE targets
		SET name = $1,
		    description = $2,
		    match_mode = $3,
		    query_text = $4,
		    any_tag_ids = $5,
		    any_tag_groups = $6,
		    exclude_tag_ids = $7,
		    update_user = $8,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $9 AND user_id = $10 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		target.Name,
		target.Description,
		target.MatchMode,
		target.QueryText,
		pq.Array(normalizeTargetIntArray(target.AnyTagIDs)),
		encodeTargetTagGroups(target.AnyTagGroups),
		pq.Array(normalizeTargetIntArray(target.ExcludeTagIDs)),
		target.UpdateUser,
		target.ID,
		target.UserID,
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

	if err := replaceTargetTags(ctx, tx, target.ID, target.TagIDs, target.UpdateUser); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *legacyTargetStore) Delete(ctx context.Context, id, userID int) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE targets
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
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

func (s *legacyTargetStore) Count(ctx context.Context, userID int) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM targets
		WHERE deleted_at IS NULL AND user_id = $1
	`, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *legacyTargetStore) FindActionLogsByTargetID(ctx context.Context, targetID, userID, limit, offset int) ([]*inferencedomain.ActionLog, error) {
	target, err := s.FindByID(ctx, targetID, userID)
	if err != nil {
		return nil, err
	}

	query, args := buildActionLogSearchByTargetQuery(target, userID, limit, offset, false)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*inferencedomain.ActionLog, 0)
	for rows.Next() {
		var log inferencedomain.ActionLog
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.OccurredAt, &log.Notes,
			&log.CreatedAt, &log.CreateUser, &log.UpdatedAt, &log.UpdateUser, &log.DeletedAt,
		); err != nil {
			return nil, err
		}

		log.TagIDs, err = s.findTagIDsByActionLogID(ctx, log.ID)
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

func (s *legacyTargetStore) CountActionLogsByTargetID(ctx context.Context, targetID, userID int) (int, error) {
	target, err := s.FindByID(ctx, targetID, userID)
	if err != nil {
		return 0, err
	}

	query, args := buildActionLogSearchByTargetQuery(target, userID, 0, 0, true)
	var count int
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func buildActionLogSearchByTargetQuery(target *inferencedomain.Target, userID, limit, offset int, countOnly bool) (string, []interface{}) {
	selectClause := `
		SELECT al.id, al.user_id, al.occurred_at, al.notes,
		       al.created_at, al.create_user, al.updated_at, al.update_user, al.deleted_at
	`
	if countOnly {
		selectClause = `SELECT COUNT(*)`
	}

	query := selectClause + `
		FROM action_logs al
		WHERE al.deleted_at IS NULL
		  AND al.user_id = $1
	`
	args := []interface{}{userID}
	paramIndex := 2

	if len(target.TagIDs) > 0 {
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
		args = append(args, pq.Array(target.TagIDs))
		paramIndex++
	}

	if len(target.AnyTagIDs) > 0 {
		query += fmt.Sprintf(`
		  AND EXISTS (
		    SELECT 1
		    FROM action_log_tags alt
		    WHERE alt.action_log_id = al.id
		      AND alt.tag_id = ANY($%d::int[])
		  )
		`, paramIndex)
		args = append(args, pq.Array(target.AnyTagIDs))
		paramIndex++
	}

	if len(target.AnyTagGroups) > 0 {
		query += " AND ("
		for i, group := range target.AnyTagGroups {
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

	if len(target.ExcludeTagIDs) > 0 {
		query += fmt.Sprintf(`
		  AND NOT EXISTS (
		    SELECT 1
		    FROM action_log_tags alt
		    WHERE alt.action_log_id = al.id
		      AND alt.tag_id = ANY($%d::int[])
		  )
		`, paramIndex)
		args = append(args, pq.Array(target.ExcludeTagIDs))
		paramIndex++
	}

	if !countOnly {
		query += fmt.Sprintf(" ORDER BY al.occurred_at DESC, al.id DESC LIMIT $%d OFFSET $%d", paramIndex, paramIndex+1)
		args = append(args, limit, offset)
	}

	return query, args
}

func (s *legacyTargetStore) findTagIDsByTargetID(ctx context.Context, targetID int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT tag_id
		FROM target_tags
		WHERE target_id = $1
		ORDER BY tag_id ASC
	`, targetID)
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

func (s *legacyTargetStore) findTagIDsByActionLogID(ctx context.Context, actionLogID int) ([]int, error) {
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

func replaceTargetTags(ctx context.Context, tx *sql.Tx, targetID int, tagIDs []int, user string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM target_tags WHERE target_id = $1`, targetID); err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO target_tags (target_id, tag_id, create_user)
		SELECT $1, tag_id, $2
		FROM unnest($3::int[]) AS tag_id
		ON CONFLICT (target_id, tag_id) DO NOTHING
	`, targetID, user, pq.Array(uniqueIntIDs(tagIDs)))
	return err
}

func normalizeTargetIntArray(ids []int) []int {
	normalized := uniqueIntIDs(ids)
	if normalized == nil {
		return []int{}
	}
	return normalized
}

func encodeTargetTagGroups(groups [][]int) string {
	if len(groups) == 0 {
		return ""
	}

	parts := make([]string, 0, len(groups))
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}

		members := make([]string, 0, len(group))
		for _, id := range uniqueIntIDs(group) {
			members = append(members, strconv.Itoa(id))
		}
		if len(members) > 0 {
			parts = append(parts, strings.Join(members, "+"))
		}
	}

	return strings.Join(parts, ",")
}

func decodeTargetTagGroups(raw string) [][]int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	groupParts := strings.Split(raw, ",")
	groups := make([][]int, 0, len(groupParts))
	for _, groupRaw := range groupParts {
		groupRaw = strings.TrimSpace(groupRaw)
		if groupRaw == "" {
			continue
		}

		itemParts := strings.Split(groupRaw, "+")
		group := make([]int, 0, len(itemParts))
		for _, item := range itemParts {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}

			id, err := strconv.Atoi(item)
			if err != nil || id <= 0 {
				continue
			}
			group = append(group, id)
		}

		group = uniqueIntIDs(group)
		if len(group) > 0 {
			groups = append(groups, group)
		}
	}

	if len(groups) == 0 {
		return nil
	}
	return groups
}
