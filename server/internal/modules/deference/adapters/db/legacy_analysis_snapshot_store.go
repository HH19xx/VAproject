package db

import (
	"context"
	"database/sql"

	deferencedomain "server/internal/modules/deference/domain"
)

type legacyAnalysisSnapshotStore struct {
	db *sql.DB
}

func newLegacyAnalysisSnapshotStore(db *sql.DB) analysisSnapshotStore {
	return &legacyAnalysisSnapshotStore{db: db}
}

func (s *legacyAnalysisSnapshotStore) FindAllByUserID(ctx context.Context, userID, limit int) ([]*deferencedomain.AnalysisSnapshot, error) {
	query := `
		SELECT
			id,
			user_id,
			query_text,
			severity,
			score,
			delta_avg_tag,
			delta_var_tag,
			delta_prototype_rate,
			p_value,
			significant,
			current_count,
			current_avg_tag,
			current_var_tag,
			current_prototype_rate,
			baseline_count,
			baseline_avg_tag,
			baseline_var_tag,
			baseline_prototype_rate,
			created_at,
			create_user,
			deleted_at
		FROM analysis_snapshots
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`
	rows, err := s.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := make([]*deferencedomain.AnalysisSnapshot, 0)
	for rows.Next() {
		var snapshot deferencedomain.AnalysisSnapshot
		var baselineCount sql.NullInt64
		var baselineAvg sql.NullFloat64
		var baselineVar sql.NullFloat64
		var baselinePrototypeRate sql.NullFloat64
		if err := rows.Scan(
			&snapshot.ID,
			&snapshot.UserID,
			&snapshot.QueryText,
			&snapshot.Severity,
			&snapshot.Score,
			&snapshot.DeltaAvgTag,
			&snapshot.DeltaVarTag,
			&snapshot.DeltaPrototypeRate,
			&snapshot.PValue,
			&snapshot.Significant,
			&snapshot.Current.Count,
			&snapshot.Current.AvgTagCount,
			&snapshot.Current.VarTagCount,
			&snapshot.Current.PrototypeRate,
			&baselineCount,
			&baselineAvg,
			&baselineVar,
			&baselinePrototypeRate,
			&snapshot.CreatedAt,
			&snapshot.CreateUser,
			&snapshot.DeletedAt,
		); err != nil {
			return nil, err
		}

		if baselineCount.Valid || baselineAvg.Valid || baselineVar.Valid || baselinePrototypeRate.Valid {
			baseline := &deferencedomain.DistributionStats{}
			if baselineCount.Valid {
				baseline.Count = int(baselineCount.Int64)
			}
			if baselineAvg.Valid {
				baseline.AvgTagCount = baselineAvg.Float64
			}
			if baselineVar.Valid {
				baseline.VarTagCount = baselineVar.Float64
			}
			if baselinePrototypeRate.Valid {
				baseline.PrototypeRate = baselinePrototypeRate.Float64
			}
			snapshot.Baseline = baseline
		}
		snapshots = append(snapshots, &snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (s *legacyAnalysisSnapshotStore) Create(ctx context.Context, snapshot *deferencedomain.AnalysisSnapshot) (int, error) {
	query := `
		INSERT INTO analysis_snapshots (
			user_id, query_text, severity, score,
			delta_avg_tag, delta_var_tag, delta_prototype_rate,
			p_value, significant,
			current_count, current_avg_tag, current_var_tag, current_prototype_rate,
			baseline_count, baseline_avg_tag, baseline_var_tag, baseline_prototype_rate,
			create_user
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, $17,
			$18
		)
		RETURNING id, created_at
	`

	var baselineCount interface{}
	var baselineAvg interface{}
	var baselineVar interface{}
	var baselinePrototypeRate interface{}
	if snapshot.Baseline != nil {
		baselineCount = snapshot.Baseline.Count
		baselineAvg = snapshot.Baseline.AvgTagCount
		baselineVar = snapshot.Baseline.VarTagCount
		baselinePrototypeRate = snapshot.Baseline.PrototypeRate
	}

	var id int
	if err := s.db.QueryRowContext(
		ctx,
		query,
		snapshot.UserID,
		snapshot.QueryText,
		snapshot.Severity,
		snapshot.Score,
		snapshot.DeltaAvgTag,
		snapshot.DeltaVarTag,
		snapshot.DeltaPrototypeRate,
		snapshot.PValue,
		snapshot.Significant,
		snapshot.Current.Count,
		snapshot.Current.AvgTagCount,
		snapshot.Current.VarTagCount,
		snapshot.Current.PrototypeRate,
		baselineCount,
		baselineAvg,
		baselineVar,
		baselinePrototypeRate,
		snapshot.CreateUser,
	).Scan(&id, &snapshot.CreatedAt); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *legacyAnalysisSnapshotStore) DeleteAllByUserID(ctx context.Context, userID int) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE analysis_snapshots
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND deleted_at IS NULL
	`, userID)
	return err
}
