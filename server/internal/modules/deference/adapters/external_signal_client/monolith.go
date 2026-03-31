package externalsignalclient

import (
	"context"
	"time"

	externalsignaldomain "server/internal/modules/externalsignal/domain"
	externalsignalapp "server/internal/modules/externalsignal/app"
)

type MonolithExternalSignalQuery struct {
	Service *externalsignalapp.QueryService
}

func (m *MonolithExternalSignalQuery) ListByRange(
	ctx context.Context,
	source, locationKey, signalType string,
	from, to time.Time,
	limit int,
) ([]*externalsignaldomain.WorldSignal, error) {
	return m.Service.ListByRange(ctx, source, locationKey, signalType, from, to, limit)
}
