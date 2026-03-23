import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useActions, type ActionAttribute, type ActionLog } from "../../hooks/useActions";
import {
  useWorldSignals,
  type AnalysisContextResult,
  type ExternalDataSource,
  type FetchWorldSignalInput,
  type WorldSignal,
} from "../../hooks/useWorldSignals";
import { eStatDashboardSignalOptions } from "../../constants/externalDataCatalog";
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

type SummaryStats = { count: number; mean: number; variance: number; stdDev: number };
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

const text = {
  title: "Quantified Self機能",
  subtitle: "Déconstruction d'Objet機能で作った記録に対して、平均差比較・相関・反復推定を行います。",
  back: "ダッシュボードへ戻る",
  dataset: "データ集",
  actionLogs: "行動記録",
  worldSignals: "外部データ",
  source: "データソース",
  signalType: "指標",
  fetchExternal: "外部データを取得",
  compare: "平均差比較を実行",
  correlation: "相関分析を実行",
  repeat: "反復推定を実行",
};

const dayMs = 24 * 60 * 60 * 1000;
const weekMs = 7 * dayMs;
const toDateTimeLocal = (date: Date) => new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
const toRFC3339 = (datetimeLocal: string) => datetimeLocal.length === 16 ? datetimeLocal + ":00Z" : datetimeLocal;
const createRange = (startDaysAgo: number, endDaysAgo: number) => ({ from: toDateTimeLocal(new Date(Date.now() - startDaysAgo * dayMs)), to: toDateTimeLocal(new Date(Date.now() - endDaysAgo * dayMs)) });
const createYearRange = (startYearsAgo: number, endYearsAgo: number) => ({ from: toDateTimeLocal(new Date(Date.UTC(new Date().getUTCFullYear() - startYearsAgo, 0, 1))), to: toDateTimeLocal(new Date(Date.UTC(new Date().getUTCFullYear() - endYearsAgo, 0, 1))) });
const avg = (values: number[]) => values.reduce((sum, value) => sum + value, 0) / values.length;
const variance = (values: number[], mean: number) => values.length < 2 ? 0 : values.reduce((sum, value) => sum + (value - mean) ** 2, 0) / (values.length - 1);
const summarize = (values: number[]): SummaryStats => {
  if (values.length === 0) return { count: 0, mean: 0, variance: 0, stdDev: 0 };
  const mean = avg(values);
  const varianceValue = variance(values, mean);
  return { count: values.length, mean, variance: varianceValue, stdDev: Math.sqrt(varianceValue) };
};
const erf = (x: number) => {
  const sign = x >= 0 ? 1 : -1;
  const absX = Math.abs(x);
  const t = 1 / (1 + 0.3275911 * absX);
  const y = 1 - (((((1.061405429 * t - 1.453152027) * t + 1.421413741) * t - 0.284496736) * t + 0.254829592) * t * Math.exp(-absX * absX));
  return sign * y;
};
const normalCdf = (x: number) => 0.5 * (1 + erf(x / Math.SQRT2));
const format = (value: number, digits = 3) => value.toFixed(digits);
const bucketLabel = (unit: RepeatBucketUnit) => (unit === "week" ? "週" : "日");
const findAttributeValue = (attributes: ActionAttribute[] | undefined, key: string) => attributes?.find((attribute) => attribute.key === key)?.value_number;
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
const computeWelchTest = (groupAValues: number[], groupBValues: number[]): TwoGroupAnalysisResult | null => {
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
  return {
    groupA,
    groupB,
    meanDiff,
    tStatistic,
    degreesOfFreedom: denominatorDf === 0 ? groupA.count + groupB.count - 2 : numerator / denominatorDf,
    pValueApprox: 2 * (1 - normalCdf(Math.abs(tStatistic))),
  };
};

