import { useContext, useState } from "react";
import { AuthContext } from "../contexts/AuthContext";

export type AnalysisSeverity = "OK" | "NOTICE" | "ALERT";

export interface DistributionStats {
  count: number;
  avg_tag_count: number;
  var_tag_count: number;
  prototype_rate: number;
}

export interface AnalysisSnapshot {
  id: number;
  user_id: number;
  query_text: string;
  severity: AnalysisSeverity;
  score: number;
  delta_avg_tag: number;
  delta_var_tag: number;
  delta_prototype_rate: number;
  p_value?: number;
  significant?: boolean;
  current: DistributionStats;
  baseline?: DistributionStats;
  created_at: string;
  create_user: string;
}

export interface CreateAnalysisSnapshotInput {
  query_text: string;
  severity: AnalysisSeverity;
  score: number;
  delta_avg_tag: number;
  delta_var_tag: number;
  delta_prototype_rate: number;
  p_value?: number;
  significant?: boolean;
  current: DistributionStats;
  baseline?: DistributionStats;
}

export const useAnalysisSnapshots = () => {
  const authContext = useContext(AuthContext);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) {
    throw new Error("useAnalysisSnapshots は AuthProvider 内で使用してください");
  }

  const { token } = authContext;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

  const authFetch = async (url: string, options: RequestInit = {}) => {
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
      ...(options.headers as HeadersInit),
    };
    const response = await fetch(url, { ...options, headers });
    if (!response.ok) {
      const errBody = await response.json().catch(() => ({}));
      throw new Error(errBody?.error?.message || "APIリクエストに失敗しました");
    }
    return response.json();
  };

  const listSnapshots = async (limit = 100): Promise<AnalysisSnapshot[]> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/analysis_snapshots?limit=${limit}`);
      if (!data.success) {
        throw new Error(data.message || "分析履歴の取得に失敗しました");
      }
      return (data.data?.snapshots || []) as AnalysisSnapshot[];
    } catch (err) {
      setError(err instanceof Error ? err.message : "分析履歴の取得に失敗しました");
      return [];
    } finally {
      setLoading(false);
    }
  };

  const createSnapshot = async (input: CreateAnalysisSnapshotInput): Promise<AnalysisSnapshot | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/analysis_snapshots`, {
        method: "POST",
        body: JSON.stringify(input),
      });
      if (!data.success) {
        throw new Error(data.message || "分析履歴の作成に失敗しました");
      }
      return data.data as AnalysisSnapshot;
    } catch (err) {
      setError(err instanceof Error ? err.message : "分析履歴の作成に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  const clearSnapshots = async (): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/analysis_snapshots`, {
        method: "DELETE",
      });
      if (!data.success) {
        throw new Error(data.message || "分析履歴の削除に失敗しました");
      }
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "分析履歴の削除に失敗しました");
      return false;
    } finally {
      setLoading(false);
    }
  };

  return {
    loading,
    error,
    listSnapshots,
    createSnapshot,
    clearSnapshots,
  };
};
