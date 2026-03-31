package domain

import "time"

type DistributionStats struct {
	Count         int     `json:"count"`
	AvgTagCount   float64 `json:"avg_tag_count"`
	VarTagCount   float64 `json:"var_tag_count"`
	PrototypeRate float64 `json:"prototype_rate"`
}

type AnalysisSnapshot struct {
	ID                 int                `json:"id"`
	UserID             int                `json:"user_id"`
	QueryText          string             `json:"query_text"`
	Severity           string             `json:"severity"`
	Score              int                `json:"score"`
	DeltaAvgTag        float64            `json:"delta_avg_tag"`
	DeltaVarTag        float64            `json:"delta_var_tag"`
	DeltaPrototypeRate float64            `json:"delta_prototype_rate"`
	PValue             *float64           `json:"p_value,omitempty"`
	Significant        *bool              `json:"significant,omitempty"`
	Current            DistributionStats  `json:"current"`
	Baseline           *DistributionStats `json:"baseline,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	CreateUser         string             `json:"create_user"`
	DeletedAt          *time.Time         `json:"deleted_at,omitempty"`
}
