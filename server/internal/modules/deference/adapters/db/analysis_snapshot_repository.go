package db

import (
	"context"
	"database/sql"

	deferencedomain "server/internal/modules/deference/domain"
)

type AnalysisSnapshotRepository struct {
	inner analysisSnapshotStore
}

type analysisSnapshotStore interface {
	FindAllByUserID(ctx context.Context, userID, limit int) ([]*deferencedomain.AnalysisSnapshot, error)
	Create(ctx context.Context, snapshot *deferencedomain.AnalysisSnapshot) (int, error)
	DeleteAllByUserID(ctx context.Context, userID int) error
}

func NewAnalysisSnapshotRepository(db *sql.DB) *AnalysisSnapshotRepository {
	return &AnalysisSnapshotRepository{
		inner: newLegacyAnalysisSnapshotStore(db),
	}
}

func (r *AnalysisSnapshotRepository) FindAllByUserID(ctx context.Context, userID, limit int) ([]*deferencedomain.AnalysisSnapshot, error) {
	return r.inner.FindAllByUserID(ctx, userID, limit)
}

func (r *AnalysisSnapshotRepository) Create(ctx context.Context, snapshot *deferencedomain.AnalysisSnapshot) (int, error) {
	return r.inner.Create(ctx, snapshot)
}

func (r *AnalysisSnapshotRepository) DeleteAllByUserID(ctx context.Context, userID int) error {
	return r.inner.DeleteAllByUserID(ctx, userID)
}
