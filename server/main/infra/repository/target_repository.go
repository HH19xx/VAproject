package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"

	"server/main/domain"
)

type targetRepository struct {
	db *sql.DB
}

func NewTargetRepository(db *sql.DB) *targetRepository {
	return &targetRepository{db: db}
}

func (r *targetRepository) FindAll(ctx context.Context, userID, limit, offset int) ([]*domain.Target, error) {
	query := `
		SELECT id, user_id, name, description, match_mode,
		       query_text, any_tag_ids, any_tag_groups, exclude_tag_ids,
		       created_at, create_user, updated_at, update_user, deleted_at
		FROM targets
		WHERE deleted_at IS NULL AND user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []*domain.Target
	for rows.Next() {
		var t domain.Target
		var anyTagGroupsRaw string
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Name, &t.Description, &t.MatchMode,
			&t.QueryText, pq.Array(&t.AnyTagIDs), &anyTagGroupsRaw, pq.Array(&t.ExcludeTagIDs),
			&t.CreatedAt, &t.CreateUser, &t.UpdatedAt, &t.UpdateUser, &t.DeletedAt,
		); err != nil {
			return nil, err
		}
		t.AnyTagGroups = decodeTagGroups(anyTagGroupsRaw)
		tagIDs, err := r.findTagIDsByTargetID(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		t.TagIDs = tagIDs
		targets = append(targets, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return targets, nil
}

func (r *targetRepository) FindByID(ctx context.Context, id, userID int) (*domain.Target, error) {
	query := `
		SELECT id, user_id, name, description, match_mode,
		       query_text, any_tag_ids, any_tag_groups, exclude_tag_ids,
		       created_at, create_user, updated_at, update_user, deleted_at
		FROM targets
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	var t domain.Target
	var anyTagGroupsRaw string
	if err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&t.ID, &t.UserID, &t.Name, &t.Description, &t.MatchMode,
		&t.QueryText, pq.Array(&t.AnyTagIDs), &anyTagGroupsRaw, pq.Array(&t.ExcludeTagIDs),
		&t.CreatedAt, &t.CreateUser, &t.UpdatedAt, &t.UpdateUser, &t.DeletedAt,
	); err != nil {
		return nil, err
	}

	t.AnyTagGroups = decodeTagGroups(anyTagGroupsRaw)
	tagIDs, err := r.findTagIDsByTargetID(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	t.TagIDs = tagIDs
	return &t, nil
}

func (r *targetRepository) Create(ctx context.Context, target *domain.Target) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
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
		pq.Array(uniqueIntIDs(target.AnyTagIDs)),
		encodeTagGroups(target.AnyTagGroups),
		pq.Array(uniqueIntIDs(target.ExcludeTagIDs)),
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

func (r *targetRepository) Update(ctx context.Context, target *domain.Target) error {
	tx, err := r.db.BeginTx(ctx, nil)
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
		pq.Array(uniqueIntIDs(target.AnyTagIDs)),
		encodeTagGroups(target.AnyTagGroups),
		pq.Array(uniqueIntIDs(target.ExcludeTagIDs)),
		target.UpdateUser,
		target.ID,
		target.UserID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	if err := replaceTargetTags(ctx, tx, target.ID, target.TagIDs, target.UpdateUser); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *targetRepository) Delete(ctx context.Context, id, userID int) error {
	query := `
		UPDATE targets
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *targetRepository) Count(ctx context.Context, userID int) (int, error) {
	query := `SELECT COUNT(*) FROM targets WHERE deleted_at IS NULL AND user_id = $1`
	var count int
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *targetRepository) FindActionLogsByTargetID(ctx context.Context, targetID, userID, limit, offset int) ([]*domain.ActionLog, error) {
	target, err := r.FindByID(ctx, targetID, userID)
	if err != nil {
		return nil, err
	}

	query, args := buildActionLogSearchByTargetQuery(target, userID, limit, offset, false)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.ActionLog
	for rows.Next() {
		var l domain.ActionLog
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.OccurredAt, &l.Notes,
			&l.CreatedAt, &l.CreateUser, &l.UpdatedAt, &l.UpdateUser, &l.DeletedAt,
		); err != nil {
			return nil, err
		}
		tagIDs, err := r.findTagIDsByActionLogID(ctx, l.ID)
		if err != nil {
			return nil, err
		}
		l.TagIDs = tagIDs
		logs = append(logs, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *targetRepository) CountActionLogsByTargetID(ctx context.Context, targetID, userID int) (int, error) {
	target, err := r.FindByID(ctx, targetID, userID)
	if err != nil {
		return 0, err
	}

	query, args := buildActionLogSearchByTargetQuery(target, userID, 0, 0, true)
	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func buildActionLogSearchByTargetQuery(target *domain.Target, userID, limit, offset int, countOnly bool) (string, []interface{}) {
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

func (r *targetRepository) findTagIDsByTargetID(ctx context.Context, targetID int) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tag_id
		FROM target_tags
		WHERE target_id = $1
		ORDER BY tag_id ASC
	`, targetID)
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

func (r *targetRepository) findTagIDsByActionLogID(ctx context.Context, actionLogID int) ([]int, error) {
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

func replaceTargetTags(ctx context.Context, tx *sql.Tx, targetID int, tagIDs []int, user string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM target_tags WHERE target_id = $1`, targetID); err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}

	uniqueTagIDs := uniqueIntIDs(tagIDs)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO target_tags (target_id, tag_id, create_user)
		SELECT $1, tag_id, $2
		FROM unnest($3::int[]) AS tag_id
		ON CONFLICT (target_id, tag_id) DO NOTHING
	`, targetID, user, pq.Array(uniqueTagIDs))
	return err
}

func uniqueIntIDs(ids []int) []int {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(ids))
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func encodeTagGroups(groups [][]int) string {
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

func decodeTagGroups(raw string) [][]int {
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
