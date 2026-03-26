import { useEffect, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useActions, type ActionAttribute, type ActionLog } from "../../hooks/useActions";
import useSearchTagSuggestions from "../../hooks/useSearchTagSuggestions";
import { useTags } from "../../hooks/useTags";
import {
  useWorldSignals,
  type ExternalDataSource,
  type FetchWorldSignalInput,
  type WorldSignal,
} from "../../hooks/useWorldSignals";
import { eStatDashboardSignalOptions } from "../../constants/externalDataCatalog";
import SearchSuggestionRail from "../../components/search/SearchSuggestionRail";
import {
  buildTagNameToID,
  parseDanbooruStyleQuery,
} from "../constructions/components/actionsSearchSectionHelpers";
import { appendTagToQueryText } from "../constructions/components/actionsSearchController";
import styles from "../../assets/styles/StatisticsAnalysis.module.scss";

type DatasetKind = "action_logs" | "world_signals";
type ActionAxisKey = "tag_count" | "prototype_id" | string;
type WorldAxisKey =
  | "temperature_c"
  | "precipitation_mm"
  | "wind_speed_ms"
  | "weather_code"
  | "signal_value"
  | "observed_year";
type RepeatBucketUnit = "day" | "week";

type SummaryStats = {
  count: number;
  mean: number;
  variance: number;
  stdDev: number;
};

type TwoGroupAnalysisResult = {
  groupA: SummaryStats;
  groupB: SummaryStats;
  meanDiff: number;
  tStatistic: number;
  degreesOfFreedom: number;
  pValueApprox: number;
};

type CorrelationAnalysisResult = {
  count: number;
  correlation: number;
  covariance: number;
  pValueApprox: number;
};

type RepeatBehaviorResult = {
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

type StatisticsNavigationState = {
  dataset?: DatasetKind;
  danbooruQuery?: string;
  from?: string;
  to?: string;
  filterTagIDs?: number[];
};

const DAY_MS = 24 * 60 * 60 * 1000;
const RECENT_ACTION_LOOKBACK_DAYS = 90;
const ACTION_COMPARE_WINDOW_DAYS = 14;
const ACTION_CORRELATION_WINDOW_DAYS = 28;
const ESTAT_COMPARE_YEARS = 7;
const ESTAT_CORRELATION_YEARS = 14;

const toDateTimeLocal = (date: Date) =>
  new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);

const createRange = (startDaysAgo: number, endDaysAgo: number) => ({
  from: toDateTimeLocal(new Date(Date.now() - startDaysAgo * DAY_MS)),
  to: toDateTimeLocal(new Date(Date.now() - endDaysAgo * DAY_MS)),
});

