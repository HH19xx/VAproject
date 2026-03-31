package distributionclient

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	deferenceapp "server/internal/modules/deference/app"
	externalsignalapp "server/internal/modules/externalsignal/app"
	externalsignaldomain "server/internal/modules/externalsignal/domain"
	inferenceapp "server/internal/modules/inference/app"
	inferencedomain "server/internal/modules/inference/domain"
)

type MonolithDistributionAnalyzer struct {
	ActionLogService   *inferenceapp.ActionLogService
	WorldSignalService *externalsignalapp.QueryService
}

type previousWindow struct {
	From time.Time
	To   time.Time
}

type residualStrategy struct {
	Name       string
	WindowSize int
}

type axisFamilyDecision struct {
	Family string
	Reason string
}

func NewMonolithDistributionAnalyzer(
	actionLogService *inferenceapp.ActionLogService,
	worldSignalService *externalsignalapp.QueryService,
) *MonolithDistributionAnalyzer {
	return &MonolithDistributionAnalyzer{
		ActionLogService:   actionLogService,
		WorldSignalService: worldSignalService,
	}
}

func (m *MonolithDistributionAnalyzer) Analyze(
	ctx context.Context,
	userID int,
	input deferenceapp.DistributionInput,
) (*deferenceapp.DistributionResult, error) {
	axis := strings.TrimSpace(input.Axis)
	if axis == "" {
		return nil, deferenceapp.ErrDistributionAxisRequired
	}

	switch input.Dataset {
	case deferenceapp.DistributionDatasetActionLogs:
		return m.analyzeActionLogs(ctx, userID, axis, input.TagIDs, input.Options)
	case deferenceapp.DistributionDatasetWorldSignal:
		return m.analyzeWorldSignals(ctx, axis, input.Source, input.LocationKey, input.SignalType, input.Options, input.WorldSignalLimit)
	default:
		return nil, deferenceapp.ErrDistributionDatasetInvalid
	}
}

func (m *MonolithDistributionAnalyzer) analyzeActionLogs(
	ctx context.Context,
	userID int,
	axis string,
	tagIDs []int,
	options inferenceapp.ActionLogListOptions,
) (*deferenceapp.DistributionResult, error) {
	if !isValidActionAxis(axis) {
		return nil, deferenceapp.ErrDistributionAxisInvalid
	}

	logs, total, meta, err := m.fetchAllActionLogs(ctx, userID, tagIDs, options, 1000)
	if err != nil {
		return nil, err
	}

	currentValues, currentMissing := collectActionAxisValues(logs, axis)
	currentStats := buildNumericDistributionStats(currentValues, currentMissing)
	currentAxisDecision := classifyActionAxisFamily(axis, currentStats)
	residualPlan := resolveResidualStrategy(currentAxisDecision.Family, len(currentValues), currentStats.UniqueValueCount)
	residualValues := buildResidualValues(currentValues, residualPlan)
	residualStats := buildNumericDistributionStats(residualValues, 0)
	rawBins := buildDistributionBins(currentValues, currentStats)
	residualBins := buildDistributionBins(residualValues, residualStats)

	var baselineStats *deferenceapp.NumericDistributionStats
	var residualBaselineStats *deferenceapp.NumericDistributionStats
	var comparison *deferenceapp.DistributionComparison
	baselineWindow := derivePreviousWindow(options.From, options.To)
	if baselineWindow != nil {
		baselineOptions := options
		baselineOptions.From = &baselineWindow.From
		baselineOptions.To = &baselineWindow.To

		baselineLogs, _, _, err := m.fetchAllActionLogs(ctx, userID, tagIDs, baselineOptions, 1000)
		if err != nil {
			return nil, err
		}
		baselineValues, baselineMissing := collectActionAxisValues(baselineLogs, axis)
		stats := buildNumericDistributionStats(baselineValues, baselineMissing)
		baselineStats = &stats
		baselineAxisDecision := classifyActionAxisFamily(axis, stats)
		baselineResidualPlan := resolveResidualStrategy(baselineAxisDecision.Family, len(baselineValues), stats.UniqueValueCount)
		baselineResidualValues := buildResidualValues(baselineValues, baselineResidualPlan)
		baselineResidualStatsValue := buildNumericDistributionStats(baselineResidualValues, 0)
		residualBaselineStats = &baselineResidualStatsValue
		comparison = buildDistributionComparison(residualStats, baselineResidualStatsValue)
	}

	issues := append(buildAxisDecisionIssues(currentAxisDecision, currentStats), inferDistributionIssues(residualStats, residualBaselineStats, axis)...)
	hiddenFactors := summarizeIssueTexts(issues, "hidden_factor")
	formatSuggestions := summarizeSuggestedActions(issues)
	suggestedTags := suggestTagsForDistributionIssues(deferenceapp.DistributionDatasetActionLogs, axis, issues)
	suggestedAxes := suggestAxesForDistributionIssues(deferenceapp.DistributionDatasetActionLogs, axis, issues)

	return &deferenceapp.DistributionResult{
		Dataset:                deferenceapp.DistributionDatasetActionLogs,
		Axis:                   axis,
		Current:                currentStats,
		ResidualCurrent:        residualStats,
		Baseline:               baselineStats,
		ResidualBaseline:       residualBaselineStats,
		Comparison:             comparison,
		Bins:                   residualBins,
		RawBins:                rawBins,
		ResidualBins:           residualBins,
		Issues:                 issues,
		HiddenFactorCandidates: hiddenFactors,
		FormatSuggestions:      formatSuggestions,
		SuggestedActions:       formatSuggestions,
		SuggestedTags:          suggestedTags,
		SuggestedAxes:          suggestedAxes,
		Meta: map[string]interface{}{
			"total_logs":           total,
			"search_meta":          meta,
			"residual_strategy":    residualPlan.Name,
			"residual_window":      residualPlan.WindowSize,
			"residual_axis_type":   currentAxisDecision.Family,
			"residual_axis_reason": currentAxisDecision.Reason,
			"unique_value_count":   currentStats.UniqueValueCount,
			"unique_ratio":         safeUniqueRatio(currentStats),
			"sample_count":         currentStats.Count,
		},
	}, nil
}

