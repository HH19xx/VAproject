import type { AnalysisContextResult } from "../../../hooks/useWorldSignals";

export type ContextVizMode = "time" | "scatter";

export type ScatterAxisKey =
  | "temperature_c"
  | "precipitation_mm"
  | "wind_speed_ms"
  | "weather_code"
  | "signal_value"
  | "observed_year";

export const axisLabelMap: Record<ScatterAxisKey, string> = {
  temperature_c: "気温",
  precipitation_mm: "降水量",
  wind_speed_ms: "風速",
  weather_code: "天気コード",
  signal_value: "指標値",
  observed_year: "観測年",
};

export const worldSignalAxisValue = (
  signal: AnalysisContextResult["world_signals"][number],
  axis: ScatterAxisKey
): number | null => {
  if (axis === "observed_year") {
    return new Date(signal.observed_at).getFullYear();
  }
  const value = signal[axis];
  return typeof value === "number" ? value : null;
};

export const buildAvailableWorldAxisEntries = (
  source: "open_meteo" | "e_stat_dashboard"
): Array<[ScatterAxisKey, string]> => {
  if (source === "e_stat_dashboard") {
    return [
      ["observed_year", axisLabelMap.observed_year],
      ["signal_value", axisLabelMap.signal_value],
    ];
  }

  return [
    ["temperature_c", axisLabelMap.temperature_c],
    ["precipitation_mm", axisLabelMap.precipitation_mm],
    ["wind_speed_ms", axisLabelMap.wind_speed_ms],
    ["weather_code", axisLabelMap.weather_code],
  ];
};

export const buildContextTimePoints = (
  contextResult: AnalysisContextResult | null,
  source: "open_meteo" | "e_stat_dashboard"
) => {
  if (!contextResult || contextResult.world_signals.length === 0) return [];
  const valueAxis: ScatterAxisKey = source === "e_stat_dashboard" ? "signal_value" : "temperature_c";
  return contextResult.world_signals
    .filter((signal) => worldSignalAxisValue(signal, valueAxis) !== null)
    .map((signal) => ({
      t: new Date(signal.observed_at).getTime(),
      value: worldSignalAxisValue(signal, valueAxis) as number,
    }));
};

export const buildContextScatterPoints = (
  contextResult: AnalysisContextResult | null,
  xAxis: ScatterAxisKey,
  yAxis: ScatterAxisKey
) => {
  if (!contextResult || contextResult.world_signals.length === 0) return [];
  return contextResult.world_signals
    .filter((signal) => worldSignalAxisValue(signal, xAxis) !== null && worldSignalAxisValue(signal, yAxis) !== null)
    .map((signal) => ({
      x: worldSignalAxisValue(signal, xAxis) as number,
      y: worldSignalAxisValue(signal, yAxis) as number,
    }));
};

export const buildContextStaleState = (
  contextResult: AnalysisContextResult | null,
  current: {
    source: string;
    signalType: string;
    locationKey: string;
    from: string;
    to: string;
  },
  isSameDateTimeByMinute: (left: string, right: string) => boolean
) => {
  if (!contextResult || !current.from || !current.to) return false;
  const metaSource = String(contextResult.meta?.source || "open_meteo");
  const metaSignalType = String(contextResult.meta?.signal_type || "");
  return (
    contextResult.location_key !== current.locationKey ||
    metaSource !== current.source ||
    metaSignalType !== current.signalType ||
    !isSameDateTimeByMinute(contextResult.from, current.from) ||
    !isSameDateTimeByMinute(contextResult.to, current.to)
  );
};
