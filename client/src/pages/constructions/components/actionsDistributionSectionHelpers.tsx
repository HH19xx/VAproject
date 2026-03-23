import type { ReactElement } from "react";
import type { ActionLog } from "../../../hooks/useActions";
import type { AnalysisSeverity, DistributionStats } from "../../../hooks/useAnalysisSnapshots";
import type {
  DistributionAnalysisResult,
  DistributionBin,
  DistributionDataset,
  NumericDistributionStats,
} from "../../../hooks/useDistributionAnalysis";

export interface SelectedDistributionBin {
  index: number;
  bin: DistributionBin;
}

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

export const ACTION_AXIS_DEFAULTS = [
  "occurred_at",
  "created_at",
  "updated_at",
  "tag_count",
  "prototype_id",
] as const;

export const actionAxisValue = (action: ActionLog, axis: string): number | null => {
  if (axis === "occurred_at") return new Date(action.occurred_at).getTime();
  if (axis === "created_at") return new Date(action.created_at).getTime();
  if (axis === "updated_at") return new Date(action.updated_at).getTime();
  if (axis === "tag_count") return action.tag_ids.length;
  if (axis === "prototype_id") return action.prototype_id ?? null;

  const attr = action.attributes?.find((item) => item.key === axis);
  return typeof attr?.value_number === "number" ? attr.value_number : null;
};

const SCORE_WEIGHT_AVG = 0.5;
const SCORE_WEIGHT_VAR = 0.3;
const SCORE_WEIGHT_PROTOTYPE = 0.2;
const SVG_VIEWBOX_WIDTH = 900;
const SVG_VIEWBOX_HEIGHT = 260;
const SVG_MARGIN_LEFT = 40;
const SVG_MARGIN_TOP = 30;
const SVG_DRAW_WIDTH = 820;
const SVG_DRAW_HEIGHT = 180;
const GAP_ZERO_LINE_Y = 130;
const GAP_BAR_HALF_WIDTH = 12;

export const formatNum = (value: number, digits = 3) => value.toFixed(digits);

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
  const se = Math.sqrt((stdCurrent ** 2) / current.count + (stdBaseline ** 2) / baseline.count);
  if (se <= 0) return undefined;

  const z = Math.abs(current.avg_tag_count - baseline.avg_tag_count) / se;
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

export const axisLabelMap = {
  temperature_c: "気温",
  precipitation_mm: "降水量",
  wind_speed_ms: "風速",
  weather_code: "天気コード",
  signal_value: "指標値",
  observed_year: "観測年",
} as const;

export const toActionAxisLabel = (axis: string): string => {
  if (axis === "occurred_at") return "発生日時";
  if (axis === "created_at") return "作成日時";
  if (axis === "updated_at") return "更新日時";
  if (axis === "tag_count") return "タグ数";
  if (axis === "prototype_id") return "プロトタイプ";
  return axis;
};

export const toResidualStrategyLabel = (value: unknown): string => {
  if (value === "rolling_mean_time") return "時系列移動平均差";
  if (value === "standardized_discrete_residuals") return "離散値の標準化残差";
  if (value === "local_expected_continuous") return "局所期待値との差";
  return String(value || "-");
};

export const toResidualAxisTypeLabel = (value: unknown): string => {
  if (value === "time") return "時間軸";
  if (value === "discrete") return "離散軸";
  if (value === "continuous") return "連続軸";
  return String(value || "-");
};

export const toResidualAxisReasonLabel = (value: unknown): string => {
  if (value === "built_in_axis") return "組み込み軸";
  if (value === "unique_value_count<=6") return "ユニーク値数が少ない";
  if (value === "unique_ratio<=0.35") return "ユニーク比率が低い";
  if (value === "unique_ratio<=0.50_and_unique<=12") return "半離散に近い";
  if (value === "default_continuous") return "既定で連続軸扱い";
  return String(value || "-");
};

export const toDistributionSeverity = (result: DistributionAnalysisResult): AnalysisSeverity => toSnapshotSeverity(result);

export const toBadgeClass = (severity: AnalysisSeverity): string => {
  if (severity === "ALERT") return "alertBadge";
  if (severity === "NOTICE") return "noticeBadge";
  return "okBadge";
};

export const toIssueClass = (category: string): string => {
  if (category === "format_gap") return "formatGap";
  if (category === "hidden_factor") return "hiddenFactor";
  return "resolvedIssue";
};

