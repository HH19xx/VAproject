import { useContext, useState } from "react";
import { AuthContext } from "../contexts/AuthContext";

export type DistributionDataset = "action_logs" | "world_signals";

export interface NumericDistributionStats {
  count: number;
  missing_count: number;
  mean: number;
  variance: number;
  std_dev: number;
  min: number;
  max: number;
  skewness: number;
  excess_kurtosis: number;
  unique_value_count: number;
  peak_count: number;
  normality_score: number;
  sample_values?: number[];
}

export interface DistributionIssue {
  code: string;
  category: string;
  severity: "ok" | "notice" | "alert";
  message: string;
  suggestion: string;
}

export interface DistributionComparison {
  mean_diff: number;
  variance_diff: number;
  std_dev_diff: number;
  skewness_diff: number;
  normality_score_diff: number;
}

export interface DistributionBin {
  start: number;
  end: number;
  center: number;
  observed_count: number;
  expected_count: number;
  gap_count: number;
}

export interface DistributionAnalysisResult {
  dataset: DistributionDataset;
  axis: string;
  current: NumericDistributionStats;
  residual_current: NumericDistributionStats;
  baseline?: NumericDistributionStats;
  residual_baseline?: NumericDistributionStats;
  comparison?: DistributionComparison;
  bins: DistributionBin[];
  raw_bins: DistributionBin[];
  residual_bins: DistributionBin[];
  issues: DistributionIssue[];
  hidden_factor_candidates: string[];
  format_suggestions: string[];
  suggested_actions: string[];
  suggested_tags: string[];
  suggested_axes: string[];
  meta?: Record<string, unknown>;
}

export interface AnalyzeDistributionInput {
  dataset: DistributionDataset;
  axis: string;
  source?: string;
  signalType?: string;
  tagIDs?: number[];
  anyTagIDs?: number[];
  anyTagGroups?: number[][];
  excludeTagIDs?: number[];
  from?: string;
  to?: string;
  locationKey?: string;
}

export const useDistributionAnalysis = () => {
  const authContext = useContext(AuthContext);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) {
    throw new Error("useDistributionAnalysis requires AuthProvider");
  }

  const { authFetch } = authContext;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";


  const analyzeDistribution = async (input: AnalyzeDistributionInput): Promise<DistributionAnalysisResult | null> => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams();
      params.set("dataset", input.dataset);
      params.set("axis", input.axis);
      if (input.source) {
        params.set("source", input.source);
      }
      if (input.signalType) {
        params.set("signal_type", input.signalType);
      }
      if (input.tagIDs && input.tagIDs.length > 0) {
        params.set("tag_ids", input.tagIDs.join(","));
      }
      if (input.anyTagIDs && input.anyTagIDs.length > 0) {
        params.set("any_tag_ids", input.anyTagIDs.join(","));
      }
      if (input.anyTagGroups && input.anyTagGroups.length > 0) {
        params.set(
          "any_tag_groups",
          input.anyTagGroups.map((group) => group.join("+")).join(",")
        );
      }
      if (input.excludeTagIDs && input.excludeTagIDs.length > 0) {
        params.set("exclude_tag_ids", input.excludeTagIDs.join(","));
      }
      if (input.from) {
        params.set("from", input.from);
      }
      if (input.to) {
        params.set("to", input.to);
      }
      if (input.locationKey) {
        params.set("location_key", input.locationKey);
      }

      const data = await authFetch(`${API_BASE_URL}/analysis/distribution?${params.toString()}`);
      if (!data.success) {
        throw new Error(data.message || "distribution analysis request failed");
      }
      const result = (data.data || {}) as Partial<DistributionAnalysisResult>;
      return {
        dataset: result.dataset || input.dataset,
        axis: result.axis || input.axis,
        current: result.current as NumericDistributionStats,
        residual_current: result.residual_current as NumericDistributionStats,
        baseline: result.baseline,
        residual_baseline: result.residual_baseline,
        comparison: result.comparison,
        bins: result.bins || [],
        raw_bins: result.raw_bins || result.bins || [],
        residual_bins: result.residual_bins || result.bins || [],
        issues: result.issues || [],
        hidden_factor_candidates: result.hidden_factor_candidates || [],
        format_suggestions: result.format_suggestions || [],
        suggested_actions: result.suggested_actions || [],
        suggested_tags: result.suggested_tags || [],
        suggested_axes: result.suggested_axes || [],
        meta: result.meta,
      } as DistributionAnalysisResult;
    } catch (err) {
      setError(err instanceof Error ? err.message : "distribution analysis request failed");
      return null;
    } finally {
      setLoading(false);
    }
  };

  return {
    loading,
    error,
    analyzeDistribution,
  };
};

