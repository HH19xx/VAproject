export type ExternalAxisKey =
  | "temperature_c"
  | "precipitation_mm"
  | "wind_speed_ms"
  | "weather_code"
  | "signal_value"
  | "observed_year";

export type ScatterAxisKey = ExternalAxisKey;

export const externalAxisLabelMap: Record<ExternalAxisKey, string> = {
  temperature_c: "気温",
  precipitation_mm: "降水量",
  wind_speed_ms: "風速",
  weather_code: "天気コード",
  signal_value: "指標値",
  observed_year: "観測年",
};

export const buildAvailableWorldAxisEntries = (
  source: "open_meteo" | "e_stat_dashboard"
): Array<[ScatterAxisKey, string]> => {
  if (source === "e_stat_dashboard") {
    return [
      ["observed_year", externalAxisLabelMap.observed_year],
      ["signal_value", externalAxisLabelMap.signal_value],
    ];
  }

  return [
    ["temperature_c", externalAxisLabelMap.temperature_c],
    ["precipitation_mm", externalAxisLabelMap.precipitation_mm],
    ["wind_speed_ms", externalAxisLabelMap.wind_speed_ms],
    ["weather_code", externalAxisLabelMap.weather_code],
  ];
};