const computeCorrelation = (pairs: Array<{ x: number; y: number }>): CorrelationAnalysisResult | null => {
  if (pairs.length < 3) return null;
  const xValues = pairs.map((pair) => pair.x);
  const yValues = pairs.map((pair) => pair.y);
  const meanX = avg(xValues);
  const meanY = avg(yValues);
  const stdX = Math.sqrt(variance(xValues, meanX));
  const stdY = Math.sqrt(variance(yValues, meanY));
  if (stdX === 0 || stdY === 0) return null;
  const covariance = pairs.reduce((sum, pair) => sum + (pair.x - meanX) * (pair.y - meanY), 0) / (pairs.length - 1);
  const correlation = covariance / (stdX * stdY);
  if (!Number.isFinite(correlation) || Math.abs(correlation) >= 1) return null;
  const tStatistic = correlation * Math.sqrt((pairs.length - 2) / (1 - correlation ** 2));
  return {
    count: pairs.length,
    correlation,
    covariance,
    pValueApprox: 2 * (1 - normalCdf(Math.abs(tStatistic))),
  };
};

const computeRepeatBehavior = (
  logs: ActionLog[],
  from: string,
  to: string,
  keyword: string,
  unit: RepeatBucketUnit,
  expectedMean: number,
): RepeatBehaviorResult | null => {
  const fromTime = new Date(from).getTime();
  const toTime = new Date(to).getTime();
  if (!Number.isFinite(fromTime) || !Number.isFinite(toTime) || fromTime >= toTime) return null;

  const width = unit === "week" ? weekMs : dayMs;
  const bucketCount = Math.max(1, Math.ceil((toTime - fromTime) / width));
  const counts = new Array<number>(bucketCount).fill(0);
  const normalizedKeyword = keyword.trim().toLowerCase();
  const matchedLogs = logs.filter(
    (log) =>
      !normalizedKeyword ||
      log.title.toLowerCase().includes(normalizedKeyword) ||
      log.notes.toLowerCase().includes(normalizedKeyword),
  );

  matchedLogs.forEach((log) => {
    const occurredAt = new Date(log.occurred_at).getTime();
    if (!Number.isFinite(occurredAt) || occurredAt < fromTime || occurredAt >= toTime) return;
    counts[Math.min(bucketCount - 1, Math.floor((occurredAt - fromTime) / width))] += 1;
  });

  const stats = summarize(counts);
  const stderr = stats.count > 1 ? stats.stdDev / Math.sqrt(stats.count) : 0;
  const margin = stderr > 0 ? 1.96 * stderr : 0;
  const tStatistic = stderr > 0 ? (stats.mean - expectedMean) / stderr : null;

  return {
    bucketUnit: unit,
    bucketCount,
    matchedActionCount: matchedLogs.length,
    observedMean: stats.mean,
    expectedMean,
    stdDev: stats.stdDev,
    confidenceLow: stats.mean - margin,
    confidenceHigh: stats.mean + margin,
    tStatistic,
    pValueApprox: tStatistic === null ? null : 2 * (1 - normalCdf(Math.abs(tStatistic))),
  };
};

