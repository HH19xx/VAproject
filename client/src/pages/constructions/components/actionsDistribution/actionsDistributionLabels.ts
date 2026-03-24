import type { AnalysisSeverity } from "../../../../hooks/useAnalysisSnapshots";

export const formatNum = (value: number, digits = 3) => value.toFixed(digits);

export const toActionAxisLabel = (axis: string): string => {
  if (axis === "occurred_at") return "発生日時";
  if (axis === "created_at") return "作成日時";
  if (axis === "updated_at") return "更新日時";
  if (axis === "tag_count") return "タグ数";
  if (axis === "prototype_id") return "プロトタイプID";
  return axis;
};

export const toResidualStrategyLabel = (value: unknown): string => {
  if (value === "rolling_mean_time") return "時間窓の移動平均差";
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
  if (value === "built_in_axis") return "既定の軸判定";
  if (value === "unique_value_count<=6") return "ユニーク値数が少ないため離散扱い";
  if (value === "unique_ratio<=0.35") return "ユニーク比率が低いため離散扱い";
  if (value === "unique_ratio<=0.50_and_unique<=12") return "値の種類が限定的なため離散扱い";
  if (value === "default_continuous") return "既定で連続軸として扱う";
  return String(value || "-");
};

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
