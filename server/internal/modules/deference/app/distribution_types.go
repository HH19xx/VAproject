package app

import inferenceapp "server/internal/modules/inference/app"

type DistributionDataset string

const (
	DistributionDatasetActionLogs  DistributionDataset = "action_logs"
	DistributionDatasetWorldSignal DistributionDataset = "world_signals"
)

type DistributionInput struct {
	Dataset          DistributionDataset
	Axis             string
	TagIDs           []int
	Options          inferenceapp.ActionLogListOptions
	Source           string
	LocationKey      string
	SignalType       string
	WorldSignalLimit int
}

type NumericDistributionStats struct {
	Count            int       `json:"count"`
	MissingCount     int       `json:"missing_count"`
	Mean             float64   `json:"mean"`
	Variance         float64   `json:"variance"`
	StdDev           float64   `json:"std_dev"`
	Min              float64   `json:"min"`
	Max              float64   `json:"max"`
	Skewness         float64   `json:"skewness"`
	ExcessKurtosis   float64   `json:"excess_kurtosis"`
	UniqueValueCount int       `json:"unique_value_count"`
	PeakCount        int       `json:"peak_count"`
	NormalityScore   float64   `json:"normality_score"`
	SampleValues     []float64 `json:"sample_values,omitempty"`
}

type DistributionIssue struct {
	Code       string `json:"code"`
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

type DistributionComparison struct {
	MeanDiff           float64 `json:"mean_diff"`
	VarianceDiff       float64 `json:"variance_diff"`
	StdDevDiff         float64 `json:"std_dev_diff"`
	SkewnessDiff       float64 `json:"skewness_diff"`
	NormalityScoreDiff float64 `json:"normality_score_diff"`
}

type DistributionBin struct {
	Start         float64 `json:"start"`
	End           float64 `json:"end"`
	Center        float64 `json:"center"`
	ObservedCount int     `json:"observed_count"`
	ExpectedCount float64 `json:"expected_count"`
	GapCount      float64 `json:"gap_count"`
}

type DistributionResult struct {
	Dataset                DistributionDataset       `json:"dataset"`
	Axis                   string                    `json:"axis"`
	Current                NumericDistributionStats  `json:"current"`
	ResidualCurrent        NumericDistributionStats  `json:"residual_current"`
	Baseline               *NumericDistributionStats `json:"baseline,omitempty"`
	ResidualBaseline       *NumericDistributionStats `json:"residual_baseline,omitempty"`
	Comparison             *DistributionComparison   `json:"comparison,omitempty"`
	Bins                   []DistributionBin         `json:"bins"`
	RawBins                []DistributionBin         `json:"raw_bins"`
	ResidualBins           []DistributionBin         `json:"residual_bins"`
	Issues                 []DistributionIssue       `json:"issues"`
	HiddenFactorCandidates []string                  `json:"hidden_factor_candidates"`
	FormatSuggestions      []string                  `json:"format_suggestions"`
	SuggestedActions       []string                  `json:"suggested_actions"`
	SuggestedTags          []string                  `json:"suggested_tags"`
	SuggestedAxes          []string                  `json:"suggested_axes"`
	Meta                   map[string]interface{}    `json:"meta"`
}