const StatisticsAnalysis = () => {
  const navigate = useNavigate();
  const { fetchActionsSnapshot, meta: actionMeta } = useActions();
  const { fetchOpenMeteo, fetchAnalysisContext, loading, error } = useWorldSignals();

  const initialA = createRange(14, 7);
  const initialB = createRange(7, 0);
  const initialCorr = createRange(14, 0);
  const initialRepeat = createRange(30, 0);

  const [dataset, setDataset] = useState<DatasetKind>("action_logs");
  const [source, setSource] = useState<ExternalDataSource>("open_meteo");
  const [locationKey, setLocationKey] = useState("tokyo_shinjuku");
  const [signalType, setSignalType] = useState("population_total");
  const [latitude, setLatitude] = useState("35.6895");
  const [longitude, setLongitude] = useState("139.6917");
  const [pastDays, setPastDays] = useState("14");
  const [forecastDays, setForecastDays] = useState("1");
  const [actionAxis, setActionAxis] = useState<ActionAxisKey>("tag_count");
  const [worldAxis, setWorldAxis] = useState<WorldAxisKey>("temperature_c");
  const [corrActionXAxis, setCorrActionXAxis] = useState<ActionAxisKey>("tag_count");
  const [corrActionYAxis, setCorrActionYAxis] = useState<ActionAxisKey>("prototype_id");
  const [corrWorldXAxis, setCorrWorldXAxis] = useState<WorldAxisKey>("temperature_c");
  const [corrWorldYAxis, setCorrWorldYAxis] = useState<WorldAxisKey>("precipitation_mm");
  const [periodAFrom, setPeriodAFrom] = useState(initialA.from);
  const [periodATo, setPeriodATo] = useState(initialA.to);
  const [periodBFrom, setPeriodBFrom] = useState(initialB.from);
  const [periodBTo, setPeriodBTo] = useState(initialB.to);
  const [correlationFrom, setCorrelationFrom] = useState(initialCorr.from);
  const [correlationTo, setCorrelationTo] = useState(initialCorr.to);
  const [repeatFrom, setRepeatFrom] = useState(initialRepeat.from);
  const [repeatTo, setRepeatTo] = useState(initialRepeat.to);
  const [repeatKeyword, setRepeatKeyword] = useState("");
  const [repeatUnit, setRepeatUnit] = useState<RepeatBucketUnit>("day");
  const [expectedRepeatMean, setExpectedRepeatMean] = useState("1");
  const [pageError, setPageError] = useState<string | null>(null);
  const [fetchMessage, setFetchMessage] = useState<string | null>(null);
  const [lastSummary, setLastSummary] = useState("まだ分析を実行していません。");
  const [twoGroupResult, setTwoGroupResult] = useState<TwoGroupAnalysisResult | null>(null);
  const [correlationResult, setCorrelationResult] = useState<CorrelationAnalysisResult | null>(null);
  const [repeatResult, setRepeatResult] = useState<RepeatBehaviorResult | null>(null);
  const actionAxisOptions = useMemo(
    () => Array.from(new Set(["tag_count", "prototype_id", ...actionMeta.axis_candidates.filter((axis) => !["occurred_at", "created_at", "updated_at"].includes(axis))])),
    [actionMeta.axis_candidates],
  );
  const worldAxisOptions = useMemo(
    () => source === "e_stat_dashboard"
      ? [
          { key: "observed_year" as WorldAxisKey, label: "観測年" },
          { key: "signal_value" as WorldAxisKey, label: "指標値" },
        ]
      : [
          { key: "temperature_c" as WorldAxisKey, label: "気温" },
          { key: "precipitation_mm" as WorldAxisKey, label: "降水量" },
          { key: "wind_speed_ms" as WorldAxisKey, label: "風速" },
          { key: "weather_code" as WorldAxisKey, label: "天気コード" },
        ],
    [source],
  );

  useEffect(() => {
    if (source === "e_stat_dashboard") {
      const rangeA = createYearRange(14, 7);
      const rangeB = createYearRange(7, 0);
      const rangeC = createYearRange(14, 0);
      setLocationKey((prev) => (prev === "tokyo_shinjuku" ? "13000" : prev));
      setLatitude("0");
      setLongitude("0");
      setWorldAxis("signal_value");
      setCorrWorldXAxis("observed_year");
      setCorrWorldYAxis("signal_value");
      setPeriodAFrom(rangeA.from);
      setPeriodATo(rangeA.to);
      setPeriodBFrom(rangeB.from);
      setPeriodBTo(rangeB.to);
      setCorrelationFrom(rangeC.from);
      setCorrelationTo(rangeC.to);
      return;
    }

    setLocationKey((prev) => (prev === "13000" ? "tokyo_shinjuku" : prev));
    setLatitude((prev) => (prev === "0" ? "35.6895" : prev));
    setLongitude((prev) => (prev === "0" ? "139.6917" : prev));
    setWorldAxis("temperature_c");
    setCorrWorldXAxis("temperature_c");
    setCorrWorldYAxis("precipitation_mm");
  }, [source]);

  const loadActionValues = async (from: string, to: string, axis: ActionAxisKey) =>
    (
      await fetchActionsSnapshot(1, 1000, [], { from: toRFC3339(from), to: toRFC3339(to), sort: "occurred_at", order: "asc" })
    )?.logs
      .map((log) => actionValue(log, axis))
      .filter((value): value is number => typeof value === "number") ?? [];

  const loadActionLogs = async (from: string, to: string) =>
    (
      await fetchActionsSnapshot(1, 1000, [], { from: toRFC3339(from), to: toRFC3339(to), sort: "occurred_at", order: "asc" })
    )?.logs ?? [];

  const loadWorldContext = async (from: string, to: string): Promise<AnalysisContextResult | null> =>
    fetchAnalysisContext(
      source,
      locationKey.trim(),
      source === "e_stat_dashboard" ? signalType : "",
      new Date(from).toISOString(),
      new Date(to).toISOString(),
      1000,
    );

  const loadWorldValues = async (from: string, to: string, axis: WorldAxisKey) =>
    (await loadWorldContext(from, to))?.world_signals
      .map((signal) => worldValue(signal, axis))
      .filter((value): value is number => typeof value === "number") ?? [];

  const handleFetchExternalData = async () => {
    setPageError(null);
    setFetchMessage(null);

    const payload: FetchWorldSignalInput = {
      source,
      locationKey: locationKey.trim(),
      latitude: source === "open_meteo" ? Number(latitude) : 0,
      longitude: source === "open_meteo" ? Number(longitude) : 0,
      pastDays: source === "open_meteo" ? Number(pastDays) : 0,
      forecastDays: source === "open_meteo" ? Number(forecastDays) : 0,
    };

    if (!payload.locationKey) {
      setPageError(source === "e_stat_dashboard" ? "e-Stat Dashboard では都道府県コードを入力してください。例: 13000" : "location_key を入力してください。");
      return;
    }

    if (source === "open_meteo" && ![payload.latitude, payload.longitude, payload.pastDays, payload.forecastDays].every(Number.isFinite)) {
      setPageError("Open-Meteo では latitude / longitude / past_days / forecast_days を正しく入力してください。");
      return;
    }

    if (!(await fetchOpenMeteo(payload))) {
      setPageError("外部データの取得に失敗しました。設定とサーバログを確認してください。");
      return;
    }

    setFetchMessage("外部データを取得しました。このまま平均差比較や相関分析を実行できます。");
  };

  const handleCompareMeans = async () => {
    setPageError(null);
    setTwoGroupResult(null);
    const valuesA = dataset === "action_logs"
      ? await loadActionValues(periodAFrom, periodATo, actionAxis)
      : await loadWorldValues(periodAFrom, periodATo, worldAxis);
    const valuesB = dataset === "action_logs"
      ? await loadActionValues(periodBFrom, periodBTo, actionAxis)
      : await loadWorldValues(periodBFrom, periodBTo, worldAxis);
    const result = computeWelchTest(valuesA, valuesB);
    if (!result) {
      setPageError(dataset === "world_signals" && source === "e_stat_dashboard"
        ? "e-Stat は年次データなので、各期間に複数年が入るよう期間を広げてください。"
        : "2期間比較に必要な数値データが不足しています。各期間で2件以上の数値が必要です。");
      return;
    }
    setTwoGroupResult(result);
    setLastSummary(`2期間比較を実行しました。平均差は ${format(result.meanDiff)} です。`);
  };

  const handleCorrelation = async () => {
    setPageError(null);
    setCorrelationResult(null);
    if (dataset === "action_logs") {
      const snapshot = await fetchActionsSnapshot(1, 1000, [], { from: correlationFrom, to: correlationTo, sort: "occurred_at", order: "asc" });
      const pairs = snapshot?.logs
        .map((log) => ({ x: actionValue(log, corrActionXAxis), y: actionValue(log, corrActionYAxis) }))
        .filter((pair): pair is { x: number; y: number } => typeof pair.x === "number" && typeof pair.y === "number") ?? [];
      const result = computeCorrelation(pairs);
      if (!result) {
        setPageError("相関分析に必要な数値の組が不足しています。3件以上必要です。");
        return;
      }
      setCorrelationResult(result);
      setLastSummary(`相関分析を実行しました。相関係数は ${format(result.correlation)} です。`);
      return;
    }

    const context = await loadWorldContext(correlationFrom, correlationTo);
    const pairs = context?.world_signals
      .map((signal) => ({ x: worldValue(signal, corrWorldXAxis), y: worldValue(signal, corrWorldYAxis) }))
      .filter((pair): pair is { x: number; y: number } => typeof pair.x === "number" && typeof pair.y === "number") ?? [];
    const result = computeCorrelation(pairs);
    if (!result) {
      setPageError("相関分析に必要な数値の組が不足しています。3件以上必要です。");
      return;
    }
    setCorrelationResult(result);
    setLastSummary(`相関分析を実行しました。相関係数は ${format(result.correlation)} です。`);
  };
  const handleRepeatEstimation = async () => {
    setPageError(null);
    setRepeatResult(null);
    const expected = Number(expectedRepeatMean);
    if (!Number.isFinite(expected) || expected < 0) {
      setPageError("期待反復回数は 0 以上の数値で入力してください。");
      return;
    }
    const logs = await loadActionLogs(repeatFrom, repeatTo);
    const result = computeRepeatBehavior(logs, repeatFrom, repeatTo, repeatKeyword, repeatUnit, expected);
    if (!result || result.bucketCount < 2) {
      setPageError("反復推定に必要な期間データが不足しています。期間を広げてください。");
      return;
    }
    setRepeatResult(result);
    setLastSummary(`反復推定を実行しました。${bucketLabel(result.bucketUnit)}あたりの平均反復回数は ${format(result.observedMean)} です。`);
  };

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>{text.title}</h1>
          <p className={styles.subtitle}>{text.subtitle}</p>
        </div>
        <button className={styles.secondaryButton} onClick={() => navigate("/dashboard")}>{text.back}</button>
      </div>

      <section className={styles.panel}>
        <h2 className={styles.sectionTitle}>分析対象</h2>
        <div className={styles.formGrid}>
          <label className={styles.field}><span>{text.dataset}</span><select value={dataset} onChange={(event) => setDataset(event.target.value as DatasetKind)}><option value="action_logs">{text.actionLogs}</option><option value="world_signals">{text.worldSignals}</option></select></label>
          {dataset === "world_signals" && <>
            <label className={styles.field}><span>{text.source}</span><select value={source} onChange={(event) => setSource(event.target.value as ExternalDataSource)}><option value="open_meteo">Open-Meteo</option><option value="e_stat_dashboard">e-Stat Dashboard</option></select></label>
            <label className={styles.field}><span>location_key</span><input value={locationKey} onChange={(event) => setLocationKey(event.target.value)} /></label>
            {source === "e_stat_dashboard" ? (
              <label className={styles.field}><span>{text.signalType}</span><select value={signalType} onChange={(event) => setSignalType(event.target.value)}>{eStatDashboardSignalOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}</select></label>
            ) : (
              <>
                <label className={styles.field}><span>latitude</span><input value={latitude} onChange={(event) => setLatitude(event.target.value)} /></label>
                <label className={styles.field}><span>longitude</span><input value={longitude} onChange={(event) => setLongitude(event.target.value)} /></label>
                <label className={styles.field}><span>past_days</span><input value={pastDays} onChange={(event) => setPastDays(event.target.value)} /></label>
                <label className={styles.field}><span>forecast_days</span><input value={forecastDays} onChange={(event) => setForecastDays(event.target.value)} /></label>
              </>
            )}
          </>}
        </div>
        {dataset === "world_signals" ? (
          <>
            <div className={styles.actions}><button className={styles.primaryButton} onClick={handleFetchExternalData} disabled={loading}>{text.fetchExternal}</button></div>
            <div className={styles.notice}>{source === "e_stat_dashboard" ? "e-Stat Dashboard を使う場合は、都道府県コードを入れて先に取得してください。例: 13000" : "Open-Meteo を使う場合は、location_key・緯度経度・過去日数を設定して取得してください。"}</div>
          </>
        ) : <div className={styles.notice}>テストデータは developer / tester に投入されます。現在のログインユーザーの記録だけが集計対象です。</div>}
      </section>

      <div className={styles.sectionGrid}>
        <section className={styles.panel}>
          <h2 className={styles.sectionTitle}>2期間の平均差比較</h2>
          <div className={styles.formGrid}>
            <label className={styles.field}><span>期間A 開始</span><input type="datetime-local" value={periodAFrom} onChange={(event) => setPeriodAFrom(event.target.value)} /></label>
            <label className={styles.field}><span>期間A 終了</span><input type="datetime-local" value={periodATo} onChange={(event) => setPeriodATo(event.target.value)} /></label>
            <label className={styles.field}><span>期間B 開始</span><input type="datetime-local" value={periodBFrom} onChange={(event) => setPeriodBFrom(event.target.value)} /></label>
            <label className={styles.field}><span>期間B 終了</span><input type="datetime-local" value={periodBTo} onChange={(event) => setPeriodBTo(event.target.value)} /></label>
            {dataset === "action_logs" ? <label className={styles.field}><span>比較軸</span><select value={actionAxis} onChange={(event) => setActionAxis(event.target.value)}>{actionAxisOptions.map((axis) => <option key={axis} value={axis}>{axis}</option>)}</select></label> : <label className={styles.field}><span>比較軸</span><select value={worldAxis} onChange={(event) => setWorldAxis(event.target.value as WorldAxisKey)}>{worldAxisOptions.map((axis) => <option key={axis.key} value={axis.key}>{axis.label}</option>)}</select></label>}
          </div>
          <div className={styles.actions}><button className={styles.primaryButton} onClick={handleCompareMeans} disabled={loading}>{text.compare}</button></div>
          {twoGroupResult && <div className={styles.resultGrid}><div className={styles.resultCard}><h3>期間A</h3><p>件数: {twoGroupResult.groupA.count}</p><p>平均: {format(twoGroupResult.groupA.mean)}</p><p>分散: {format(twoGroupResult.groupA.variance)}</p></div><div className={styles.resultCard}><h3>期間B</h3><p>件数: {twoGroupResult.groupB.count}</p><p>平均: {format(twoGroupResult.groupB.mean)}</p><p>分散: {format(twoGroupResult.groupB.variance)}</p></div><div className={styles.resultCard}><h3>比較結果</h3><p>平均差: {format(twoGroupResult.meanDiff)}</p><p>t値: {format(twoGroupResult.tStatistic)}</p><p>自由度: {format(twoGroupResult.degreesOfFreedom, 2)}</p><p>近似p値: {format(twoGroupResult.pValueApprox, 4)}</p></div></div>}
        </section>

        <section className={styles.panel}>
          <h2 className={styles.sectionTitle}>相関分析</h2>
          <div className={styles.formGrid}>
            <label className={styles.field}><span>開始</span><input type="datetime-local" value={correlationFrom} onChange={(event) => setCorrelationFrom(event.target.value)} /></label>
            <label className={styles.field}><span>終了</span><input type="datetime-local" value={correlationTo} onChange={(event) => setCorrelationTo(event.target.value)} /></label>
            {dataset === "action_logs" ? <><label className={styles.field}><span>X軸</span><select value={corrActionXAxis} onChange={(event) => setCorrActionXAxis(event.target.value)}>{actionAxisOptions.map((axis) => <option key={axis} value={axis}>{axis}</option>)}</select></label><label className={styles.field}><span>Y軸</span><select value={corrActionYAxis} onChange={(event) => setCorrActionYAxis(event.target.value)}>{actionAxisOptions.map((axis) => <option key={axis} value={axis}>{axis}</option>)}</select></label></> : <><label className={styles.field}><span>X軸</span><select value={corrWorldXAxis} onChange={(event) => setCorrWorldXAxis(event.target.value as WorldAxisKey)}>{worldAxisOptions.map((axis) => <option key={axis.key} value={axis.key}>{axis.label}</option>)}</select></label><label className={styles.field}><span>Y軸</span><select value={corrWorldYAxis} onChange={(event) => setCorrWorldYAxis(event.target.value as WorldAxisKey)}>{worldAxisOptions.map((axis) => <option key={axis.key} value={axis.key}>{axis.label}</option>)}</select></label></>}
          </div>
          <div className={styles.actions}><button className={styles.primaryButton} onClick={handleCorrelation} disabled={loading}>{text.correlation}</button></div>
          {correlationResult && <div className={styles.resultGrid}><div className={styles.resultCard}><h3>相関結果</h3><p>件数: {correlationResult.count}</p><p>相関係数: {format(correlationResult.correlation)}</p><p>共分散: {format(correlationResult.covariance)}</p><p>近似p値: {format(correlationResult.pValueApprox, 4)}</p></div></div>}
        </section>
      </div>
      <section className={styles.panel}>
        <h2 className={styles.sectionTitle}>反復行動の推定と検定</h2>
        <p className={styles.description}>同じ行動が繰り返されると仮定した場合の、単位期間あたりの平均反復回数を推定します。</p>
        {dataset !== "action_logs" ? <div className={styles.notice}>反復行動の推定は行動記録専用です。データ集を行動記録に切り替えてください。</div> : <>
          <div className={styles.formGrid}>
            <label className={styles.field}><span>対象キーワード</span><input value={repeatKeyword} onChange={(event) => setRepeatKeyword(event.target.value)} placeholder="例: 山手線 / 弁当 / 混雑" /></label>
            <label className={styles.field}><span>開始</span><input type="datetime-local" value={repeatFrom} onChange={(event) => setRepeatFrom(event.target.value)} /></label>
            <label className={styles.field}><span>終了</span><input type="datetime-local" value={repeatTo} onChange={(event) => setRepeatTo(event.target.value)} /></label>
            <label className={styles.field}><span>集計単位</span><select value={repeatUnit} onChange={(event) => setRepeatUnit(event.target.value as RepeatBucketUnit)}><option value="day">日</option><option value="week">週</option></select></label>
            <label className={styles.field}><span>期待反復回数 / {bucketLabel(repeatUnit)}</span><input value={expectedRepeatMean} onChange={(event) => setExpectedRepeatMean(event.target.value)} /></label>
          </div>
          <div className={styles.actions}><button className={styles.primaryButton} onClick={handleRepeatEstimation} disabled={loading}>{text.repeat}</button></div>
          {repeatResult && <div className={styles.resultGrid}><div className={styles.resultCard}><h3>集計結果</h3><p>期間バケット数: {repeatResult.bucketCount}</p><p>一致した記録数: {repeatResult.matchedActionCount}</p><p>平均反復回数 / {bucketLabel(repeatResult.bucketUnit)}: {format(repeatResult.observedMean)}</p><p>標準偏差: {format(repeatResult.stdDev)}</p></div><div className={styles.resultCard}><h3>推定区間</h3><p>95%信頼区間下限: {format(repeatResult.confidenceLow)}</p><p>95%信頼区間上限: {format(repeatResult.confidenceHigh)}</p><p>期待値: {format(repeatResult.expectedMean)}</p></div><div className={styles.resultCard}><h3>検定結果</h3><p>t値: {repeatResult.tStatistic === null ? "-" : format(repeatResult.tStatistic)}</p><p>近似p値: {repeatResult.pValueApprox === null ? "-" : format(repeatResult.pValueApprox, 4)}</p><p>{repeatResult.pValueApprox !== null && repeatResult.pValueApprox < 0.05 ? "期待値との差が比較的大きいとみなせます。" : "期待値との差は大きいとは言い切れません。"}</p></div></div>}
        </>}
      </section>

      <section className={styles.panel}>
        <h2 className={styles.sectionTitle}>実行メモ</h2>
        <p className={styles.summary}>{lastSummary}</p>
        {fetchMessage && <div className={styles.successBox}>{fetchMessage}</div>}
        {(pageError || error) && <div className={styles.errorBox}>{pageError || error}</div>}
      </section>
    </div>
  );
};

export default StatisticsAnalysis;