export const defaultAxisByDataset = (dataset: DistributionDataset, actionAxisCandidates: string[]): string => {
  if (dataset === "world_signals") return "temperature_c";
  return actionAxisCandidates.includes("tag_count") ? "tag_count" : actionAxisCandidates[0] || "tag_count";
};

export const renderDistributionOverviewSvg = (
  bins: DistributionBin[],
  selectedIndex: number | null,
  onSelect: (index: number) => void
): ReactElement => {
  const maxObserved = Math.max(1, ...bins.map((bin) => bin.observed_count), ...bins.map((bin) => bin.expected_count));
  const binWidth = SVG_DRAW_WIDTH / Math.max(1, bins.length);

  return (
    <svg className="scatterSvg" viewBox={`0 0 ${SVG_VIEWBOX_WIDTH} ${SVG_VIEWBOX_HEIGHT}`} preserveAspectRatio="none">
      {bins.map((bin, index) => {
        const barHeight = (bin.observed_count / maxObserved) * SVG_DRAW_HEIGHT;
        const x = SVG_MARGIN_LEFT + index * binWidth;
        const y = SVG_MARGIN_TOP + (SVG_DRAW_HEIGHT - barHeight);
        const isActive = selectedIndex === index;
        return (
          <rect
            key={`dist-bin-${bin.start}-${bin.end}`}
            x={x + 2}
            y={y}
            width={Math.max(2, binWidth - 4)}
            height={barHeight}
            fill={isActive ? "#0056b3" : "#66b2ff"}
            opacity="0.85"
            role="button"
            tabIndex={0}
            onClick={() => onSelect(index)}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                onSelect(index);
              }
            }}
          >
            <title>{`${formatNum(bin.start, 2)} - ${formatNum(bin.end, 2)} / 観測=${formatNum(bin.observed_count, 2)} / 期待=${formatNum(bin.expected_count, 2)}`}</title>
          </rect>
        );
      })}
      <polyline
        fill="none"
        stroke="#ff7f0e"
        strokeWidth="3"
        points={bins
          .map((bin, index) => {
            const x = SVG_MARGIN_LEFT + index * binWidth + binWidth / 2;
            const y = SVG_MARGIN_TOP + (SVG_DRAW_HEIGHT - (bin.expected_count / maxObserved) * SVG_DRAW_HEIGHT);
            return `${x},${y}`;
          })
          .join(" ")}
      />
    </svg>
  );
};

export const renderGapSvg = (
  bins: DistributionBin[],
  selectedIndex: number | null,
  onSelect: (index: number) => void
): ReactElement => {
  const maxGap = Math.max(1, ...bins.map((bin) => Math.abs(bin.gap_count)));
  const binWidth = SVG_DRAW_WIDTH / Math.max(1, bins.length);

  return (
    <svg className="scatterSvg" viewBox={`0 0 ${SVG_VIEWBOX_WIDTH} ${SVG_VIEWBOX_HEIGHT}`} preserveAspectRatio="none">
      <line
        x1={SVG_MARGIN_LEFT}
        y1={GAP_ZERO_LINE_Y}
        x2={SVG_MARGIN_LEFT + SVG_DRAW_WIDTH}
        y2={GAP_ZERO_LINE_Y}
        stroke="#666"
        strokeDasharray="4 3"
      />
      {bins.map((bin, index) => {
        const x = SVG_MARGIN_LEFT + index * binWidth + binWidth / 2;
        const height = (Math.abs(bin.gap_count) / maxGap) * 90;
        const shortage = bin.gap_count > 0;
        const y = shortage ? GAP_ZERO_LINE_Y - height : GAP_ZERO_LINE_Y;
        const isActive = selectedIndex === index;
        return (
          <rect
            key={`gap-bin-${bin.start}-${bin.end}`}
            x={x - GAP_BAR_HALF_WIDTH}
            y={y}
            width={GAP_BAR_HALF_WIDTH * 2}
            height={height}
            fill={shortage ? (isActive ? "#198754" : "#7bc67b") : isActive ? "#b02a37" : "#ff8a80"}
            role="button"
            tabIndex={0}
            onClick={() => onSelect(index)}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                onSelect(index);
              }
            }}
          >
            <title>{`${formatNum(bin.start, 2)} - ${formatNum(bin.end, 2)} / 差分=${formatNum(bin.gap_count, 2)}`}</title>
          </rect>
        );
      })}
    </svg>
  );
};
