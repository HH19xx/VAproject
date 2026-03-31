package db

import (
	"context"
	"database/sql"

	inferenceapp "server/internal/modules/inference/app"
	inferencedomain "server/internal/modules/inference/domain"
)

type ActionLogRepository struct {
	inner actionLogStore
}

type actionLogStore interface {
	FindAll(ctx context.Context, userID int, tagIDs []int, limit, offset int, options inferenceapp.ActionLogListOptions) ([]*inferencedomain.ActionLog, error)
	Count(ctx context.Context, userID int, tagIDs []int, options inferenceapp.ActionLogListOptions) (int, error)
	FindByID(ctx context.Context, id, userID int) (*inferencedomain.ActionLog, error)
	ListAttributeKeys(ctx context.Context, userID int) ([]string, error)
	Create(ctx context.Context, log *inferencedomain.ActionLog) (int, error)
	Update(ctx context.Context, log *inferencedomain.ActionLog) error
	Delete(ctx context.Context, id, userID int) error
}

func NewActionLogRepository(db *sql.DB) *ActionLogRepository {
	return &ActionLogRepository{
		inner: newLegacyActionLogStore(db),
	}
}

func (r *ActionLogRepository) FindAll(
	ctx context.Context,
	userID int,
	tagIDs []int,
	limit, offset int,
	options inferenceapp.ActionLogListOptions,
) ([]*inferencedomain.ActionLog, error) {
	return r.inner.FindAll(ctx, userID, tagIDs, limit, offset, options)
}

func (r *ActionLogRepository) Count(
	ctx context.Context,
	userID int,
	tagIDs []int,
	options inferenceapp.ActionLogListOptions,
) (int, error) {
	return r.inner.Count(ctx, userID, tagIDs, options)
}

func (r *ActionLogRepository) FindByID(ctx context.Context, id, userID int) (*inferencedomain.ActionLog, error) {
	return r.inner.FindByID(ctx, id, userID)
}

func (r *ActionLogRepository) ListAttributeKeys(ctx context.Context, userID int) ([]string, error) {
	return r.inner.ListAttributeKeys(ctx, userID)
}

func (r *ActionLogRepository) Create(ctx context.Context, log *inferencedomain.ActionLog) (int, error) {
	return r.inner.Create(ctx, log)
}

func (r *ActionLogRepository) Update(ctx context.Context, log *inferencedomain.ActionLog) error {
	return r.inner.Update(ctx, log)
}

func (r *ActionLogRepository) Delete(ctx context.Context, id, userID int) error {
	return r.inner.Delete(ctx, id, userID)
}