func (m *MonolithDistributionAnalyzer) analyzeWorldSignals(
	ctx context.Context,
	axis, source, locationKey, signalType string,
	options inferenceapp.ActionLogListOptions,
	limit int,
) (*deferenceapp.DistributionResult, error) {
	if !isValidWorldSignalAxis(axis) {
		return nil, deferenceapp.ErrDistributionAxisInvalid
	}
	if strings.TrimSpace(locationKey) == "" {
		return nil, deferenceapp.ErrDistributionLocationNeeded
	}
	if options.From == nil || options.To == nil || options.From.IsZero() || options.To.IsZero() || options.From.After(*options.To) {
		return nil, deferenceapp.ErrDistributionWindowInvalid
	}
	if limit < 1 || limit > 1000 {
		limit = 1000
	}
	if strings.TrimSpace(source) == "" {
		source = "open_meteo"
	}

	signals, err := m.WorldSignalService.ListByRange(ctx, source, locationKey, signalType, *options.From, *options.To, limit)
	if err != nil {
		return nil, err
	}
	currentValues, currentMissing := collectWorldSignalValues(signals, axis)
	currentStats := buildNumericDistributionStats(currentValues, currentMissing)
	currentAxisDecision := classifyWorldSignalAxisFamily(axis, currentStats)
	residualPlan := resolveResidualStrategy(currentAxisDecision.Family, len(currentValues), currentStats.UniqueValueCount)
	residualValues := buildResidualValues(currentValues, residualPlan)
	residualStats := buildNumericDistributionStats(residualValues, 0)
	rawBins := buildDistributionBins(currentValues, currentStats)
	residualBins := buildDistributionBins(residualValues, residualStats)

	var baselineStats *deferenceapp.NumericDistributionStats
	var residualBaselineStats *deferenceapp.NumericDistributionStats
	var comparison *deferenceapp.DistributionComparison
	baselineWindow := derivePreviousWindow(options.From, options.To)
	if baselineWindow != nil {
		baselineSignals, err := m.WorldSignalService.ListByRange(ctx, source, locationKey, signalType, baselineWindow.From, baselineWindow.To, limit)
		if err != nil {
			return nil, err
		}
		baselineValues, baselineMissing := collectWorldSignalValues(baselineSignals, axis)
		stats := buildNumericDistributionStats(baselineValues, baselineMissing)
		baselineStats = &stats
		baselineAxisDecision := classifyWorldSignalAxisFamily(axis, stats)
		baselineResidualPlan := resolveResidualStrategy(baselineAxisDecision.Family, len(baselineValues), stats.UniqueValueCount)
		baselineResidualValues := buildResidualValues(baselineValues, baselineResidualPlan)
		baselineResidualStatsValue := buildNumericDistributionStats(baselineResidualValues, 0)
		residualBaselineStats = &baselineResidualStatsValue
		comparison = buildDistributionComparison(residualStats, baselineResidualStatsValue)
	}

	issues := append(buildAxisDecisionIssues(currentAxisDecision, currentStats), inferDistributionIssues(residualStats, residualBaselineStats, axis)...)
	hiddenFactors := summarizeIssueTexts(issues, "hidden_factor")
	formatSuggestions := summarizeSuggestedActions(issues)
	suggestedTags := suggestTagsForDistributionIssues(deferenceapp.DistributionDatasetWorldSignal, axis, issues)
	suggestedAxes := suggestAxesForDistributionIssues(deferenceapp.DistributionDatasetWorldSignal, axis, issues)

	return &deferenceapp.DistributionResult{
		Dataset:                deferenceapp.DistributionDatasetWorldSignal,
		Axis:                   axis,
		Current:                currentStats,
		ResidualCurrent:        residualStats,
		Baseline:               baselineStats,
		ResidualBaseline:       residualBaselineStats,
		Comparison:             comparison,
		Bins:                   residualBins,
		RawBins:                rawBins,
		ResidualBins:           residualBins,
		Issues:                 issues,
		HiddenFactorCandidates: hiddenFactors,
		FormatSuggestions:      formatSuggestions,
		SuggestedActions:       formatSuggestions,
		SuggestedTags:          suggestedTags,
		SuggestedAxes:          suggestedAxes,
		Meta: map[string]interface{}{
			"source":               source,
			"location_key":         locationKey,
			"signal_type":          signalType,
			"signal_count":         len(signals),
			"residual_strategy":    residualPlan.Name,
			"residual_window":      residualPlan.WindowSize,
			"residual_axis_type":   currentAxisDecision.Family,
			"residual_axis_reason": currentAxisDecision.Reason,
			"unique_value_count":   currentStats.UniqueValueCount,
			"unique_ratio":         safeUniqueRatio(currentStats),
			"sample_count":         currentStats.Count,
		},
	}, nil
}

