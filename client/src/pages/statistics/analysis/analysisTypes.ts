export type DatasetKind = "action_logs" | "world_signals";

export type ActionAxisKey = "tag_count" | "prototype_id" | string;

export type WorldAxisKey =
  | "temperature_c"
  | "precipitation_mm"
  | "wind_speed_ms"
  | "weather_code"
  | "signal_value"
  | "observed_year";

export type RepeatBucketUnit = "day" | "week";

export type SummaryStats = {
  count: number;
  mean: number;
  variance: number;
  stdDev: number;
};

export type TwoGroupAnalysisResult = {
  groupA: SummaryStats;
  groupB: SummaryStats;
  meanDiff: number;
  tStatistic: number;
  degreesOfFreedom: number;
  pValueApprox: number;
};

export type CorrelationAnalysisResult = {
  count: number;
  correlation: number;
  covariance: number;
  pValueApprox: number;
};

export type RepeatBehaviorResult = {
  bucketUnit: RepeatBucketUnit;
  bucketCount: number;
  matchedActionCount: number;
  observedMean: number;
  expectedMean: number;
  stdDev: number;
  confidenceLow: number;
  confidenceHigh: number;
  tStatistic: number | null;
  pValueApprox: number | null;
};

export type StatisticsNavigationState = {
  dataset?: DatasetKind;
  danbooruQuery?: string;
  from?: string;
  to?: string;
  filterTagIDs?: number[];
};
