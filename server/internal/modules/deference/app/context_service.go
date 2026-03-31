package app

import (
	"context"
	"time"

	externalsignaldomain "server/internal/modules/externalsignal/domain"
	inferencedomain "server/internal/modules/inference/domain"
)

type ContextService struct {
	InferenceQuery     InferenceQueryPort
	ExternalSignalQuery ExternalSignalQueryPort
}

func (u *ContextService) Build(
	ctx context.Context,
	userID int,
	source string,
	locationKey string,
	signalType string,
	from, to time.Time,
	limit int,
) (*ContextResult, error) {
	if limit < 1 || limit > 500 {
		limit = 200
	}

	actionOptions := ActionLogQueryOptions{
		From:  &from,
		To:    &to,
		Sort:  "occurred_at",
		Order: "desc",
	}
	actionLogs, total, meta, err := u.InferenceQuery.GetActionLogs(ctx, userID, nil, 1, limit, actionOptions)
	if err != nil {
		return nil, err
	}

	if source == "" {
		source = "open_meteo"
	}
	worldSignals, err := u.ExternalSignalQuery.ListByRange(ctx, source, locationKey, signalType, from, to, limit)
	if err != nil {
		return nil, err
	}

	summary := buildAnalysisContextSummary(actionLogs, worldSignals)
	return &ContextResult{
		From:         from,
		To:           to,
		LocationKey:  locationKey,
		ActionLogs:   actionLogs,
		WorldSignals: worldSignals,
		Summary:      summary,
		Meta: map[string]interface{}{
			"action_total": total,
			"action_meta":  meta,
			"source":       source,
			"signal_type":  signalType,
		},
	}, nil
}

func buildAnalysisContextSummary(actionLogs []*inferencedomain.ActionLog, worldSignals []*externalsignaldomain.WorldSignal) ContextSummary {
	summary := ContextSummary{
		ActionCount: len(actionLogs),
		SignalCount: len(worldSignals),
	}

	if len(actionLogs) > 0 {
		sumTags := 0
		for _, log := range actionLogs {
			sumTags += len(log.TagIDs)
		}
		summary.AvgTagsPerAction = float64(sumTags) / float64(len(actionLogs))
	}

	if len(worldSignals) > 0 {
		tempCount := 0
		tempSum := 0.0
		precipSum := 0.0
		for _, signal := range worldSignals {
			if signal.TemperatureC != nil {
				tempSum += *signal.TemperatureC
				tempCount++
			}
			if signal.PrecipitationMM != nil {
				precipSum += *signal.PrecipitationMM
			}
		}
		signalValueCount := 0
		signalValueSum := 0.0
		for _, signal := range worldSignals {
			if signal.SignalValue != nil {
				signalValueSum += *signal.SignalValue
				signalValueCount++
			}
		}
		if tempCount > 0 {
			avg := tempSum / float64(tempCount)
			summary.AvgTemperatureC = &avg
		}
		if signalValueCount > 0 {
			avg := signalValueSum / float64(signalValueCount)
			summary.AvgSignalValue = &avg
		}
		summary.TotalPrecipitationM = precipSum
	}

	return summary
}