func (m *MonolithDistributionAnalyzer) fetchAllActionLogs(
	ctx context.Context,
	userID int,
	tagIDs []int,
	options inferenceapp.ActionLogListOptions,
	maxCount int,
) ([]*inferencedomain.ActionLog, int, inferenceapp.ActionLogListMeta, error) {
	page := 1
	limit := 100
	all := make([]*inferencedomain.ActionLog, 0, minAnalysisInt(maxCount, 100))
	var total int
	var meta inferenceapp.ActionLogListMeta

	for {
		logs, count, listMeta, err := m.ActionLogService.GetActionLogs(ctx, userID, tagIDs, page, limit, options)
		if err != nil {
			return nil, 0, meta, err
		}
		total = count
		meta = listMeta
		all = append(all, logs...)
		if len(logs) < limit || len(all) >= maxCount || len(all) >= total {
			break
		}
		page++
	}

	if len(all) > maxCount {
		all = all[:maxCount]
	}
	return all, total, meta, nil
}

func derivePreviousWindow(from, to *time.Time) *previousWindow {
	if from == nil || to == nil || from.IsZero() || to.IsZero() || !from.Before(*to) {
		return nil
	}
	duration := to.Sub(*from)
	return &previousWindow{
		From: from.Add(-duration),
		To:   *from,
	}
}

func isValidActionAxis(axis string) bool {
	switch axis {
	case "occurred_at", "created_at", "updated_at", "tag_count", "prototype_id":
		return true
	default:
		return strings.TrimSpace(axis) != ""
	}
}

func isValidWorldSignalAxis(axis string) bool {
	switch axis {
	case "temperature_c", "precipitation_mm", "wind_speed_ms", "weather_code", "signal_value":
		return true
	default:
		return false
	}
}

func collectActionAxisValues(logs []*inferencedomain.ActionLog, axis string) ([]float64, int) {
	values := make([]float64, 0, len(logs))
	missing := 0
	for _, log := range logs {
		value, ok := actionAxisValue(log, axis)
		if !ok {
			missing++
			continue
		}
		values = append(values, value)
	}
	return values, missing
}

func actionAxisValue(log *inferencedomain.ActionLog, axis string) (float64, bool) {
	switch axis {
	case "occurred_at":
		return float64(log.OccurredAt.Unix()), true
	case "created_at":
		return float64(log.CreatedAt.Unix()), true
	case "updated_at":
		return float64(log.UpdatedAt.Unix()), true
	case "tag_count":
		return float64(len(log.TagIDs)), true
	case "prototype_id":
		if log.PrototypeID == nil {
			return 0, false
		}
		return float64(*log.PrototypeID), true
	default:
		for _, attr := range log.Attributes {
			if attr.Key == axis && attr.ValueNumber != nil && !math.IsNaN(*attr.ValueNumber) && !math.IsInf(*attr.ValueNumber, 0) {
				return *attr.ValueNumber, true
			}
		}
		return 0, false
	}
}

