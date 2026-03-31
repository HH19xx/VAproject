package db

import (
	"context"
	"database/sql"
	"time"

	externalsignaldomain "server/internal/modules/externalsignal/domain"
)

type WorldSignalRepository struct {
	inner worldSignalStore
}

type worldSignalStore interface {
	UpsertMany(ctx context.Context, signals []*externalsignaldomain.WorldSignal) (int, error)
	FindByRange(ctx context.Context, source, locationKey, signalType string, from, to time.Time, limit int) ([]*externalsignaldomain.WorldSignal, error)
}

func NewWorldSignalRepository(db *sql.DB) *WorldSignalRepository {
	return &WorldSignalRepository{
		inner: newLegacyWorldSignalStore(db),
	}
}

func (r *WorldSignalRepository) UpsertMany(ctx context.Context, signals []*externalsignaldomain.WorldSignal) (int, error) {
	return r.inner.UpsertMany(ctx, signals)
}

func (r *WorldSignalRepository) FindByRange(
	ctx context.Context,
	source, locationKey, signalType string,
	from, to time.Time,
	limit int,
) ([]*externalsignaldomain.WorldSignal, error) {
	return r.inner.FindByRange(ctx, source, locationKey, signalType, from, to, limit)
}
