import type { ActionAttribute, ActionLog } from "../../../hooks/useActions";
import type { WorldSignal } from "../../../hooks/useWorldSignals";
import type {
  ActionAxisKey,
  CorrelationAnalysisResult,
  RepeatBehaviorResult,
  RepeatBucketUnit,
  SummaryStats,
  TwoGroupAnalysisResult,
  WorldAxisKey,
} from "./analysisTypes";

export const DAY_MS = 24 * 60 * 60 * 1000;
export const RECENT_ACTION_LOOKBACK_DAYS = 90;
export const ACTION_COMPARE_WINDOW_DAYS = 14;
export const ACTION_CORRELATION_WINDOW_DAYS = 28;
export const ESTAT_COMPARE_YEARS = 7;
export const ESTAT_CORRELATION_YEARS = 14;

export const toDateTimeLocal = (date: Date) =>
  new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);

export const createRange = (startDaysAgo: number, endDaysAgo: number) => ({
  from: toDateTimeLocal(new Date(Date.now() - startDaysAgo * DAY_MS)),
  to: toDateTimeLocal(new Date(Date.now() - endDaysAgo * DAY_MS)),
});

export const createYearRange = (startYearsAgo: number, endYearsAgo: number) => ({
  from: toDateTimeLocal(new Date(Date.UTC(new Date().getUTCFullYear() - startYearsAgo, 0, 1))),
  to: toDateTimeLocal(new Date(Date.UTC(new Date().getUTCFullYear() - endYearsAgo, 0, 1))),
});

const average = (values: number[]) => values.reduce((sum, value) => sum + value, 0) / values.length;

const sampleVariance = (values: number[], mean: number) =>
  values.length < 2 ? 0 : values.reduce((sum, value) => sum + (value - mean) ** 2, 0) / (values.length - 1);

const summarize = (values: number[]): SummaryStats => {
  if (values.length === 0) {
    return { count: 0, mean: 0, variance: 0, stdDev: 0 };
  }
  const mean = average(values);
  const variance = sampleVariance(values, mean);
  return { count: values.length, mean, variance, stdDev: Math.sqrt(variance) };
};

const erf = (x: number) => {
  const sign = x >= 0 ? 1 : -1;
  const absX = Math.abs(x);
  const t = 1 / (1 + 0.3275911 * absX);
  const y =
    1 -
    (((((1.061405429 * t - 1.453152027) * t + 1.421413741) * t - 0.284496736) * t +
      0.254829592) *
      t *
      Math.exp(-absX * absX));
  return sign * y;
};

const normalCdf = (x: number) => 0.5 * (1 + erf(x / Math.SQRT2));

export const formatNumber = (value: number, digits = 3) => value.toFixed(digits);

export const repeatBucketLabel = (unit: RepeatBucketUnit) =>
  unit === "week" ? "週単位" : "日単位";

const findAttributeValue = (attributes: ActionAttribute[] | undefined, key: string) =>
  attributes?.find((attribute) => attribute.key === key)?.value_number;

export const actionValue = (log: ActionLog, axis: ActionAxisKey) => {
  if (axis === "tag_count") return log.tag_ids.length;
  if (axis === "prototype_id") return log.prototype_id ?? null;
  return findAttributeValue(log.attributes, axis) ?? null;
};

export const worldValue = (signal: WorldSignal, axis: WorldAxisKey) => {
  if (axis === "observed_year") return new Date(signal.observed_at).getUTCFullYear();
  if (axis === "signal_value") return signal.signal_value ?? null;
  return signal[axis] ?? null;
};

const filterActionLogsByKeyword = (logs: ActionLog[], keyword: string) => {
  const normalized = keyword.trim().toLowerCase();
  if (!normalized) return logs;
  return logs.filter((log) => {
    const title = log.title?.toLowerCase() ?? "";
    const notes = log.notes?.toLowerCase() ?? "";
    return title.includes(normalized) || notes.includes(normalized);
  });
};

export const collectActionAxisOptions = (logs: ActionLog[]) => {
  const keys = new Set<string>(["tag_count", "prototype_id"]);
  logs.forEach((log) => {
    log.attributes?.forEach((attribute) => {
      if (attribute.key) keys.add(attribute.key);
    });
  });
  return Array.from(keys);
};

