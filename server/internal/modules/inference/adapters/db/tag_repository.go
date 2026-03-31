package db

import (
	"context"
	"database/sql"

	inferencedomain "server/internal/modules/inference/domain"
)

type TagRepository struct {
	inner tagStore
}

type tagStore interface {
	FindAll(ctx context.Context, userID int) ([]*inferencedomain.Tag, error)
	FindByID(ctx context.Context, id, userID int) (*inferencedomain.Tag, error)
	Create(ctx context.Context, tag *inferencedomain.Tag) (int, error)
	Update(ctx context.Context, tag *inferencedomain.Tag) error
	Delete(ctx context.Context, id, userID int) error
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{
		inner: newLegacyTagStore(db),
	}
}

func (r *TagRepository) FindAll(ctx context.Context, userID int) ([]*inferencedomain.Tag, error) {
	return r.inner.FindAll(ctx, userID)
}

func (r *TagRepository) FindByID(ctx context.Context, id, userID int) (*inferencedomain.Tag, error) {
	return r.inner.FindByID(ctx, id, userID)
}

func (r *TagRepository) Create(ctx context.Context, tag *inferencedomain.Tag) (int, error) {
	return r.inner.Create(ctx, tag)
}

func (r *TagRepository) Update(ctx context.Context, tag *inferencedomain.Tag) error {
	return r.inner.Update(ctx, tag)
}

func (r *TagRepository) Delete(ctx context.Context, id, userID int) error {
	return r.inner.Delete(ctx, id, userID)
}