func collectWorldSignalValues(signals []*externalsignaldomain.WorldSignal, axis string) ([]float64, int) {
	values := make([]float64, 0, len(signals))
	missing := 0
	for _, signal := range signals {
		value, ok := worldSignalAxisValue(signal, axis)
		if !ok {
			missing++
			continue
		}
		values = append(values, value)
	}
	return values, missing
}

func worldSignalAxisValue(signal *externalsignaldomain.WorldSignal, axis string) (float64, bool) {
	switch axis {
	case "temperature_c":
		if signal.TemperatureC == nil {
			return 0, false
		}
		return *signal.TemperatureC, true
	case "precipitation_mm":
		if signal.PrecipitationMM == nil {
			return 0, false
		}
		return *signal.PrecipitationMM, true
	case "wind_speed_ms":
		if signal.WindSpeedMS == nil {
			return 0, false
		}
		return *signal.WindSpeedMS, true
	case "weather_code":
		if signal.WeatherCode == nil {
			return 0, false
		}
		return float64(*signal.WeatherCode), true
	case "signal_value":
		if signal.SignalValue == nil {
			return 0, false
		}
		return *signal.SignalValue, true
	default:
		return 0, false
	}
}

func buildNumericDistributionStats(values []float64, missingCount int) deferenceapp.NumericDistributionStats {
	stats := deferenceapp.NumericDistributionStats{
		Count:        len(values),
		MissingCount: missingCount,
	}
	if len(values) == 0 {
		return stats
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	stats.Min = sorted[0]
	stats.Max = sorted[len(sorted)-1]
	stats.UniqueValueCount = countUniqueValues(sorted)

	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	stats.Mean = mean

	variance := 0.0
	for _, value := range values {
		diff := value - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	stats.Variance = variance
	stats.StdDev = math.Sqrt(math.Max(variance, 0))

	if stats.StdDev > 0 && len(values) >= 3 {
		m3 := 0.0
		m4 := 0.0
		for _, value := range values {
			z := (value - mean) / stats.StdDev
			m3 += math.Pow(z, 3)
			m4 += math.Pow(z, 4)
		}
		m3 /= float64(len(values))
		m4 /= float64(len(values))
		stats.Skewness = m3
		stats.ExcessKurtosis = m4 - 3
	}

	stats.PeakCount = estimatePeakCount(sorted)
	stats.NormalityScore = estimateNormalityScore(stats)
	if len(sorted) > 12 {
		stats.SampleValues = append([]float64(nil), sorted[:12]...)
	} else {
		stats.SampleValues = append([]float64(nil), sorted...)
	}
	return stats
}

func buildResidualValues(values []float64, strategy residualStrategy) []float64 {
	if len(values) == 0 {
		return []float64{}
	}
	if strategy.Name == "standardized_discrete_residuals" {
		return buildStandardizedDiscreteResiduals(values)
	}
	if strategy.Name == "local_expected_continuous" {
		return buildLocalExpectedResiduals(values, strategy.WindowSize)
	}

	windowSize := strategy.WindowSize
	if windowSize < 2 {
		windowSize = 2
	}

	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))

	residuals := make([]float64, 0, int(math.Ceil(float64(len(values))/float64(windowSize))))
	for start := 0; start < len(values); start += windowSize {
		end := start + windowSize
		if end > len(values) {
			end = len(values)
		}
		if end-start < 2 {
			break
		}
		groupMean := 0.0
		for _, value := range values[start:end] {
			groupMean += value
		}
		groupMean /= float64(end - start)
		residuals = append(residuals, groupMean-mean)
	}

	if len(residuals) == 0 {
		return []float64{0}
	}
	return residuals
}

func buildLocalExpectedResiduals(values []float64, windowSize int) []float64 {
	if len(values) == 0 {
		return []float64{}
	}
	if windowSize < 3 {
		windowSize = 3
	}
	if windowSize%2 == 0 {
		windowSize++
	}

	radius := windowSize / 2
	residuals := make([]float64, 0, len(values))
	for index, value := range values {
		start := index - radius
		if start < 0 {
			start = 0
		}
		end := index + radius + 1
		if end > len(values) {
			end = len(values)
		}

		sum := 0.0
		count := 0
		for i := start; i < end; i++ {
			if i == index {
				continue
			}
			sum += values[i]
			count++
		}
		if count == 0 {
			residuals = append(residuals, 0)
			continue
		}
		expected := sum / float64(count)
		residuals = append(residuals, value-expected)
	}
	return residuals
}

