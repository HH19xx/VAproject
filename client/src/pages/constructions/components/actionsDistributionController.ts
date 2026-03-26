import type {
  CreateAnalysisSnapshotInput,
  AnalysisSnapshot,
} from "../../../hooks/useAnalysisSnapshots";
import type {
  AnalyzeDistributionInput,
  DistributionAnalysisResult,
  DistributionDataset,
} from "../../../hooks/useDistributionAnalysis";
import type { ExternalDataSource } from "../../../hooks/useWorldSignals";
import { toLegacyDistributionStats, toSnapshotScore, toSnapshotSeverity } from "./actionsDistribution/actionsDistributionScore";
import type { SelectedDistributionBin } from "./actionsDistribution/actionsDistributionTypes";
import type { ParsedQuery } from "./actionsSearch/actionsSearchQuery";

export const resolveBaselineWindow = (from: string, to: string): { from?: string; to?: string } => {
  if (!from || !to) return {};
  const fromDate = new Date(from);
  const toDate = new Date(to);
  const duration = toDate.getTime() - fromDate.getTime();
  const baselineTo = fromDate;
  const baselineFrom = new Date(fromDate.getTime() - duration);
  return {
    from: baselineFrom.toISOString(),
    to: baselineTo.toISOString(),
  };
};

export const saveDistributionSnapshot = async (
  createSnapshot: (input: CreateAnalysisSnapshotInput) => Promise<AnalysisSnapshot | null>,
  danbooruQuery: string,
  result: DistributionAnalysisResult
): Promise<AnalysisSnapshot | null> =>
  createSnapshot({
    query_text: `${danbooruQuery.trim()} axis:${result.axis} dataset:${result.dataset}`.trim(),
    severity: toSnapshotSeverity(result),
    score: Math.round(toSnapshotScore(result)),
    delta_avg_tag: result.comparison?.mean_diff ?? 0,
    delta_var_tag: result.comparison?.variance_diff ?? 0,
    delta_prototype_rate: 0,
    current: toLegacyDistributionStats(result.current),
    baseline: result.baseline ? toLegacyDistributionStats(result.baseline) : undefined,
  });

export const executeDistributionAnalysis = async (
  analyzeDistribution: (input: AnalyzeDistributionInput) => Promise<DistributionAnalysisResult | null>,
  input: AnalyzeDistributionInput
): Promise<DistributionAnalysisResult | null> => analyzeDistribution(input);

export const buildDistributionExecutionState = (
  danbooruQuery: string,
  tagNameToID: Map<string, number>,
  filterTagIDs: number[],
  from: string,
  to: string,
  parseQuery: (query: string, tagNameToID: Map<string, number>) => ParsedQuery
): {
  parsed: ParsedQuery | null;
  mergedAnd: number[];
  range: { from?: string; to?: string };
  error: string | null;
} => {
  const parsed = parseQuery(danbooruQuery, tagNameToID);
  if (parsed.unknownTokens.length > 0) {
    return {
      parsed: null,
      mergedAnd: [],
      range: { from: from ? new Date(from).toISOString() : undefined, to: to ? new Date(to).toISOString() : undefined },
      error: `未知のタグがあります: ${parsed.unknownTokens.join(", ")}`,
    };
  }

  return {
    parsed,
    mergedAnd: Array.from(new Set([...filterTagIDs, ...parsed.andTagIDs])),
    range: { from: from ? new Date(from).toISOString() : undefined, to: to ? new Date(to).toISOString() : undefined },
    error: null,
  };
};

export const getDistributionDatasetLabel = (dataset: DistributionDataset): string =>
  dataset === "action_logs" ? "行動記録" : "外部ビッグデータ";

export const getSelectedDatasetCount = (
  dataset: DistributionDataset,
  contextResult: { summary: { action_count: number; signal_count: number } } | null,
  actionsLength: number
): number =>
  dataset === "action_logs" ? contextResult?.summary.action_count ?? actionsLength : contextResult?.summary.signal_count ?? 0;

export const getDistributionEmptyHint = (
  dataset: DistributionDataset,
  selectedDatasetCount: number
): string | null => {
  if (dataset === "action_logs" && selectedDatasetCount === 0) {
    return "行動記録が0件です。検索条件を見直すか、先にデータを追加してください。";
  }
  if (dataset === "world_signals" && selectedDatasetCount === 0) {
    return "外部ビッグデータが0件です。Open-Meteo取得か分析期間を見直してください。";
  }
  return null;
};

export const buildSelectedBinStatus = (selectedDistributionBin: SelectedDistributionBin | null) => {
  if (!selectedDistributionBin) return null;

  if (selectedDistributionBin.bin.gap_count > 0) {
    return {
      kind: "shortage" as const,
      label: "不足帯域",
      message: "この帯域には、期待より少ない値しか入っていません。条件追加や群分割の候補を試してください。",
    };
  }

  if (selectedDistributionBin.bin.gap_count < 0) {
    return {
      kind: "excess" as const,
      label: "過剰帯域",
      message: "この帯域には、期待より値が集まりすぎています。偏りを生む条件が残っている可能性があります。",
    };
  }

  return {
    kind: "balanced" as const,
    label: "均衡帯域",
    message: "この帯域は期待分布に近い状態です。ほかの帯域との差分確認に使ってください。",
  };
};

export const selectDistributionBin = (
  distributionResult: DistributionAnalysisResult | null,
  index: number
): SelectedDistributionBin | null => {
  if (!distributionResult?.residual_bins[index]) return null;
  return { index, bin: distributionResult.residual_bins[index] };
};

export const resolveNextDistributionAxis = (
  dataset: DistributionDataset,
  externalDataSource: ExternalDataSource,
  actionAxisCandidates: string[],
  resolveDefaultAxis: (dataset: DistributionDataset, actionAxisCandidates: string[]) => string
): string => {
  if (dataset === "world_signals") {
    return externalDataSource === "e_stat_dashboard" ? "signal_value" : "temperature_c";
  }
  return resolveDefaultAxis(dataset, actionAxisCandidates);
};

export const buildResetDateRangeState = (
  syncDateRange: (pastDays: string, forecastDays: string) => { from: string; to: string } | null,
  pastDays: string,
  forecastDays: string
) => syncDateRange(pastDays, forecastDays);
