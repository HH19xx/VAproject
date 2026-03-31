package db

import (
	"context"
	"database/sql"

	inferencedomain "server/internal/modules/inference/domain"
)

type PrototypeRepository struct {
	inner prototypeStore
}

type prototypeStore interface {
	FindAll(ctx context.Context, userID int) ([]*inferencedomain.Prototype, error)
	FindByID(ctx context.Context, id, userID int) (*inferencedomain.Prototype, error)
	Create(ctx context.Context, prototype *inferencedomain.Prototype) (int, error)
	Update(ctx context.Context, prototype *inferencedomain.Prototype) error
	Delete(ctx context.Context, id, userID int) error
}

func NewPrototypeRepository(db *sql.DB) *PrototypeRepository {
	return &PrototypeRepository{
		inner: newLegacyPrototypeStore(db),
	}
}

func (r *PrototypeRepository) FindAll(ctx context.Context, userID int) ([]*inferencedomain.Prototype, error) {
	return r.inner.FindAll(ctx, userID)
}

func (r *PrototypeRepository) FindByID(ctx context.Context, id, userID int) (*inferencedomain.Prototype, error) {
	return r.inner.FindByID(ctx, id, userID)
}

func (r *PrototypeRepository) Create(ctx context.Context, prototype *inferencedomain.Prototype) (int, error) {
	return r.inner.Create(ctx, prototype)
}

func (r *PrototypeRepository) Update(ctx context.Context, prototype *inferencedomain.Prototype) error {
	return r.inner.Update(ctx, prototype)
}

func (r *PrototypeRepository) Delete(ctx context.Context, id, userID int) error {
	return r.inner.Delete(ctx, id, userID)
}