export const actionAxisLabel = (axis: ActionAxisKey) => {
  if (axis === "tag_count") return "タグ数";
  if (axis === "prototype_id") return "プロトタイプID";
  return axis;
};

export const computeWelchTest = (
  groupAValues: number[],
  groupBValues: number[]
): TwoGroupAnalysisResult | null => {
  if (groupAValues.length < 2 || groupBValues.length < 2) return null;

  const groupA = summarize(groupAValues);
  const groupB = summarize(groupBValues);
  const varA = groupA.variance / groupA.count;
  const varB = groupB.variance / groupB.count;
  const denominator = Math.sqrt(varA + varB);
  if (denominator === 0) return null;

  const meanDiff = groupA.mean - groupB.mean;
  const tStatistic = meanDiff / denominator;
  const numerator = (varA + varB) ** 2;
  const denominatorDf = varA ** 2 / (groupA.count - 1) + varB ** 2 / (groupB.count - 1);
  const degreesOfFreedom =
    denominatorDf === 0 ? groupA.count + groupB.count - 2 : numerator / denominatorDf;

  return {
    groupA,
    groupB,
    meanDiff,
    tStatistic,
    degreesOfFreedom,
    pValueApprox: 2 * (1 - normalCdf(Math.abs(tStatistic))),
  };
};

export const computeCorrelation = (
  pairs: Array<{ x: number; y: number }>
): CorrelationAnalysisResult | null => {
  if (pairs.length < 3) return null;

  const xValues = pairs.map((pair) => pair.x);
  const yValues = pairs.map((pair) => pair.y);
  const meanX = average(xValues);
  const meanY = average(yValues);

  let covarianceNumerator = 0;
  let varianceX = 0;
  let varianceY = 0;

  pairs.forEach(({ x, y }) => {
    const dx = x - meanX;
    const dy = y - meanY;
    covarianceNumerator += dx * dy;
    varianceX += dx * dx;
    varianceY += dy * dy;
  });

  if (varianceX === 0 || varianceY === 0) return null;

  const covariance = covarianceNumerator / (pairs.length - 1);
  const correlation = covarianceNumerator / Math.sqrt(varianceX * varianceY);
  const denominator = 1 - correlation ** 2;
  const tStatistic =
    denominator <= 0 ? 0 : correlation * Math.sqrt((pairs.length - 2) / denominator);

  return {
    count: pairs.length,
    correlation,
    covariance,
    pValueApprox: denominator <= 0 ? 0 : 2 * (1 - normalCdf(Math.abs(tStatistic))),
  };
};

const startOfBucket = (timestamp: number, unit: RepeatBucketUnit) => {
  const date = new Date(timestamp);
  if (unit === "week") {
    const copy = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate()));
    const weekday = copy.getUTCDay();
    const diff = weekday === 0 ? -6 : 1 - weekday;
    copy.setUTCDate(copy.getUTCDate() + diff);
    return copy.getTime();
  }
  return Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate());
};

export const computeRepeatBehavior = (
  logs: ActionLog[],
  keyword: string,
  bucketUnit: RepeatBucketUnit,
  expectedMean: number
): RepeatBehaviorResult | null => {
  const filtered = filterActionLogsByKeyword(logs, keyword);
  if (filtered.length === 0) return null;

  const bucketMap = new Map<number, number>();
  filtered.forEach((log) => {
    const bucket = startOfBucket(new Date(log.occurred_at).getTime(), bucketUnit);
    bucketMap.set(bucket, (bucketMap.get(bucket) ?? 0) + 1);
  });

  const counts = Array.from(bucketMap.values());
  const summary = summarize(counts);
  const margin = counts.length >= 2 ? 1.96 * (summary.stdDev / Math.sqrt(counts.length)) : 0;

  const tStatistic =
    counts.length >= 2 && summary.stdDev > 0
      ? (summary.mean - expectedMean) / (summary.stdDev / Math.sqrt(counts.length))
      : null;

  return {
    bucketUnit,
    bucketCount: counts.length,
    matchedActionCount: filtered.length,
    observedMean: summary.mean,
    expectedMean,
    stdDev: summary.stdDev,
    confidenceLow: summary.mean - margin,
    confidenceHigh: summary.mean + margin,
    tStatistic,
    pValueApprox: tStatistic === null ? null : 2 * (1 - normalCdf(Math.abs(tStatistic))),
  };
};
