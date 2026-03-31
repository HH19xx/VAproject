package app

import (
	"time"

	externalsignaldomain "server/internal/modules/externalsignal/domain"
	inferencedomain "server/internal/modules/inference/domain"
)

type ActionLogQueryOptions struct {
	From          *time.Time
	To            *time.Time
	Sort          string
	Order         string
	AnyTagIDs     []int
	AnyTagGroups  [][]int
	ExcludeTagIDs []int
}

type ActionLogQueryMeta struct {
	From              *time.Time `json:"from,omitempty"`
	To                *time.Time `json:"to,omitempty"`
	Sort              string     `json:"sort"`
	Order             string     `json:"order"`
	UsedTagIDs        []int      `json:"used_tag_ids"`
	UsedAnyTagIDs     []int      `json:"used_any_tag_ids"`
	UsedAnyTagGroups  [][]int    `json:"used_any_tag_groups"`
	UsedExcludeTagIDs []int      `json:"used_exclude_tag_ids"`
	AxisCandidates    []string   `json:"axis_candidates"`
}

type ContextSummary struct {
	ActionCount         int      `json:"action_count"`
	SignalCount         int      `json:"signal_count"`
	AvgTagsPerAction    float64  `json:"avg_tags_per_action"`
	AvgTemperatureC     *float64 `json:"avg_temperature_c,omitempty"`
	AvgSignalValue      *float64 `json:"avg_signal_value,omitempty"`
	TotalPrecipitationM float64  `json:"total_precipitation_mm"`
}

type ContextResult struct {
	From         time.Time                         `json:"from"`
	To           time.Time                         `json:"to"`
	LocationKey  string                            `json:"location_key"`
	ActionLogs   []*inferencedomain.ActionLog      `json:"action_logs"`
	WorldSignals []*externalsignaldomain.WorldSignal `json:"world_signals"`
	Summary      ContextSummary                    `json:"summary"`
	Meta         map[string]interface{}            `json:"meta"`
}
