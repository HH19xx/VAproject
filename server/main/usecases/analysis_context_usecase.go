package usecases

import (
	"context"
	"time"

	"server/main/domain"
)

type AnalysisContextSummary struct {
	ActionCount         int      `json:"action_count"`
	SignalCount         int      `json:"signal_count"`
	AvgTagsPerAction    float64  `json:"avg_tags_per_action"`
	AvgTemperatureC     *float64 `json:"avg_temperature_c,omitempty"`
	TotalPrecipitationM float64  `json:"total_precipitation_mm"`
}

type AnalysisContextResult struct {
	From         time.Time              `json:"from"`
	To           time.Time              `json:"to"`
	LocationKey  string                 `json:"location_key"`
	ActionLogs   []*domain.ActionLog    `json:"action_logs"`
	WorldSignals []*domain.WorldSignal  `json:"world_signals"`
	Summary      AnalysisContextSummary `json:"summary"`
	Meta         map[string]interface{} `json:"meta"`
}

type AnalysisContextUsecase struct {
	ActionLogUsecase   *ActionLogUsecase
	WorldSignalUsecase *WorldSignalUsecase
}

func (u *AnalysisContextUsecase) Build(
	ctx context.Context,
	userID int,
	locationKey string,
	from, to time.Time,
	limit int,
) (*AnalysisContextResult, error) {
	if limit < 1 || limit > 500 {
		limit = 200
	}

	actionOptions := ActionLogListOptions{
		From:  &from,
		To:    &to,
		Sort:  "occurred_at",
		Order: "desc",
	}
	actionLogs, total, meta, err := u.ActionLogUsecase.GetActionLogs(ctx, userID, nil, 1, limit, actionOptions)
	if err != nil {
		return nil, err
	}

	worldSignals, err := u.WorldSignalUsecase.ListByRange(ctx, "open_meteo", locationKey, from, to, limit)
	if err != nil {
		return nil, err
	}

	summary := buildAnalysisContextSummary(actionLogs, worldSignals)
	return &AnalysisContextResult{
		From:         from,
		To:           to,
		LocationKey:  locationKey,
		ActionLogs:   actionLogs,
		WorldSignals: worldSignals,
		Summary:      summary,
		Meta: map[string]interface{}{
			"action_total": total,
			"action_meta":  meta,
		},
	}, nil
}

func buildAnalysisContextSummary(actionLogs []*domain.ActionLog, worldSignals []*domain.WorldSignal) AnalysisContextSummary {
	summary := AnalysisContextSummary{
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
		if tempCount > 0 {
			avg := tempSum / float64(tempCount)
			summary.AvgTemperatureC = &avg
		}
		summary.TotalPrecipitationM = precipSum
	}

	return summary
}
