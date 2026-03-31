package db

import (
	"context"
	"database/sql"

	inferencedomain "server/internal/modules/inference/domain"
)

type TargetRepository struct {
	inner targetStore
}

type targetStore interface {
	FindAll(ctx context.Context, userID, limit, offset int) ([]*inferencedomain.Target, error)
	FindByID(ctx context.Context, id, userID int) (*inferencedomain.Target, error)
	Create(ctx context.Context, target *inferencedomain.Target) (int, error)
	Update(ctx context.Context, target *inferencedomain.Target) error
	Delete(ctx context.Context, id, userID int) error
	Count(ctx context.Context, userID int) (int, error)
	FindActionLogsByTargetID(ctx context.Context, targetID, userID, limit, offset int) ([]*inferencedomain.ActionLog, error)
	CountActionLogsByTargetID(ctx context.Context, targetID, userID int) (int, error)
}

func NewTargetRepository(db *sql.DB) *TargetRepository {
	return &TargetRepository{inner: newLegacyTargetStore(db)}
}

func (r *TargetRepository) FindAll(ctx context.Context, userID, limit, offset int) ([]*inferencedomain.Target, error) {
	return r.inner.FindAll(ctx, userID, limit, offset)
}

func (r *TargetRepository) FindByID(ctx context.Context, id, userID int) (*inferencedomain.Target, error) {
	return r.inner.FindByID(ctx, id, userID)
}

func (r *TargetRepository) Create(ctx context.Context, target *inferencedomain.Target) (int, error) {
	return r.inner.Create(ctx, target)
}

func (r *TargetRepository) Update(ctx context.Context, target *inferencedomain.Target) error {
	return r.inner.Update(ctx, target)
}

func (r *TargetRepository) Delete(ctx context.Context, id, userID int) error {
	return r.inner.Delete(ctx, id, userID)
}

func (r *TargetRepository) Count(ctx context.Context, userID int) (int, error) {
	return r.inner.Count(ctx, userID)
}

func (r *TargetRepository) FindActionLogsByTargetID(ctx context.Context, targetID, userID, limit, offset int) ([]*inferencedomain.ActionLog, error) {
	return r.inner.FindActionLogsByTargetID(ctx, targetID, userID, limit, offset)
}

func (r *TargetRepository) CountActionLogsByTargetID(ctx context.Context, targetID, userID int) (int, error) {
	return r.inner.CountActionLogsByTargetID(ctx, targetID, userID)
}

func cloneIntSlice(src []int) []int {
	if len(src) == 0 {
		return nil
	}

	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

func cloneIntMatrix(src [][]int) [][]int {
	if len(src) == 0 {
		return nil
	}

	dst := make([][]int, 0, len(src))
	for _, row := range src {
		dst = append(dst, cloneIntSlice(row))
	}

	return dst
}
