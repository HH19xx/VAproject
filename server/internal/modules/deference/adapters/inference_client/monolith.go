package inferenceclient

import (
	"context"

	deferenceapp "server/internal/modules/deference/app"
	inferenceapp "server/internal/modules/inference/app"
	inferencedomain "server/internal/modules/inference/domain"
)

type MonolithInferenceQuery struct {
	Service *inferenceapp.ActionLogService
}

func (m *MonolithInferenceQuery) GetActionLogs(
	ctx context.Context,
	actorUserID int,
	tagIDs []int,
	page, limit int,
	options deferenceapp.ActionLogQueryOptions,
) ([]*inferencedomain.ActionLog, int, deferenceapp.ActionLogQueryMeta, error) {
	logs, total, meta, err := m.Service.GetActionLogs(ctx, actorUserID, tagIDs, page, limit, inferenceapp.ActionLogListOptions{
		From:          options.From,
		To:            options.To,
		Sort:          options.Sort,
		Order:         options.Order,
		AnyTagIDs:     options.AnyTagIDs,
		AnyTagGroups:  options.AnyTagGroups,
		ExcludeTagIDs: options.ExcludeTagIDs,
	})
	if err != nil {
		return nil, 0, deferenceapp.ActionLogQueryMeta{}, err
	}
	return logs, total, deferenceapp.ActionLogQueryMeta{
		From:              meta.From,
		To:                meta.To,
		Sort:              meta.Sort,
		Order:             meta.Order,
		UsedTagIDs:        meta.UsedTagIDs,
		UsedAnyTagIDs:     meta.UsedAnyTagIDs,
		UsedAnyTagGroups:  meta.UsedAnyTagGroups,
		UsedExcludeTagIDs: meta.UsedExcludeTagIDs,
		AxisCandidates:    meta.AxisCandidates,
	}, nil
}
