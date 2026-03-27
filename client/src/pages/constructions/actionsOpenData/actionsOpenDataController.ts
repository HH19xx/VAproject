import type { AnalysisContextResult, ExternalDataSource } from "../../../hooks/useWorldSignals";

export interface WorldSignalInputValues {
  externalDataSource: ExternalDataSource;
  locationKey: string;
  latitude: string;
  longitude: string;
  pastDays: string;
  forecastDays: string;
}

export interface ValidatedWorldSignalInputs {
  lat: number;
  lon: number;
  pDays: number;
  fDays: number;
}

export interface LoadAnalysisContextParams {
  fetchAnalysisContext: (
    source: ExternalDataSource,
    locationKey: string,
    signalType: string,
    from: string,
    to: string,
    limit?: number
  ) => Promise<AnalysisContextResult | null>;
  externalDataSource: ExternalDataSource;
  locationKey: string;
  effectiveExternalSignalType: string;
  from: string;
  to: string;
}

export interface LoadAnalysisContextResult {
  result: AnalysisContextResult | null;
  message: string;
}

export const validateWorldSignalInputs = (
  values: WorldSignalInputValues
): { values: ValidatedWorldSignalInputs | null; message?: string } => {
  if (values.externalDataSource === "e_stat_dashboard") {
    if (!values.locationKey.trim()) {
      return {
        values: null,
        message: "e-Stat Dashboard では location_key に都道府県コードが必要です。例: 13000",
      };
    }
    return {
      values: {
        lat: Number(values.latitude) || 0,
        lon: Number(values.longitude) || 0,
        pDays: Number(values.pastDays) || 0,
        fDays: Number(values.forecastDays) || 0,
      },
    };
  }

  if (!values.locationKey.trim()) {
    return { values: null, message: "location_key を入力してください。" };
  }

  const lat = Number(values.latitude);
  const lon = Number(values.longitude);
  const pDays = Number(values.pastDays);
  const fDays = Number(values.forecastDays);

  if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
    return { values: null, message: "latitude と longitude は数値で入力してください。" };
  }
  if (!Number.isFinite(pDays) || !Number.isFinite(fDays)) {
    return { values: null, message: "past_days と forecast_days は数値で入力してください。" };
  }

  return {
    values: { lat, lon, pDays, fDays },
  };
};

export const loadAnalysisContext = async ({
  fetchAnalysisContext,
  externalDataSource,
  locationKey,
  effectiveExternalSignalType,
  from,
  to,
}: LoadAnalysisContextParams): Promise<LoadAnalysisContextResult> => {
  if (!from || !to) {
    return {
      result: null,
      message: "分析コンテキストを取得するには from / to が必要です。",
    };
  }

  const result = await fetchAnalysisContext(
    externalDataSource,
    locationKey.trim(),
    effectiveExternalSignalType,
    new Date(from).toISOString(),
    new Date(to).toISOString(),
    200
  );

  return {
    result,
    message: result ? "分析コンテキストを更新しました。" : "分析コンテキストの取得に失敗しました。",
  };
};