func buildStandardizedDiscreteResiduals(values []float64) []float64 {
	if len(values) == 0 {
		return []float64{}
	}

	counts := make(map[float64]int)
	ordered := make([]float64, 0, len(values))
	for _, value := range values {
		key := roundValue(value)
		if _, exists := counts[key]; !exists {
			ordered = append(ordered, key)
		}
		counts[key]++
	}
	sort.Float64s(ordered)

	categoryCount := len(ordered)
	if categoryCount == 0 {
		return []float64{0}
	}

	expected := float64(len(values)) / float64(categoryCount)
	if expected <= 0 {
		return []float64{0}
	}

	residuals := make([]float64, 0, categoryCount)
	for _, key := range ordered {
		observed := float64(counts[key])
		residual := (observed - expected) / math.Sqrt(expected)
		residuals = append(residuals, residual)
	}
	if len(residuals) == 0 {
		return []float64{0}
	}
	return residuals
}

func inputAxisFamilyAction(axis string) string {
	switch axis {
	case "occurred_at", "created_at", "updated_at":
		return "time"
	case "tag_count", "prototype_id":
		return "discrete"
	default:
		return "continuous"
	}
}

func classifyActionAxisFamily(axis string, stats deferenceapp.NumericDistributionStats) axisFamilyDecision {
	baseFamily := inputAxisFamilyAction(axis)
	if baseFamily != "continuous" {
		return axisFamilyDecision{Family: baseFamily, Reason: "built_in_axis"}
	}
	if stats.Count == 0 {
		return axisFamilyDecision{Family: "continuous", Reason: "empty_sample"}
	}
	if stats.UniqueValueCount <= 6 {
		return axisFamilyDecision{Family: "discrete", Reason: "unique_value_count<=6"}
	}
	uniqueRatio := float64(stats.UniqueValueCount) / float64(stats.Count)
	if stats.Count >= 12 && uniqueRatio <= 0.35 {
		return axisFamilyDecision{Family: "discrete", Reason: "unique_ratio<=0.35"}
	}
	if stats.Count >= 24 && uniqueRatio <= 0.50 && stats.UniqueValueCount <= 12 {
		return axisFamilyDecision{Family: "discrete", Reason: "unique_ratio<=0.50_and_unique<=12"}
	}
	return axisFamilyDecision{Family: "continuous", Reason: "default_continuous"}
}

func inputAxisFamilyWorldSignal(axis string) string {
	switch axis {
	case "weather_code":
		return "discrete"
	default:
		return "continuous"
	}
}

func classifyWorldSignalAxisFamily(axis string, stats deferenceapp.NumericDistributionStats) axisFamilyDecision {
	baseFamily := inputAxisFamilyWorldSignal(axis)
	if baseFamily != "continuous" {
		return axisFamilyDecision{Family: baseFamily, Reason: "built_in_axis"}
	}
	if stats.Count == 0 {
		return axisFamilyDecision{Family: "continuous", Reason: "empty_sample"}
	}
	return axisFamilyDecision{Family: "continuous", Reason: "default_continuous"}
}

func resolveResidualStrategy(axisFamily string, sampleCount int, uniqueValueCount int) residualStrategy {
	if sampleCount <= 0 {
		return residualStrategy{Name: "none", WindowSize: 2}
	}
	switch axisFamily {
	case "time":
		window := int(math.Round(math.Sqrt(float64(sampleCount))))
		if window < 6 {
			window = 6
		}
		if window > 24 {
			window = 24
		}
		return residualStrategy{Name: "rolling_mean_time", WindowSize: window}
	case "discrete":
		window := uniqueValueCount
		if window < 2 {
			window = 2
		}
		return residualStrategy{Name: "standardized_discrete_residuals", WindowSize: window}
	default:
		window := int(math.Round(math.Sqrt(float64(sampleCount))))
		if window < 5 {
			window = 5
		}
		if window > 21 {
			window = 21
		}
		if window%2 == 0 {
			window++
		}
		return residualStrategy{Name: "local_expected_continuous", WindowSize: window}
	}
}

func safeUniqueRatio(stats deferenceapp.NumericDistributionStats) float64 {
	if stats.Count <= 0 {
		return 0
	}
	return float64(stats.UniqueValueCount) / float64(stats.Count)
}

