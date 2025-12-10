package repository

import (
	"context"
	"database/sql"
	"server/main/domain"
)

// TargetRepositoryインターフェースの具体実装
type targetRepository struct {
	db *sql.DB
}

// targetRepositoryの生成関数。
func NewTargetRepository(db *sql.DB) *targetRepository {
	return &targetRepository{db: db}
}

// 削除されていない観察対象の一覧を取得（ページネーション対応）
// limit: 取得件数の上限、offset: スキップする件数
func (r *targetRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Target, error) {
	// 論理削除されていないレコードのみを取得
	query := `
		SELECT id, name, description, created_at, create_user, updated_at, update_user, deleted_at
		FROM targets
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 取得した行をTargetエンティティのスライスに変換
	var targets []*domain.Target
	for rows.Next() {
		var t domain.Target
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.CreateUser, &t.UpdatedAt, &t.UpdateUser, &t.DeletedAt); err != nil {
			return nil, err
		}
		targets = append(targets, &t)
	}

	// ループ中のエラーを確認
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

// 指定IDの削除されていない観察対象を1件取得
func (r *targetRepository) FindByID(ctx context.Context, id int) (*domain.Target, error) {
	// 論理削除されていないレコードのみを取得
	query := `
		SELECT id, name, description, created_at, create_user, updated_at, update_user, deleted_at
		FROM targets
		WHERE id = $1 AND deleted_at IS NULL
	`
	var t domain.Target
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&t.ID, &t.Name, &t.Description, &t.CreatedAt, &t.CreateUser, &t.UpdatedAt, &t.UpdateUser, &t.DeletedAt)

	if err != nil {
		return nil, err
	}
	return &t, nil
}

// 新しい観察対象をデータベースに登録し生成されたIDを返す
func (r *targetRepository) Create(ctx context.Context, target *domain.Target) (int, error) {
	// INSERTを実行し自動生成されたIDを取得
	query := `
		INSERT INTO targets (name, description, create_user, update_user)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var id int
	err := r.db.QueryRowContext(ctx, query,
		target.Name,
		target.Description,
		target.CreateUser,
		target.UpdateUser,
	).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

// 削除されていない既存の観察対象の情報を更新
func (r *targetRepository) Update(ctx context.Context, target *domain.Target) error {
	// 論理削除されていないレコードのみを更新
	query := `
		UPDATE targets
		SET name = $1, description = $2, update_user = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query,
		target.Name,
		target.Description,
		target.UpdateUser,
		target.ID,
	)
	if err != nil {
		return err
	}

	// 更新対象の行が存在したか確認
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// 指定IDの観察対象を論理削除
// deleted_atに現在時刻を設定することで削除済みとする
func (r *targetRepository) Delete(ctx context.Context, id int) error {
	// deleted_atをCURRENT_TIMESTAMPで更新し論理削除を実行
	query := `
		UPDATE targets
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// 削除対象の行が存在したか確認
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// 削除されていない観察対象の総数を取得（ページネーション用）
func (r *targetRepository) Count(ctx context.Context) (int, error) {
	// 論理削除されていないレコードのみをカウント
	query := `SELECT COUNT(*) FROM targets WHERE deleted_at IS NULL`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
