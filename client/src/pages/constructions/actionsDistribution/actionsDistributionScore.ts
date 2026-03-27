import type { ActionLog } from "../../../hooks/useActions";
import type { AnalysisSeverity, DistributionStats } from "../../../hooks/useAnalysisSnapshots";
import type {
  DistributionAnalysisResult,
  NumericDistributionStats,
} from "../../../hooks/useDistributionAnalysis";

export interface AnalysisViewState {
  severity: AnalysisSeverity;
  score: number;
  pValue?: number;
  significant?: boolean;
  deltaAvgTag: number;
  deltaVarTag: number;
  deltaPrototypeRate: number;
  current: DistributionStats;
  baseline?: DistributionStats;
}

const SCORE_WEIGHT_AVG = 0.5;
const SCORE_WEIGHT_VAR = 0.3;
const SCORE_WEIGHT_PROTOTYPE = 0.2;

export const calculateDistributionStats = (logs: ActionLog[]): DistributionStats => {
  if (logs.length === 0) {
    return {
      count: 0,
      avg_tag_count: 0,
      var_tag_count: 0,
      prototype_rate: 0,
    };
  }

  const tagCounts = logs.map((log) => log.tag_ids.length);
  const mean = tagCounts.reduce((sum, value) => sum + value, 0) / tagCounts.length;
  const variance = tagCounts.reduce((sum, value) => sum + (value - mean) ** 2, 0) / tagCounts.length;
  const prototypeCount = logs.filter((log) => Boolean(log.prototype_id)).length;

  return {
    count: logs.length,
    avg_tag_count: mean,
    var_tag_count: variance,
    prototype_rate: prototypeCount / logs.length,
  };
};

const computeStdDev = (variance: number): number => Math.sqrt(Math.max(0, variance));

const approxNormalCDF = (x: number): number => {
  const absX = Math.abs(x);
  const t = 1 / (1 + 0.2316419 * absX);
  const d = 0.3989423 * Math.exp((-absX * absX) / 2);
  const prob =
    1 -
    d *
      t *
      (0.3193815 + t * (-0.3565638 + t * (1.781478 + t * (-1.821256 + t * 1.330274))));
  return x >= 0 ? prob : 1 - prob;
};

const estimatePValueByMeanDiff = (current: DistributionStats, baseline?: DistributionStats): number | undefined => {
  if (!baseline) return undefined;
  if (current.count < 2 || baseline.count < 2) return undefined;

  const stdCurrent = computeStdDev(current.var_tag_count);
  const stdBaseline = computeStdDev(baseline.var_tag_count);
  const standardError = Math.sqrt((stdCurrent ** 2) / current.count + (stdBaseline ** 2) / baseline.count);
  if (standardError <= 0) return undefined;

  const z = Math.abs(current.avg_tag_count - baseline.avg_tag_count) / standardError;
  return Math.max(0, Math.min(1, 2 * (1 - approxNormalCDF(z))));
};

export const deriveSeverity = (
  current: DistributionStats,
  baseline?: DistributionStats
): Omit<AnalysisViewState, "current" | "baseline"> => {
  const base = baseline ?? {
    count: 0,
    avg_tag_count: 0,
    var_tag_count: 0,
    prototype_rate: 0,
  };

  const deltaAvgTag = current.avg_tag_count - base.avg_tag_count;
  const deltaVarTag = current.var_tag_count - base.var_tag_count;
  const deltaPrototypeRate = current.prototype_rate - base.prototype_rate;
  const score =
    Math.abs(deltaAvgTag) * SCORE_WEIGHT_AVG +
    Math.abs(deltaVarTag) * SCORE_WEIGHT_VAR +
    Math.abs(deltaPrototypeRate) * 100 * SCORE_WEIGHT_PROTOTYPE;
  const pValue = estimatePValueByMeanDiff(current, baseline);
  const significant = pValue !== undefined ? pValue < 0.05 : undefined;

  let severity: AnalysisSeverity = "OK";
  if (score >= 12 || significant === true) {
    severity = "ALERT";
  } else if (score >= 5) {
    severity = "NOTICE";
  }

  return {
    severity,
    score,
    pValue,
    significant,
    deltaAvgTag,
    deltaVarTag,
    deltaPrototypeRate,
  };
};

export const toLegacyDistributionStats = (stats: NumericDistributionStats): DistributionStats => ({
  count: stats.count,
  avg_tag_count: stats.mean,
  var_tag_count: stats.variance,
  prototype_rate: 0,
});

export const toSnapshotSeverity = (result: DistributionAnalysisResult): AnalysisSeverity => {
  if (result.issues.some((issue) => issue.severity === "alert")) return "ALERT";
  if (result.issues.some((issue) => issue.severity === "notice")) return "NOTICE";
  return "OK";
};

export const toSnapshotScore = (result: DistributionAnalysisResult): number => {
  const meanScore = Math.abs(result.comparison?.mean_diff ?? 0);
  const varianceScore = Math.abs(result.comparison?.variance_diff ?? 0);
  const normalityScore = Math.abs(result.comparison?.normality_score_diff ?? 0);
  return meanScore * 0.5 + varianceScore * 0.2 + normalityScore * 0.3 + result.issues.length;
};

export const toDistributionSeverity = (result: DistributionAnalysisResult): AnalysisSeverity => toSnapshotSeverity(result);