func buildAxisDecisionIssues(decision axisFamilyDecision, stats deferenceapp.NumericDistributionStats) []deferenceapp.DistributionIssue {
	if stats.Count == 0 {
		return nil
	}

	severity := "ok"
	category := "axis_decision"
	message := "軸の分布特性を評価しました。"
	suggestion := ""

	switch decision.Reason {
	case "empty_sample":
		severity = "notice"
		category = "format_gap"
		message = "分析対象のサンプル数が不足しています。"
		suggestion = "検索条件か期間を広げてサンプル数を増やしてください。"
	case "unique_value_count<=6", "unique_ratio<=0.35", "unique_ratio<=0.50_and_unique<=12":
		severity = "notice"
		category = "format_gap"
		message = "この軸は連続量というより離散カテゴリに近い分布です。"
		suggestion = "より細かい軸を追加するか、別の連続値属性も併せて記録してください。"
	case "default_continuous":
		message = "この軸は連続量として扱うのが妥当です。"
	case "built_in_axis":
		message = "この軸は定義済みの特性に基づいて扱われています。"
	}

	return []deferenceapp.DistributionIssue{{
		Code:       "axis_family_decision",
		Category:   category,
		Severity:   severity,
		Message:    message,
		Suggestion: suggestion,
	}}
}

func countUniqueValues(sorted []float64) int {
	if len(sorted) == 0 {
		return 0
	}
	count := 1
	prev := roundValue(sorted[0])
	for i := 1; i < len(sorted); i++ {
		current := roundValue(sorted[i])
		if current != prev {
			count++
			prev = current
		}
	}
	return count
}

func roundValue(value float64) float64 {
	return math.Round(value*1000) / 1000
}

func estimatePeakCount(sorted []float64) int {
	if len(sorted) < 5 || sorted[0] == sorted[len(sorted)-1] {
		return 1
	}

	binCount := int(math.Round(math.Sqrt(float64(len(sorted)))))
	if binCount < 5 {
		binCount = 5
	}
	if binCount > 12 {
		binCount = 12
	}
	binWidth := (sorted[len(sorted)-1] - sorted[0]) / float64(binCount)
	if binWidth <= 0 {
		return 1
	}

	histogram := make([]int, binCount)
	for _, value := range sorted {
		idx := int(math.Floor((value - sorted[0]) / binWidth))
		if idx >= binCount {
			idx = binCount - 1
		}
		if idx < 0 {
			idx = 0
		}
		histogram[idx]++
	}

	peaks := 0
	for i := 0; i < len(histogram); i++ {
		left := -1
		right := -1
		if i > 0 {
			left = histogram[i-1]
		}
		if i < len(histogram)-1 {
			right = histogram[i+1]
		}
		if histogram[i] == 0 {
			continue
		}
		if (left <= histogram[i] || left == -1) && (right <= histogram[i] || right == -1) && (histogram[i] > left || histogram[i] > right) {
			peaks++
		}
	}
	if peaks < 1 {
		return 1
	}
	return peaks
}