const createYearRange = (startYearsAgo: number, endYearsAgo: number) => ({
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
const formatNumber = (value: number, digits = 3) => value.toFixed(digits);

const repeatBucketLabel = (unit: RepeatBucketUnit) =>
  unit === "week" ? "週単位" : "日単位";

const findAttributeValue = (attributes: ActionAttribute[] | undefined, key: string) =>
  attributes?.find((attribute) => attribute.key === key)?.value_number;

const actionValue = (log: ActionLog, axis: ActionAxisKey) => {
  if (axis === "tag_count") return log.tag_ids.length;
  if (axis === "prototype_id") return log.prototype_id ?? null;
  return findAttributeValue(log.attributes, axis) ?? null;
};

const worldValue = (signal: WorldSignal, axis: WorldAxisKey) => {
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


const collectActionAxisOptions = (logs: ActionLog[]) => {
  const keys = new Set<string>(["tag_count", "prototype_id"]);
  logs.forEach((log) => {
    log.attributes?.forEach((attribute) => {
      if (attribute.key) keys.add(attribute.key);
    });
  });
  return Array.from(keys);
};

const actionAxisLabel = (axis: ActionAxisKey) => {
  if (axis === "tag_count") return "タグ数";
  if (axis === "prototype_id") return "プロトタイプID";
  return axis;
};

const computeWelchTest = (
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

const computeCorrelation = (
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
    pValueApprox:
      denominator <= 0 ? 0 : 2 * (1 - normalCdf(Math.abs(tStatistic))),
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

const computeRepeatBehavior = (
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

const StatisticsAnalysis = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const navigationState = (location.state ?? {}) as StatisticsNavigationState;
  const { fetchActionsSnapshot } = useActions();
  const { tags, fetchTags } = useTags();
  const { fetchOpenMeteo, fetchAnalysisContext, loading, error } = useWorldSignals();

  const actionCompareRangeA = useMemo(() => createRange(ACTION_COMPARE_WINDOW_DAYS, 7), []);
  const actionCompareRangeB = useMemo(() => createRange(7, 0), []);
  const actionCorrelationRange = useMemo(() => createRange(ACTION_CORRELATION_WINDOW_DAYS, 0), []);
  const actionRepeatRange = useMemo(() => createRange(30, 0), []);
  const eStatCompareRangeA = useMemo(
    () => createYearRange(ESTAT_COMPARE_YEARS * 2, ESTAT_COMPARE_YEARS),
    []
  );
  const eStatCompareRangeB = useMemo(() => createYearRange(ESTAT_COMPARE_YEARS, 0), []);
  const eStatCorrelationRange = useMemo(() => createYearRange(ESTAT_CORRELATION_YEARS, 0), []);

  const [dataset, setDataset] = useState<DatasetKind>(navigationState.dataset ?? "action_logs");
  const [source, setSource] = useState<ExternalDataSource>("open_meteo");
  const [locationKey, setLocationKey] = useState("tokyo_shinjuku");
  const [latitude, setLatitude] = useState("35.6895");
  const [longitude, setLongitude] = useState("139.6917");
  const [pastDays, setPastDays] = useState("14");
  const [forecastDays, setForecastDays] = useState("1");
  const [signalType, setSignalType] = useState("population_total");
  const [actionKeyword, setActionKeyword] = useState(navigationState.danbooruQuery ?? "");
  const [selectedTagIDs] = useState<number[]>(navigationState.filterTagIDs ?? []);

  const [compareFromA, setCompareFromA] = useState(navigationState.from ?? actionCompareRangeA.from);
  const [compareToA, setCompareToA] = useState(navigationState.to ?? actionCompareRangeA.to);
  const [compareFromB, setCompareFromB] = useState(actionCompareRangeB.from);
  const [compareToB, setCompareToB] = useState(actionCompareRangeB.to);
  const [correlationFrom, setCorrelationFrom] = useState(navigationState.from ?? actionCorrelationRange.from);
  const [correlationTo, setCorrelationTo] = useState(navigationState.to ?? actionCorrelationRange.to);
  const [repeatKeyword, setRepeatKeyword] = useState(navigationState.danbooruQuery ?? "");
  const [repeatFrom, setRepeatFrom] = useState(actionRepeatRange.from);
  const [repeatTo, setRepeatTo] = useState(actionRepeatRange.to);
  const [repeatUnit, setRepeatUnit] = useState<RepeatBucketUnit>("day");
  const [repeatExpected, setRepeatExpected] = useState("1");

  const [actionAxisOptions, setActionAxisOptions] = useState<ActionAxisKey[]>([
    "tag_count",
    "prototype_id",
  ]);
  const [searchedActionLogs, setSearchedActionLogs] = useState<ActionLog[]>([]);
  const [hasSearchedActionLogs, setHasSearchedActionLogs] = useState(false);
  const [compareActionAxis, setCompareActionAxis] = useState<ActionAxisKey>("tag_count");
  const [correlationActionXAxis, setCorrelationActionXAxis] = useState<ActionAxisKey>("tag_count");
  const [correlationActionYAxis, setCorrelationActionYAxis] =
    useState<ActionAxisKey>("prototype_id");
  const [compareWorldAxis, setCompareWorldAxis] = useState<WorldAxisKey>("temperature_c");
  const [correlationWorldXAxis, setCorrelationWorldXAxis] =
    useState<WorldAxisKey>("temperature_c");
  const [correlationWorldYAxis, setCorrelationWorldYAxis] =
    useState<WorldAxisKey>("precipitation_mm");

  const [compareResult, setCompareResult] = useState<TwoGroupAnalysisResult | null>(null);
  const [correlationResult, setCorrelationResult] = useState<CorrelationAnalysisResult | null>(
    null
  );
  const [repeatResult, setRepeatResult] = useState<RepeatBehaviorResult | null>(null);
  const [pageMessage, setPageMessage] = useState(
    "まだ分析を実行していません。"
  );
  const [pageError, setPageError] = useState<string | null>(null);
  const tagNameToID = useMemo(() => buildTagNameToID(tags), [tags]);
  const actionSearchSuggestions = useSearchTagSuggestions(tags, actionKeyword, selectedTagIDs, 12);

  useEffect(() => {
    void fetchTags();
  }, []);

  useEffect(() => {
    const loadRecentActions = async () => {
      const range = createRange(RECENT_ACTION_LOOKBACK_DAYS, 0);
      const snapshot = await fetchActionsSnapshot(1, 300, [], {
        from: new Date(range.from).toISOString(),
        to: new Date(range.to).toISOString(),
        sort: "occurred_at",
        order: "desc",
      });
      if (!snapshot) return;

      const options = collectActionAxisOptions(snapshot.logs);
      setActionAxisOptions(options);
      if (!options.includes(compareActionAxis)) setCompareActionAxis("tag_count");
      if (!options.includes(correlationActionXAxis)) setCorrelationActionXAxis("tag_count");
      if (!options.includes(correlationActionYAxis)) {
        setCorrelationActionYAxis(options.includes("prototype_id") ? "prototype_id" : "tag_count");
      }
    };
    void loadRecentActions();
  }, []);

  useEffect(() => {
    setPageError(null);
    setPageMessage(
      "まだ分析を実行していません。"
    );
    setHasSearchedActionLogs(false);
    setSearchedActionLogs([]);
    setCompareResult(null);
    setCorrelationResult(null);
    setRepeatResult(null);
  }, [dataset, source]);

  useEffect(() => {
    if (source === "e_stat_dashboard") {
      if (locationKey === "tokyo_shinjuku") setLocationKey("13000");
      setLatitude("0");
      setLongitude("0");
      setCompareWorldAxis("signal_value");
      setCorrelationWorldXAxis("observed_year");
      setCorrelationWorldYAxis("signal_value");
      setCompareFromA(eStatCompareRangeA.from);
      setCompareToA(eStatCompareRangeA.to);
      setCompareFromB(eStatCompareRangeB.from);
      setCompareToB(eStatCompareRangeB.to);
      setCorrelationFrom(eStatCorrelationRange.from);
      setCorrelationTo(eStatCorrelationRange.to);
      return;
    }
    if (locationKey === "13000") setLocationKey("tokyo_shinjuku");
    if (latitude === "0") setLatitude("35.6895");
    if (longitude === "0") setLongitude("139.6917");
    setCompareWorldAxis("temperature_c");
    setCorrelationWorldXAxis("temperature_c");
    setCorrelationWorldYAxis("precipitation_mm");
    setCompareFromA(actionCompareRangeA.from);
    setCompareToA(actionCompareRangeA.to);
    setCompareFromB(actionCompareRangeB.from);
    setCompareToB(actionCompareRangeB.to);
    setCorrelationFrom(actionCorrelationRange.from);
    setCorrelationTo(actionCorrelationRange.to);
  }, [source]);

  const currentSignalType = source === "e_stat_dashboard" ? signalType : "";

  const actionSearchWindow = useMemo(() => {
    const timestamps = [
      compareFromA,
      compareToA,
      compareFromB,
      compareToB,
      correlationFrom,
      correlationTo,
      repeatFrom,
      repeatTo,
    ]
      .map((value) => new Date(value).getTime())
      .filter((value) => Number.isFinite(value));

    if (timestamps.length === 0) return null;

    return {
      from: new Date(Math.min(...timestamps)).toISOString(),
      to: new Date(Math.max(...timestamps)).toISOString(),
    };
  }, [compareFromA, compareToA, compareFromB, compareToB, correlationFrom, correlationTo, repeatFrom, repeatTo]);

  const correlationPreviewPairs = useMemo(() => {
    if (searchedActionLogs.length === 0) return 0;
    const fromTime = new Date(correlationFrom).getTime();
    const toTime = new Date(correlationTo).getTime();

    return searchedActionLogs
      .filter((log) => {
        const occurredAt = new Date(log.occurred_at).getTime();
        if (!Number.isFinite(occurredAt)) return false;
        return occurredAt >= fromTime && occurredAt <= toTime;
      })
      .map((log) => ({
        x: actionValue(log, correlationActionXAxis),
        y: actionValue(log, correlationActionYAxis),
      }))
      .filter(
        (pair) =>
          typeof pair.x === "number" &&
          Number.isFinite(pair.x) &&
          typeof pair.y === "number" &&
          Number.isFinite(pair.y)
      ).length;
  }, [searchedActionLogs, correlationFrom, correlationTo, correlationActionXAxis, correlationActionYAxis]);

  const appendActionSearchSuggestion = (tagName: string) => {
    const next = appendTagToQueryText(actionKeyword, tagName);
    if (!next.alreadyIncluded) {
      setActionKeyword(next.updatedQuery);
    }
  };

  const handleSearchActionLogs = async () => {
    setPageError(null);
    setPageMessage("è¡Œå‹•è¨˜éŒ²ã‚’æ¤œç´¢ã—ã¦ã„ã¾ã™â€¦");

    const parsed = parseDanbooruStyleQuery(actionKeyword, tagNameToID);
    if (parsed.unknownTokens.length > 0) {
      setPageError("æœªç™»éŒ²ã‚¿ã‚°ãŒã‚ã‚Šã¾ã™: " + parsed.unknownTokens.join(", "));
      return;
    }
    if (!actionSearchWindow) {
      setPageError("æ¤œç´¢ã«å¿…è¦ãªæœŸé–“ãŒä¸æ­£ã§ã™ã€‚");
      return;
    }

    const mergedTagIDs = Array.from(new Set([...selectedTagIDs, ...parsed.andTagIDs]));
    const snapshot = await fetchActionsSnapshot(1, 500, mergedTagIDs, {
      from: actionSearchWindow.from,
      to: actionSearchWindow.to,
      sort: "occurred_at",
      order: "desc",
      anyTagIDs: parsed.anyTagIDs,
      anyTagGroups: parsed.anyTagGroups,
      excludeTagIDs: parsed.excludeTagIDs,
    });
    if (!snapshot) return;

    setHasSearchedActionLogs(true);
    setSearchedActionLogs(snapshot.logs);
    const options = collectActionAxisOptions(snapshot.logs);
    setActionAxisOptions(options);
    if (!options.includes(compareActionAxis)) setCompareActionAxis("tag_count");
    if (!options.includes(correlationActionXAxis)) setCorrelationActionXAxis("tag_count");
    if (!options.includes(correlationActionYAxis)) {
      setCorrelationActionYAxis(options.includes("prototype_id") ? "prototype_id" : "tag_count");
    }

    setPageMessage("è¡Œå‹•è¨˜éŒ²ã®æ¤œç´¢çµæžœã‚’æ›´æ–°ã—ã¾ã—ãŸã€‚ä»¶æ•°=" + snapshot.logs.length);
  };

  const fetchActionValues = async (from: string, to: string, axis: ActionAxisKey) => {
    const parsed = parseDanbooruStyleQuery(actionKeyword, tagNameToID);
    if (parsed.unknownTokens.length > 0) {
      setPageError("未登録タグがあります: " + parsed.unknownTokens.join(", "));
      return null;
    }

    const mergedTagIDs = Array.from(new Set([...selectedTagIDs, ...parsed.andTagIDs]));
    const snapshot = await fetchActionsSnapshot(1, 500, mergedTagIDs, {
      from: new Date(from).toISOString(),
      to: new Date(to).toISOString(),
      sort: "occurred_at",
      order: "desc",
      anyTagIDs: parsed.anyTagIDs,
      anyTagGroups: parsed.anyTagGroups,
      excludeTagIDs: parsed.excludeTagIDs,
    });
    if (!snapshot) return null;

    const values = snapshot.logs
      .map((log) => actionValue(log, axis))
      .filter((value): value is number => typeof value === "number" && Number.isFinite(value));

    setActionAxisOptions(collectActionAxisOptions(snapshot.logs));
    return { logs: snapshot.logs, values };
  };

  const fetchWorldValues = async (from: string, to: string, axis: WorldAxisKey) => {
    const context = await fetchAnalysisContext(
      source,
      locationKey.trim(),
      currentSignalType,
      new Date(from).toISOString(),
      new Date(to).toISOString(),
      500
    );
    if (!context) return null;

    const values = context.world_signals
      .map((signal) => worldValue(signal, axis))
      .filter((value): value is number => typeof value === "number" && Number.isFinite(value));

    return { signals: context.world_signals, values };
  };

  const handleFetchExternalData = async () => {
    setPageError(null);
    setPageMessage(
      "外部データを取得しています…"
    );

    const payload: FetchWorldSignalInput = {
      source,
      locationKey: locationKey.trim(),
      latitude: source === "open_meteo" ? Number(latitude) : 0,
      longitude: source === "open_meteo" ? Number(longitude) : 0,
      pastDays: source === "open_meteo" ? Number(pastDays) : 0,
      forecastDays: source === "open_meteo" ? Number(forecastDays) : 0,
    };

    if (!payload.locationKey) {
      setPageError(
        source === "e_stat_dashboard"
          ? "e-Stat Dashboard では都道府県コードを入力してください。例: 13000"
          : "location_key を入力してください。"
      );
      return;
    }

    if (
      source === "open_meteo" &&
      (!Number.isFinite(payload.latitude) ||
        !Number.isFinite(payload.longitude) ||
        !Number.isFinite(payload.pastDays) ||
        !Number.isFinite(payload.forecastDays))
    ) {
      setPageError(
        "Open-Meteo では緯度・経度・past_days・forecast_days を数値で入力してください。"
      );
      return;
    }

    const ok = await fetchOpenMeteo(payload);
    if (!ok) {
      setPageError(
        "外部データの取得に失敗しました。API設定と入力値を確認してください。"
      );
      return;
    }

    const selectedLabel =
      eStatDashboardSignalOptions.find((option) => option.value === signalType)?.label ??
      signalType;
    setPageMessage(
      source === "e_stat_dashboard"
        ? "e-Stat Dashboard の " + selectedLabel + " を取得しました。"
        : "Open-Meteo の外部データを取得しました。"
    );
  };

  const handleCompareMeans = async () => {
    setPageError(null);
    setPageMessage(
      "平均値比較を実行しています…"
    );
    setCompareResult(null);

    if (dataset === "action_logs") {
      const groupA = await fetchActionValues(compareFromA, compareToA, compareActionAxis);
      const groupB = await fetchActionValues(compareFromB, compareToB, compareActionAxis);
      if (!groupA || !groupB) return;

      const result = computeWelchTest(groupA.values, groupB.values);
      if (!result) {
        setPageError(
          "2期間比較に必要な数値データが不足しています。各期間でで2件以上の数値が必要です。"
        );
        return;
      }

      setCompareResult(result);
      setPageMessage(
        "行動記録の平均値比較を更新しました。対象件数 A=" + groupA.values.length + " / B=" + groupB.values.length
      );
      return;
    }

    const groupA = await fetchWorldValues(compareFromA, compareToA, compareWorldAxis);
    const groupB = await fetchWorldValues(compareFromB, compareToB, compareWorldAxis);
    if (!groupA || !groupB) return;

    const result = computeWelchTest(groupA.values, groupB.values);
    if (!result) {
      setPageError(
        source === "e_stat_dashboard"
          ? "e-Stat は年次データなので、各期間に複数年が入るよう期間を広げてください。"
          : "2期間比較に必要な外部データが不足しています。各期間でで2件以上の数値が必要です。"
      );
      return;
    }

    setCompareResult(result);
    setPageMessage(
      "外部データの平均値比較を更新しました。対象件数 A=" + groupA.values.length + " / B=" + groupB.values.length
    );
  };

  const handleCorrelation = async () => {
    setPageError(null);
    setPageMessage(
      "相関分析を実行しています…"
    );
    setCorrelationResult(null);

    if (dataset === "action_logs") {
      const actionData = await fetchActionValues(correlationFrom, correlationTo, correlationActionXAxis);
      if (!actionData) return;

      const pairs = actionData.logs
        .map((log) => ({
          x: actionValue(log, correlationActionXAxis),
          y: actionValue(log, correlationActionYAxis),
        }))
        .filter(
          (pair): pair is { x: number; y: number } =>
            typeof pair.x === "number" &&
            Number.isFinite(pair.x) &&
            typeof pair.y === "number" &&
            Number.isFinite(pair.y)
        );

      const result = computeCorrelation(pairs);
      if (!result) {
        setPageError(
          "相関分析には、同一期間内で3件以上の数値ペアが必要です。"
        );
        return;
      }

      setCorrelationResult(result);
      setPageMessage(
        "行動記録の相関分析を更新しました。対象ペア数=" + result.count
      );
      return;
    }

    const context = await fetchAnalysisContext(
      source,
      locationKey.trim(),
      currentSignalType,
      new Date(correlationFrom).toISOString(),
      new Date(correlationTo).toISOString(),
      500
    );
    if (!context) return;

    const pairs = context.world_signals
      .map((signal) => ({
        x: worldValue(signal, correlationWorldXAxis),
        y: worldValue(signal, correlationWorldYAxis),
      }))
      .filter(
        (pair): pair is { x: number; y: number } =>
          typeof pair.x === "number" &&
          Number.isFinite(pair.x) &&
          typeof pair.y === "number" &&
          Number.isFinite(pair.y)
      );

    const result = computeCorrelation(pairs);
    if (!result) {
      setPageError(
        "相関分析には、同一期間内で3件以上の数値ペアが必要です。"
      );
      return;
    }

    setCorrelationResult(result);
    setPageMessage(
      "外部データの相関分析を更新しました。対象ペア数=" + result.count
    );
  };

  const handleRepeatBehavior = async () => {
    setPageError(null);
    setPageMessage(
      "反復行動の推定と検定を実行しています…"
    );
    setRepeatResult(null);

    const actionData = await fetchActionValues(repeatFrom, repeatTo, "tag_count");
    if (!actionData) return;

    const expected = Number(repeatExpected);
    if (!Number.isFinite(expected)) {
      setPageError(
        "期待値には数値を入力してください。"
      );
      return;
    }

    const result = computeRepeatBehavior(actionData.logs, repeatKeyword, repeatUnit, expected);
    if (!result) {
      setPageError(
        "対象キーワードに一致する行動記録が見つかりませんでした。"
      );
      return;
    }

    setRepeatResult(result);
    setPageMessage(
      "反復推定を更新しました。対象件数=" + result.matchedActionCount + " / バケット数=" + result.bucketCount
    );
  };

  const selectedSignalLabel =
    eStatDashboardSignalOptions.find((option) => option.value === signalType)?.label ?? signalType;

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>
            {"Quantified Self機能"}
          </h1>
          <p className={styles.subtitle}>
            {"推定と検定を中心に、行動記録と外部データを比較します。"}
          </p>
        </div>
        <button className={styles.secondaryButton} onClick={() => navigate("/dashboard")}>
          {"ダッシュボードへ戻る"}
        </button>
      </div>

      <section className={styles.panel}>
        <h2 className={styles.sectionTitle}>{"分析対象"}</h2>
        <div className={styles.formGrid}>
          <label className={styles.field}>
            <span>{"データ系"}</span>
            <select
              value={dataset}
              onChange={(event) => setDataset(event.target.value as DatasetKind)}
            >
              <option value="action_logs">{"行動記録"}</option>
              <option value="world_signals">{"外部データ"}</option>
            </select>
          </label>

          {dataset === "world_signals" ? (
            <>
              <label className={styles.field}>
                <span>データソース</span>
                <select
                  value={source}
                  onChange={(event) => setSource(event.target.value as ExternalDataSource)}
                >
                  <option value="open_meteo">Open-Meteo</option>
                  <option value="e_stat_dashboard">e-Stat Dashboard</option>
                </select>
              </label>

              <label className={styles.field}>
                <span>location_key</span>
                <input
                  value={locationKey}
                  onChange={(event) => setLocationKey(event.target.value)}
                />
              </label>

              {source === "e_stat_dashboard" ? (
                <label className={styles.field}>
                  <span>指標</span>
                  <select
                    value={signalType}
                    onChange={(event) => setSignalType(event.target.value)}
                  >
                    {eStatDashboardSignalOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </label>
              ) : (
                <>
                  <label className={styles.field}>
                    <span>latitude</span>
                    <input
                      value={latitude}
                      onChange={(event) => setLatitude(event.target.value)}
                    />
                  </label>
                  <label className={styles.field}>
                    <span>longitude</span>
                    <input
                      value={longitude}
                      onChange={(event) => setLongitude(event.target.value)}
                    />
                  </label>
                  <label className={styles.field}>
                    <span>past_days</span>
                    <input
                      value={pastDays}
                      onChange={(event) => setPastDays(event.target.value)}
                    />
                  </label>
                  <label className={styles.field}>
                    <span>forecast_days</span>
                    <input
                      value={forecastDays}
                      onChange={(event) => setForecastDays(event.target.value)}
                    />
                  </label>
                </>
              )}
            </>
          ) : (
            <div className={styles.field}>
              <span>Danbooru形式検索クエリ</span>
              <input
                value={actionKeyword}
                onChange={(event) => setActionKeyword(event.target.value)}
                placeholder="例: 物価 弁当 / 交通 山手線 / 昼_弁当 / 価格帯=高 / 交通 山手線 | 物価 昼_弁当"
              />
              <SearchSuggestionRail
                title={actionKeyword.trim() ? "入力候補" : "候補タグ"}
                suggestions={actionSearchSuggestions}
                emptyMessage={
                  actionKeyword.trim()
                    ? "一致するタグはありません。"
                    : "候補に出せるタグがまだありません。"
                }
                onSelectSuggestion={appendActionSearchSuggestion}
              />
              <div className={styles.description}>
                半角空白区切りです。まず検索して対象集合を確定し、その検索結果から軸を選んで分析します。`A B | C D` は `(A B)~(C D)`、つまり `(A AND B) OR (C AND D)` を意味します。タグ名に半角スペースを入れたい場合は `_` を使ってください。
              </div>
              <div className={styles.actions}>
                <button
                  type="button"
                  className={styles.primaryButton}
                  onClick={handleSearchActionLogs}
                  disabled={loading}
                >
                  記録を検索
                </button>
              </div>
              {hasSearchedActionLogs && (
                <div className={styles.notice}>
                  <div>検索結果件数: {searchedActionLogs.length}</div>
                  <div>選択可能な軸: {actionAxisOptions.map((axis) => actionAxisLabel(axis)).join(" / ")}</div>
                  <div>現在の相関対象ペア数: {correlationPreviewPairs}</div>
                </div>
              )}
            </div>
          )}
        </div>

        {dataset === "world_signals" && (
          <>
            <p className={styles.description}>
              統計分析ページでも外部データを取得できます。e-Stat Dashboard では都道府県コードを使います。
            </p>
            <div className={styles.actions}>
              <button
                className={styles.primaryButton}
                onClick={handleFetchExternalData}
                disabled={loading}
              >
                外部データを取得
              </button>
            </div>
          </>
        )}

        {dataset === "action_logs" && (
          <p className={styles.notice}>
            テストデータは developer / tester に投入されます。検索後に件数と軸候補を確認してから分析してください。
          </p>
        )}
      </section>

      <div className={styles.sectionGrid}>
        <section className={styles.panel}>
          <h2 className={styles.sectionTitle}>
            {"2期間の平均値比較"}
          </h2>
          <div className={styles.formGrid}>
            <label className={styles.field}>
              <span>{"期間A 開始"}</span>
              <input
                type="datetime-local"
                value={compareFromA}
                onChange={(event) => setCompareFromA(event.target.value)}
              />
            </label>
            <label className={styles.field}>
              <span>{"期間A 終了"}</span>
              <input
                type="datetime-local"
                value={compareToA}
                onChange={(event) => setCompareToA(event.target.value)}
              />
            </label>
            <label className={styles.field}>
              <span>{"期間B 開始"}</span>
              <input
                type="datetime-local"
                value={compareFromB}
                onChange={(event) => setCompareFromB(event.target.value)}
              />
            </label>
            <label className={styles.field}>
              <span>{"期間B 終了"}</span>
              <input
                type="datetime-local"
                value={compareToB}
                onChange={(event) => setCompareToB(event.target.value)}
              />
            </label>
            <label className={styles.field}>
              <span>{"比較軸"}</span>
              {dataset === "action_logs" ? (
                <select
                  value={compareActionAxis}
                  onChange={(event) => setCompareActionAxis(event.target.value)}
                >
                  {actionAxisOptions.map((axis) => (
                    <option key={axis} value={axis}>
                      {actionAxisLabel(axis)}
                    </option>
                  ))}
                </select>
              ) : (
                <select
                  value={compareWorldAxis}
                  onChange={(event) =>
                    setCompareWorldAxis(event.target.value as WorldAxisKey)
                  }
                >
                  {source === "e_stat_dashboard" ? (
                    <option value="signal_value">{selectedSignalLabel}</option>
                  ) : (
                    <>
                      <option value="temperature_c">{"気温"}</option>
                      <option value="precipitation_mm">{"降水量"}</option>
                      <option value="wind_speed_ms">{"風速"}</option>
                      <option value="weather_code">{"天気コード"}</option>
                    </>
                  )}
                </select>
              )}
            </label>
          </div>
          <div className={styles.actions}>
            <button
              className={styles.primaryButton}
              onClick={handleCompareMeans}
              disabled={loading}
            >
              {"平均値比較を実行"}
            </button>
          </div>

          {compareResult && (
            <div className={styles.resultGrid}>
              <div className={styles.resultCard}>
                <h3>{"期間A"}</h3>
                <p>{"件数"}: {compareResult.groupA.count}</p>
                <p>{"平均"}: {formatNumber(compareResult.groupA.mean)}</p>
                <p>{"標準偶差"}: {formatNumber(compareResult.groupA.stdDev)}</p>
              </div>
              <div className={styles.resultCard}>
                <h3>{"期間B"}</h3>
                <p>{"件数"}: {compareResult.groupB.count}</p>
                <p>{"平均"}: {formatNumber(compareResult.groupB.mean)}</p>
                <p>{"標準偶差"}: {formatNumber(compareResult.groupB.stdDev)}</p>
              </div>
              <div className={styles.resultCard}>
                <h3>{"比較結果"}</h3>
                <p>{"平均差"}: {formatNumber(compareResult.meanDiff)}</p>
                <p>{"t値"}: {formatNumber(compareResult.tStatistic)}</p>
                <p>{"自由度"}: {formatNumber(compareResult.degreesOfFreedom, 1)}</p>
                <p>{"両側p値"}: {formatNumber(compareResult.pValueApprox, 4)}</p>
              </div>
            </div>
          )}
        </section>

        <section className={styles.panel}>
          <h2 className={styles.sectionTitle}>{"相関分析"}</h2>
          <div className={styles.formGrid}>
            <label className={styles.field}>
              <span>{"開始"}</span>
              <input
                type="datetime-local"
                value={correlationFrom}
                onChange={(event) => setCorrelationFrom(event.target.value)}
              />
            </label>
            <label className={styles.field}>
              <span>{"終了"}</span>
              <input
                type="datetime-local"
                value={correlationTo}
                onChange={(event) => setCorrelationTo(event.target.value)}
              />
            </label>
            <label className={styles.field}>
              <span>{"X軸"}</span>
              {dataset === "action_logs" ? (
                <select
                  value={correlationActionXAxis}
                  onChange={(event) => setCorrelationActionXAxis(event.target.value)}
                >
                  {actionAxisOptions.map((axis) => (
                    <option key={axis} value={axis}>
                      {actionAxisLabel(axis)}
                    </option>
                  ))}
                </select>
              ) : (
                <select
                  value={correlationWorldXAxis}
                  onChange={(event) =>
                    setCorrelationWorldXAxis(event.target.value as WorldAxisKey)
                  }
                >
                  {source === "e_stat_dashboard" ? (
                    <>
                      <option value="observed_year">{"年"}</option>
                      <option value="signal_value">{selectedSignalLabel}</option>
                    </>
                  ) : (
                    <>
                      <option value="temperature_c">{"気温"}</option>
                      <option value="precipitation_mm">{"降水量"}</option>
                      <option value="wind_speed_ms">{"風速"}</option>
                      <option value="weather_code">{"天気コード"}</option>
                    </>
                  )}
                </select>
              )}
            </label>
            <label className={styles.field}>
              <span>{"Y軸"}</span>
              {dataset === "action_logs" ? (
                <select
                  value={correlationActionYAxis}
                  onChange={(event) => setCorrelationActionYAxis(event.target.value)}
                >
                  {actionAxisOptions.map((axis) => (
                    <option key={axis} value={axis}>
                      {actionAxisLabel(axis)}
                    </option>
                  ))}
                </select>
              ) : (
                <select
                  value={correlationWorldYAxis}
                  onChange={(event) =>
                    setCorrelationWorldYAxis(event.target.value as WorldAxisKey)
                  }
                >
                  {source === "e_stat_dashboard" ? (
                    <>
                      <option value="signal_value">{selectedSignalLabel}</option>
                      <option value="observed_year">{"年"}</option>
                    </>
                  ) : (
                    <>
                      <option value="temperature_c">{"気温"}</option>
                      <option value="precipitation_mm">{"降水量"}</option>
                      <option value="wind_speed_ms">{"風速"}</option>
                      <option value="weather_code">{"天気コード"}</option>
                    </>
                  )}
                </select>
              )}
            </label>
          </div>
          <div className={styles.actions}>
            <button
              className={styles.primaryButton}
              onClick={handleCorrelation}
              disabled={loading}
            >
              {"相関分析を実行"}
            </button>
          </div>

          {correlationResult && (
            <div className={styles.resultGrid}>
              <div className={styles.resultCard}>
                <h3>{"相関係数"}</h3>
                <p>{formatNumber(correlationResult.correlation, 4)}</p>
              </div>
              <div className={styles.resultCard}>
                <h3>{"共分散"}</h3>
                <p>{formatNumber(correlationResult.covariance, 4)}</p>
              </div>
              <div className={styles.resultCard}>
                <h3>{"概要"}</h3>
                <p>{"対象ペア数"}: {correlationResult.count}</p>
                <p>{"両側p値"}: {formatNumber(correlationResult.pValueApprox, 4)}</p>
              </div>
            </div>
          )}
        </section>
      </div>

      <section className={styles.panel}>
        <h2 className={styles.sectionTitle}>
          {"反復行動の推定と検定"}
        </h2>
        <p className={styles.description}>
          {"同じ行動が繰り返されると想定した場合の、期間あたりの平均期待値を推定します。"}
        </p>
        <div className={styles.formGrid}>
          <label className={styles.field}>
            <span>{"対象キーワード"}</span>
            <input
              value={repeatKeyword}
              onChange={(event) => setRepeatKeyword(event.target.value)}
              placeholder={"例: 通勤 / 運動 / 購買"}
            />
          </label>
          <label className={styles.field}>
            <span>{"開始"}</span>
            <input
              type="datetime-local"
              value={repeatFrom}
              onChange={(event) => setRepeatFrom(event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>{"終了"}</span>
            <input
              type="datetime-local"
              value={repeatTo}
              onChange={(event) => setRepeatTo(event.target.value)}
            />
          </label>
          <label className={styles.field}>
            <span>{"集計単位"}</span>
            <select
              value={repeatUnit}
              onChange={(event) => setRepeatUnit(event.target.value as RepeatBucketUnit)}
            >
              <option value="day">{"日"}</option>
              <option value="week">{"週"}</option>
            </select>
          </label>
          <label className={styles.field}>
            <span>
              {"期待値 / "}{repeatBucketLabel(repeatUnit)}
            </span>
            <input
              value={repeatExpected}
              onChange={(event) => setRepeatExpected(event.target.value)}
            />
          </label>
        </div>
        <div className={styles.actions}>
          <button
            className={styles.primaryButton}
            onClick={handleRepeatBehavior}
            disabled={loading}
          >
            {"反復推定を実行"}
          </button>
        </div>

        {repeatResult && (
          <div className={styles.resultGrid}>
            <div className={styles.resultCard}>
              <h3>{"要約結果"}</h3>
              <p>{"一致した記録数"}: {repeatResult.matchedActionCount}</p>
              <p>{"バケット数"}: {repeatResult.bucketCount}</p>
              <p>{"平均値"}: {formatNumber(repeatResult.observedMean)}</p>
            </div>
            <div className={styles.resultCard}>
              <h3>{"推定区間"}</h3>
              <p>{"標準偶差"}: {formatNumber(repeatResult.stdDev)}</p>
              <p>{"95%下限"}: {formatNumber(repeatResult.confidenceLow)}</p>
              <p>{"95%上限"}: {formatNumber(repeatResult.confidenceHigh)}</p>
            </div>
            <div className={styles.resultCard}>
              <h3>{"検定"}</h3>
              <p>{"期待値"}: {formatNumber(repeatResult.expectedMean)}</p>
              <p>
                {"t値"}:{" "}
                {repeatResult.tStatistic === null ? "-" : formatNumber(repeatResult.tStatistic)}
              </p>
              <p>
                {"両側p値"}:{" "}
                {repeatResult.pValueApprox === null
                  ? "-"
                  : formatNumber(repeatResult.pValueApprox, 4)}
              </p>
            </div>
          </div>
        )}
      </section>

      <section className={styles.panel}>
        <h2 className={styles.sectionTitle}>{"実行メモ"}</h2>
        <p className={styles.summary}>{pageMessage}</p>
        {error && <div className={styles.errorBox}>{error}</div>}
        {pageError && <div className={styles.errorBox}>{pageError}</div>}
        {!error &&
          !pageError &&
          pageMessage.includes("取得しました") && (
            <div className={styles.successBox}>{pageMessage}</div>
          )}
      </section>
    </div>
  );
};

export default StatisticsAnalysis;










