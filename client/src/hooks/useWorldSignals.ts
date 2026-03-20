import { useContext, useState } from "react";
import { AuthContext } from "../contexts/AuthContext";
import type { ActionLog } from "./useActions";

export interface WorldSignal {
  id: number;
  source: string;
  location_key: string;
  latitude: number;
  longitude: number;
  observed_at: string;
  temperature_c?: number;
  precipitation_mm?: number;
  wind_speed_ms?: number;
  weather_code?: number;
}

export interface AnalysisContextSummary {
  action_count: number;
  signal_count: number;
  avg_tags_per_action: number;
  avg_temperature_c?: number;
  total_precipitation_mm: number;
}

export interface AnalysisContextResult {
  from: string;
  to: string;
  location_key: string;
  action_logs: ActionLog[];
  world_signals: WorldSignal[];
  summary: AnalysisContextSummary;
  meta?: Record<string, unknown>;
}

export interface FetchWorldSignalInput {
  locationKey: string;
  latitude: number;
  longitude: number;
  pastDays: number;
  forecastDays: number;
}

export const useWorldSignals = () => {
  const authContext = useContext(AuthContext);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) {
    throw new Error("useWorldSignals は AuthProvider 内でのみ利用できます");
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

  const fetchOpenMeteo = async (input: FetchWorldSignalInput): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      const payload = {
        location_key: input.locationKey,
        latitude: input.latitude,
        longitude: input.longitude,
        past_days: input.pastDays,
        forecast_days: input.forecastDays,
      };
      const data = await authFetch(`${API_BASE_URL}/world_signals/fetch`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      if (!data.success) {
        throw new Error(data.message || "オープンデータ取得に失敗しました");
      }
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "オープンデータ取得に失敗しました");
      return false;
    } finally {
      setLoading(false);
    }
  };

  const fetchAnalysisContext = async (
    locationKey: string,
    fromISO: string,
    toISO: string,
    limit = 200
  ): Promise<AnalysisContextResult | null> => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams();
      params.set("location_key", locationKey);
      params.set("from", fromISO);
      params.set("to", toISO);
      params.set("limit", String(limit));
      const data = await authFetch(`${API_BASE_URL}/analysis/context?${params.toString()}`);
      if (!data.success) {
        throw new Error(data.message || "分析コンテキスト取得に失敗しました");
      }
      return data.data as AnalysisContextResult;
    } catch (err) {
      setError(err instanceof Error ? err.message : "分析コンテキスト取得に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  return {
    loading,
    error,
    fetchOpenMeteo,
    fetchAnalysisContext,
  };
};