func estimateNormalityScore(stats deferenceapp.NumericDistributionStats) float64 {
	if stats.Count == 0 {
		return 0
	}
	score := 100.0
	score -= math.Abs(stats.Skewness) * 18
	score -= math.Abs(stats.ExcessKurtosis) * 10
	if stats.PeakCount > 1 {
		score -= float64(stats.PeakCount-1) * 15
	}
	if stats.UniqueValueCount <= 5 && stats.Count >= 10 {
		score -= 18
	}
	missingRatio := float64(stats.MissingCount) / float64(stats.Count+stats.MissingCount)
	score -= missingRatio * 20
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func buildDistributionBins(values []float64, stats deferenceapp.NumericDistributionStats) []deferenceapp.DistributionBin {
	if len(values) == 0 {
		return []deferenceapp.DistributionBin{}
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	binCount := int(math.Round(math.Sqrt(float64(len(sorted)))))
	if binCount < 6 {
		binCount = 6
	}
	if binCount > 16 {
		binCount = 16
	}

	minValue := sorted[0]
	maxValue := sorted[len(sorted)-1]
	if minValue == maxValue {
		return []deferenceapp.DistributionBin{{
			Start:         minValue - 0.5,
			End:           maxValue + 0.5,
			Center:        minValue,
			ObservedCount: len(sorted),
			ExpectedCount: float64(len(sorted)),
			GapCount:      0,
		}}
	}

	width := (maxValue - minValue) / float64(binCount)
	if width <= 0 {
		width = 1
	}

	bins := make([]deferenceapp.DistributionBin, binCount)
	for i := 0; i < binCount; i++ {
		start := minValue + float64(i)*width
		end := start + width
		bins[i] = deferenceapp.DistributionBin{
			Start:  start,
			End:    end,
			Center: start + width/2,
		}
	}

	for _, value := range sorted {
		index := int(math.Floor((value - minValue) / width))
		if index < 0 {
			index = 0
		}
		if index >= binCount {
			index = binCount - 1
		}
		bins[index].ObservedCount++
	}

	for i := range bins {
		expected := 0.0
		if stats.StdDev <= 0 {
			expected = 1 / float64(binCount)
		} else {
			expected = normalCDF(bins[i].End, stats.Mean, stats.StdDev) - normalCDF(bins[i].Start, stats.Mean, stats.StdDev)
		}
		bins[i].ExpectedCount = expected * float64(len(sorted))
		bins[i].GapCount = bins[i].ExpectedCount - float64(bins[i].ObservedCount)
	}

	return bins
}

func normalCDF(x, mean, stdDev float64) float64 {
	if stdDev <= 0 {
		return 0
	}
	z := (x - mean) / (stdDev * math.Sqrt2)
	return 0.5 * (1 + math.Erf(z))
}

func buildDistributionComparison(current, baseline deferenceapp.NumericDistributionStats) *deferenceapp.DistributionComparison {
	return &deferenceapp.DistributionComparison{
		MeanDiff:           current.Mean - baseline.Mean,
		VarianceDiff:       current.Variance - baseline.Variance,
		StdDevDiff:         current.StdDev - baseline.StdDev,
		SkewnessDiff:       current.Skewness - baseline.Skewness,
		NormalityScoreDiff: current.NormalityScore - baseline.NormalityScore,
	}
}

func inferDistributionIssues(current deferenceapp.NumericDistributionStats, baseline *deferenceapp.NumericDistributionStats, axis string) []deferenceapp.DistributionIssue {
	issues := make([]deferenceapp.DistributionIssue, 0)

	totalObserved := current.Count + current.MissingCount
	missingRatio := 0.0
	if totalObserved > 0 {
		missingRatio = float64(current.MissingCount) / float64(totalObserved)
	}
	if missingRatio >= 0.30 {
		issues = append(issues, deferenceapp.DistributionIssue{
			Code:       "axis_missing_values",
			Category:   "format_gap",
			Severity:   "alert",
			Message:    axis + " に欠損が多く、分布の解釈が不安定です。",
			Suggestion: "この軸の値を埋める入力導線を追加してください。",
		})
	} else if missingRatio >= 0.15 {
		issues = append(issues, deferenceapp.DistributionIssue{
			Code:       "axis_missing_values",
			Category:   "format_gap",
			Severity:   "notice",
			Message:    axis + " に一定量の欠損があります。",
			Suggestion: "欠損が起きる条件を切り分けて補完方針を決めてください。",
		})
	}

	if current.UniqueValueCount <= 5 && current.Count >= 10 {
		issues = append(issues, deferenceapp.DistributionIssue{
			Code:       "axis_too_discrete",
			Category:   "format_gap",
			Severity:   "notice",
			Message:    axis + " は値の種類が少なく、粗い軸になっています。",
			Suggestion: "より細かい属性や補助タグを追加してください。",
		})
	}

	if math.Abs(current.Skewness) >= 1.5 {
		issues = append(issues, deferenceapp.DistributionIssue{
			Code:       "high_skewness",
			Category:   "hidden_factor",
			Severity:   "alert",
			Message:    axis + " の分布に強い歪みがあります。",
			Suggestion: "外れ値群や未分類の条件が混ざっていないか確認してください。",
		})
	} else if math.Abs(current.Skewness) >= 0.8 {
		issues = append(issues, deferenceapp.DistributionIssue{
			Code:       "high_skewness",
			Category:   "hidden_factor",
			Severity:   "notice",
			Message:    axis + " の分布に偏りがあります。",
			Suggestion: "分布の左右で条件差がないか調べてください。",
		})
	}

	if math.Abs(current.ExcessKurtosis) >= 2.0 {
		issues = append(issues, deferenceapp.DistributionIssue{
			Code:       "heavy_tails",
			Category:   "hidden_factor",
			Severity:   "alert",
			Message:    axis + " に裾の重い分布が見られます。",
			Suggestion: "極端な値を生む状況をタグや別軸で切り分けてください。",
		})
	}

	if current.PeakCount >= 2 {
		issues = append(issues, deferenceapp.DistributionIssue{
			Code:       "multi_modal",
			Category:   "hidden_factor",
			Severity:   "alert",
			Message:    axis + " の分布に複数の山があります。",
			Suggestion: "異なる集団が混ざっていないか検索条件を分割してください。",
		})
	}

	if baseline != nil && baseline.Count > 0 {
		if current.NormalityScore+20 < baseline.NormalityScore {
			issues = append(issues, deferenceapp.DistributionIssue{
				Code:       "less_stable_than_baseline",
				Category:   "hidden_factor",
				Severity:   "alert",
				Message:    "基準期間より分布が不安定です。",
				Suggestion: "最近の条件変化や分類漏れを見直してください。",
			})
		}
		if math.Abs(current.Mean-baseline.Mean) > baseline.StdDev && baseline.StdDev > 0 {
			issues = append(issues, deferenceapp.DistributionIssue{
				Code:       "mean_shift",
				Category:   "hidden_factor",
				Severity:   "notice",
				Message:    axis + " の平均値が基準期間から大きくずれています。",
				Suggestion: "期間差を生むイベントや条件タグを確認してください。",
			})
		}
	}

	if len(issues) == 0 {
		issues = append(issues, deferenceapp.DistributionIssue{
			Code:       "stable_enough",
			Category:   "resolved",
			Severity:   "ok",
			Message:    "大きな分布上の問題は見当たりません。",
			Suggestion: "別軸や別期間との比較で次の分析を進めてください。",
		})
	}

	return issues
}

func summarizeIssueTexts(issues []deferenceapp.DistributionIssue, category string) []string {
	seen := make(map[string]struct{}, len(issues))
	items := make([]string, 0, len(issues))
	for _, issue := range issues {
		if issue.Category != category || issue.Message == "" {
			continue
		}
		if _, ok := seen[issue.Message]; ok {
			continue
		}
		seen[issue.Message] = struct{}{}
		items = append(items, issue.Message)
	}
	return items
}

func summarizeSuggestedActions(issues []deferenceapp.DistributionIssue) []string {
	seen := make(map[string]struct{}, len(issues))
	actions := make([]string, 0, len(issues))
	for _, issue := range issues {
		if issue.Suggestion == "" {
			continue
		}
		if _, ok := seen[issue.Suggestion]; ok {
			continue
		}
		seen[issue.Suggestion] = struct{}{}
		actions = append(actions, issue.Suggestion)
	}
	return actions
}

func suggestTagsForDistributionIssues(dataset deferenceapp.DistributionDataset, axis string, issues []deferenceapp.DistributionIssue) []string {
	seen := make(map[string]struct{})
	tags := make([]string, 0, 8)

	appendTag := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		tags = append(tags, value)
	}

	for _, issue := range issues {
		switch issue.Code {
		case "axis_missing_values":
			appendTag(axis + "_missing")
		case "axis_too_discrete":
			appendTag(axis + "_granularity")
		case "high_skewness":
			appendTag("outlier_candidate")
		case "heavy_tails":
			appendTag("extreme_case")
		case "multi_modal":
			appendTag("mixed_population")
		case "less_stable_than_baseline":
			appendTag("baseline_shift")
		case "mean_shift":
			appendTag("mean_shift")
		}
	}

	if dataset == deferenceapp.DistributionDatasetWorldSignal {
		appendTag("external_condition")
	}

	return tags
}

func suggestAxesForDistributionIssues(dataset deferenceapp.DistributionDataset, axis string, issues []deferenceapp.DistributionIssue) []string {
	seen := make(map[string]struct{})
	axes := make([]string, 0, 6)

	appendAxis := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || value == axis {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		axes = append(axes, value)
	}

	switch dataset {
	case deferenceapp.DistributionDatasetActionLogs:
		for _, issue := range issues {
			switch issue.Code {
			case "high_skewness", "less_stable_than_baseline", "mean_shift":
				appendAxis("occurred_at")
			case "multi_modal":
				appendAxis("prototype_id")
				appendAxis("tag_count")
			case "axis_missing_values", "axis_too_discrete":
				appendAxis("tag_count")
			case "heavy_tails":
				appendAxis("created_at")
			}
		}
	case deferenceapp.DistributionDatasetWorldSignal:
		for _, candidate := range []string{"temperature_c", "precipitation_mm", "wind_speed_ms", "weather_code", "signal_value"} {
			appendAxis(candidate)
		}
	}

	return axes
}

func minAnalysisInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
