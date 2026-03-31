package app

import (
	"context"
	"time"

	externalsignaldomain "server/internal/modules/externalsignal/domain"
	inferencedomain "server/internal/modules/inference/domain"
)

type InferenceQueryPort interface {
	GetActionLogs(
		ctx context.Context,
		actorUserID int,
		tagIDs []int,
		page, limit int,
		options ActionLogQueryOptions,
	) ([]*inferencedomain.ActionLog, int, ActionLogQueryMeta, error)
}

type ExternalSignalQueryPort interface {
	ListByRange(
		ctx context.Context,
		source, locationKey, signalType string,
		from, to time.Time,
		limit int,
	) ([]*externalsignaldomain.WorldSignal, error)
}

type DistributionAnalyzerPort interface {
	Analyze(
		ctx context.Context,
		userID int,
		input DistributionInput,
	) (*DistributionResult, error)
}
