import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useActions } from "../../hooks/useActions";
import { usePrototypes } from "../../hooks/usePrototypes";
import { useTags } from "../../hooks/useTags";
import { useTargets } from "../../hooks/useTargets";
import {
  useAnalysisSnapshots,
  type AnalysisSeverity,
  type AnalysisSnapshot,
} from "../../hooks/useAnalysisSnapshots";
import {
  useDistributionAnalysis,
  type DistributionAnalysisResult,
  type DistributionDataset,
} from "../../hooks/useDistributionAnalysis";
import {
  useWorldSignals,
  type AnalysisContextResult,
  type ExternalDataSource,
} from "../../hooks/useWorldSignals";
import { eStatDashboardSignalOptions } from "../../constants/externalDataCatalog";
import styles from "../../assets/styles/Actions.module.scss";
import ActionsSearchSection from "./ActionsSearchSection";
import ActionsOpenDataSection from "./ActionsOpenDataSection";
import ActionsDistributionSection from "./ActionsDistributionSection";
import ActionsHistorySection from "./ActionsHistorySection";
import ActionsResultsSection from "./ActionsResultsSection";
import {
  type ParsedQuery,
  type SortKey,
  type SortOrder,
  buildActionAxisCandidates,
  buildActionScatterPoints,
  buildGroupedTags,
  buildPrototypeNameMap,
  buildTagNameMap,
  buildTagNameToID,
  formatTagNames,
  isSameDateTimeByMinute,
  isValidDateRange,
  parseDanbooruStyleQuery,
  syncDateRangeWithOpenDataWindow,
  toLocalDateTimeValue,
} from "./components/actionsSearchSectionHelpers";
import {
  type ContextVizMode,
  type ScatterAxisKey,
  buildAvailableWorldAxisEntries,
  buildContextScatterPoints,
  buildContextStaleState,
  buildContextTimePoints,
} from "./components/actionsOpenDataSectionHelpers";
import {
  ACTION_AXIS_DEFAULTS,
  actionAxisValue,
  calculateDistributionStats,
  defaultAxisByDataset,
  deriveSeverity,
  toLegacyDistributionStats,
  toSnapshotScore,
  toSnapshotSeverity,
  type AnalysisViewState,
  type SelectedDistributionBin,
} from "./components/actionsDistributionSectionHelpers";
import { type HistorySortKey } from "./components/actionsHistorySectionHelpers";

const Actions = () => {
  const navigate = useNavigate();
  const { actions, pagination, meta, loading, error, fetchActions, fetchActionsSnapshot, deleteAction } = useActions();
  const { tags, fetchTags } = useTags();
  const { prototypes, fetchPrototypes } = usePrototypes();
  const { createTarget, loading: targetSaving } = useTargets();
  const { listSnapshots, createSnapshot, clearSnapshots, loading: snapshotLoading, error: snapshotError } =
    useAnalysisSnapshots();
  const {
    loading: worldSignalLoading,
    error: worldSignalError,
    fetchOpenMeteo,
    fetchAnalysisContext,
  } = useWorldSignals();
  const {
    loading: distributionLoading,
    error: distributionError,
    analyzeDistribution,
  } = useDistributionAnalysis();

  const [sort, setSort] = useState<SortKey>("occurred_at");
  const [order, setOrder] = useState<SortOrder>("desc");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [isDateRangeManual, setIsDateRangeManual] = useState(false);
  const [filterTagIDs, setFilterTagIDs] = useState<number[]>([]);
  const [danbooruQuery, setDanbooruQuery] = useState("");
  const [queryError, setQueryError] = useState<string | null>(null);
  const [saveMessage, setSaveMessage] = useState<string | null>(null);
  const [analysisState, setAnalysisState] = useState<AnalysisViewState | null>(null);
  const [analysisHistory, setAnalysisHistory] = useState<AnalysisSnapshot[]>([]);
  const [analysisMessage, setAnalysisMessage] = useState<string | null>(null);
  const [distributionDataset, setDistributionDataset] = useState<DistributionDataset>("action_logs");
  const [distributionAxis, setDistributionAxis] = useState<string>("tag_count");
  const [distributionResult, setDistributionResult] = useState<DistributionAnalysisResult | null>(null);
  const [selectedDistributionBin, setSelectedDistributionBin] = useState<SelectedDistributionBin | null>(null);
  const [locationKey, setLocationKey] = useState("tokyo_shinjuku");
  const [externalDataSource, setExternalDataSource] = useState<ExternalDataSource>("open_meteo");
  const [externalSignalType, setExternalSignalType] = useState("population_total");
  const [latitude, setLatitude] = useState("35.6895");
  const [longitude, setLongitude] = useState("139.6917");
  const [pastDays, setPastDays] = useState("7");
  const [forecastDays, setForecastDays] = useState("1");
  const [contextResult, setContextResult] = useState<AnalysisContextResult | null>(null);
  const [contextVizMode, setContextVizMode] = useState<ContextVizMode>("time");
  const [scatterXAxis, setScatterXAxis] = useState<ScatterAxisKey>("temperature_c");
  const [scatterYAxis, setScatterYAxis] = useState<ScatterAxisKey>("precipitation_mm");
  const [actionScatterXAxis, setActionScatterXAxis] = useState<string>("occurred_at");
  const [actionScatterYAxis, setActionScatterYAxis] = useState<string>("tag_count");
  const [historySeverityFilter, setHistorySeverityFilter] = useState<AnalysisSeverity | "ALL">("ALL");
  const [historySort, setHistorySort] = useState<HistorySortKey>("created_desc");
  const [openDataPanelOpen, setOpenDataPanelOpen] = useState(false);

  useEffect(() => {
    void fetchActions();
    void fetchTags();
    void fetchPrototypes();
    void loadAnalysisHistory();
    const now = new Date();
    const past = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
    setFrom(toLocalDateTimeValue(past));
    setTo(toLocalDateTimeValue(now));
  }, []);

  useEffect(() => {
    const synced = syncDateRangeWithOpenDataWindow(pastDays, forecastDays);
    if (!synced || isDateRangeManual) return;
    setFrom(synced.from);
    setTo(synced.to);
  }, [pastDays, forecastDays, isDateRangeManual]);

  useEffect(() => {
    if (externalDataSource === "e_stat_dashboard") {
      setScatterXAxis("observed_year");
      setScatterYAxis("signal_value");
      if (distributionDataset === "world_signals") {
        setDistributionAxis("signal_value");
      }
      if (locationKey === "tokyo_shinjuku") {
        setLocationKey("13000");
      }
      return;
    }

    setScatterXAxis("temperature_c");
    setScatterYAxis("weather_code");
    if (distributionDataset === "world_signals" && distributionAxis === "signal_value") {
      setDistributionAxis("temperature_c");
    }
    if (locationKey === "13000") {
      setLocationKey("tokyo_shinjuku");
    }
  }, [externalDataSource, distributionDataset, distributionAxis, locationKey]);

  const effectiveExternalSignalType = externalDataSource === "e_stat_dashboard" ? externalSignalType : "";

  const loadAnalysisHistory = async () => {
    const snapshots = await listSnapshots(100);
    setAnalysisHistory(snapshots);
  };

  const tagNameMap = useMemo(() => buildTagNameMap(tags), [tags]);
  const tagNameToID = useMemo(() => buildTagNameToID(tags), [tags]);
  const prototypeNameMap = useMemo(() => buildPrototypeNameMap(prototypes), [prototypes]);
  const groupedTags = useMemo(() => buildGroupedTags(tags), [tags]);
  const actionAxisCandidates = useMemo(() => buildActionAxisCandidates(meta, [...ACTION_AXIS_DEFAULTS]), [meta]);

  useEffect(() => {
    if (actionAxisCandidates.length === 0) return;
    const defaultActionAxis = actionAxisCandidates.includes("tag_count") ? "tag_count" : actionAxisCandidates[0];
    if (!actionAxisCandidates.includes(actionScatterXAxis)) {
      setActionScatterXAxis(defaultActionAxis);
    }
    if (!actionAxisCandidates.includes(actionScatterYAxis)) {
      setActionScatterYAxis(actionAxisCandidates[1] || actionAxisCandidates[0]);
    }
    if (distributionDataset === "action_logs" && !actionAxisCandidates.includes(distributionAxis)) {
      setDistributionAxis(defaultActionAxis);
    }
  }, [actionAxisCandidates, actionScatterXAxis, actionScatterYAxis, distributionAxis, distributionDataset]);

  const toggleFilterTag = (tagID: number) => {
    setFilterTagIDs((prev) => (prev.includes(tagID) ? prev.filter((id) => id !== tagID) : [...prev, tagID]));
  };

  const resolveBaselineWindow = (): { from?: string; to?: string } => {
    if (!isValidDateRange(from, to)) return {};
    const fromDate = new Date(from);
    const toDate = new Date(to);
    const duration = toDate.getTime() - fromDate.getTime();
    const baselineTo = fromDate;
    const baselineFrom = new Date(fromDate.getTime() - duration);
    return {
      from: baselineFrom.toISOString(),
      to: baselineTo.toISOString(),
    };
  };

  const analyzeCurrentResult = async (
    mergedAnd: number[],
    parsed: ParsedQuery,
    currentFrom?: string,
    currentTo?: string
  ) => {
    const currentSnapshot = await fetchActionsSnapshot(1, 500, mergedAnd, {
      sort,
      order,
      from: currentFrom,
      to: currentTo,
      anyTagIDs: parsed.anyTagIDs,
      anyTagGroups: parsed.anyTagGroups,
      excludeTagIDs: parsed.excludeTagIDs,
    });
    if (!currentSnapshot) return;

    const baselineWindow = resolveBaselineWindow();
    const baselineSnapshot = await fetchActionsSnapshot(1, 500, mergedAnd, {
      sort: "occurred_at",
      order: "desc",
      from: baselineWindow.from,
      to: baselineWindow.to,
      anyTagIDs: parsed.anyTagIDs,
      anyTagGroups: parsed.anyTagGroups,
      excludeTagIDs: parsed.excludeTagIDs,
    });

    const currentStats = calculateDistributionStats(currentSnapshot.logs);
    const baselineStats = baselineSnapshot ? calculateDistributionStats(baselineSnapshot.logs) : undefined;
    const summary = deriveSeverity(currentStats, baselineStats);

    setAnalysisState({
      ...summary,
      current: currentStats,
      baseline: baselineStats,
    });
  };

  const saveDistributionSnapshot = async (result: DistributionAnalysisResult) => {
    const saved = await createSnapshot({
      query_text: `${danbooruQuery.trim()} axis:${result.axis} dataset:${result.dataset}`.trim(),
      severity: toSnapshotSeverity(result),
      score: Math.round(toSnapshotScore(result)),
      delta_avg_tag: result.comparison?.mean_diff ?? 0,
      delta_var_tag: result.comparison?.variance_diff ?? 0,
      delta_prototype_rate: 0,
      current: toLegacyDistributionStats(result.current),
      baseline: result.baseline ? toLegacyDistributionStats(result.baseline) : undefined,
    });

    if (saved) {
      setAnalysisHistory((prev) => [saved, ...prev].slice(0, 100));
      setAnalysisMessage("分布分析の結果を履歴に保存しました。");
    }
  };

  const runDistributionAnalysis = async (
    mergedAnd: number[],
    parsed: ParsedQuery,
    currentFrom?: string,
    currentTo?: string,
    datasetOverride?: DistributionDataset,
    axisOverride?: string
  ) => {
    const dataset = datasetOverride ?? distributionDataset;
    const axis = axisOverride ?? distributionAxis;
    const result = await analyzeDistribution({
      dataset,
      axis,
      source: dataset === "world_signals" ? externalDataSource : undefined,
      signalType: dataset === "world_signals" ? effectiveExternalSignalType : undefined,
      tagIDs: dataset === "action_logs" ? mergedAnd : undefined,
      anyTagIDs: dataset === "action_logs" ? parsed.anyTagIDs : undefined,
      anyTagGroups: dataset === "action_logs" ? parsed.anyTagGroups : undefined,
      excludeTagIDs: dataset === "action_logs" ? parsed.excludeTagIDs : undefined,
      from: currentFrom,
      to: currentTo,
      locationKey: dataset === "world_signals" ? locationKey.trim() : undefined,
    });

    if (result) {
      setDistributionDataset(dataset);
      setDistributionAxis(axis);
      setDistributionResult(result);
      setSelectedDistributionBin(null);
      await saveDistributionSnapshot(result);
    }
  };

  const runDistributionAnalysisFromCurrent = async () => {
    const parsed = parseDanbooruStyleQuery(danbooruQuery, tagNameToID);
    if (parsed.unknownTokens.length > 0) {
      setQueryError(`未知のタグがあります: ${parsed.unknownTokens.join(", ")}`);
      return;
    }
    setQueryError(null);
    const mergedAnd = Array.from(new Set([...filterTagIDs, ...parsed.andTagIDs]));
    const fromISO = from ? new Date(from).toISOString() : undefined;
    const toISO = to ? new Date(to).toISOString() : undefined;
    await runDistributionAnalysis(mergedAnd, parsed, fromISO, toISO);
  };

  const runDistributionFromCurrentInputs = async (datasetOverride?: DistributionDataset, axisOverride?: string) => {
    const parsed = parseDanbooruStyleQuery(danbooruQuery, tagNameToID);
    if (parsed.unknownTokens.length > 0) {
      setQueryError(`未知のタグがあります: ${parsed.unknownTokens.join(", ")}`);
      return;
    }
    setQueryError(null);
    const mergedAnd = Array.from(new Set([...filterTagIDs, ...parsed.andTagIDs]));
    const fromISO = from ? new Date(from).toISOString() : undefined;
    const toISO = to ? new Date(to).toISOString() : undefined;
    await runDistributionAnalysis(mergedAnd, parsed, fromISO, toISO, datasetOverride, axisOverride);
  };

  const handleSearch = async (queryOverride?: string) => {
    const activeQuery = queryOverride ?? danbooruQuery;
    const parsed = parseDanbooruStyleQuery(activeQuery, tagNameToID);
    if (parsed.unknownTokens.length > 0) {
      setQueryError(`未知のタグがあります: ${parsed.unknownTokens.join(", ")}`);
      return;
    }

    setQueryError(null);
    setSaveMessage(null);
    setAnalysisMessage(null);
    setSelectedDistributionBin(null);

    const mergedAnd = Array.from(new Set([...filterTagIDs, ...parsed.andTagIDs]));
    const fromISO = from ? new Date(from).toISOString() : undefined;
    const toISO = to ? new Date(to).toISOString() : undefined;

    await fetchActions(1, 20, mergedAnd, {
      sort,
      order,
      from: fromISO,
      to: toISO,
      anyTagIDs: parsed.anyTagIDs,
      anyTagGroups: parsed.anyTagGroups,
      excludeTagIDs: parsed.excludeTagIDs,
    });
    await analyzeCurrentResult(mergedAnd, parsed, fromISO, toISO);
    await runDistributionAnalysis(mergedAnd, parsed, fromISO, toISO);
  };

  const appendSuggestedTagToQuery = async (tag: string, mode: "append" | "search" = "append") => {
    const tokens = danbooruQuery.trim() ? danbooruQuery.trim().split(/\s+/) : [];
    if (tokens.includes(tag)) {
      setAnalysisMessage(`タグ ${tag} はすでにクエリへ含まれています。`);
      if (mode === "search") {
        await handleSearch();
      }
      return;
    }

    const updated = [...tokens, tag].join(" ").trim();
    setDanbooruQuery(updated);

    if (mode === "search") {
      setAnalysisMessage(`タグ ${tag} を追加して再検索します。`);
      await handleSearch(updated);
      return;
    }

    setAnalysisMessage(`タグ ${tag} をクエリに追加しました。`);
  };

  const selectDistributionBin = (index: number) => {
    if (!distributionResult?.residual_bins[index]) return;
    setSelectedDistributionBin({ index, bin: distributionResult.residual_bins[index] });
    setAnalysisMessage("補完グラフの帯域を選択しました。候補タグや候補軸の試行に使えます。");
  };

  const handleClear = async () => {
    setFilterTagIDs([]);
    setDanbooruQuery("");
    setQueryError(null);
    setSaveMessage(null);
    setAnalysisMessage(null);
    setAnalysisState(null);
    setDistributionResult(null);
    setSelectedDistributionBin(null);
    setSort("occurred_at");
    setOrder("desc");
    setIsDateRangeManual(false);
    setFrom("");
    setTo("");
    await fetchActions(1, 20, [], { sort: "occurred_at", order: "desc" });
  };

  const handleSaveAsTarget = async () => {
    const parsed = parseDanbooruStyleQuery(danbooruQuery, tagNameToID);
    if (parsed.unknownTokens.length > 0) {
      setQueryError(`未知のタグがあります: ${parsed.unknownTokens.join(", ")}`);
      return;
    }
    setQueryError(null);

    const mergedAnd = Array.from(new Set([...filterTagIDs, ...parsed.andTagIDs]));
    if (mergedAnd.length === 0 && parsed.anyTagIDs.length === 0 && parsed.anyTagGroups.length === 0) {
      setSaveMessage("保存するには、少なくとも1つ以上の条件が必要です。");
      return;
    }

    const suggestedName = `saved-search-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, "-")}`;
    const name = window.prompt("保存名を入力してください", suggestedName)?.trim();
    if (!name) return;

    const normalizedLocationKey = locationKey.trim();
    const locationToken = normalizedLocationKey ? `location_key:${normalizedLocationKey}` : "";
    const queryTextToSave = [danbooruQuery.trim(), locationToken].filter(Boolean).join(" ");
    const descriptionToSave = queryTextToSave || "保存済み検索条件";

    const created = await createTarget(name, descriptionToSave, mergedAnd, {
      queryText: queryTextToSave,
      anyTagIDs: parsed.anyTagIDs,
      anyTagGroups: parsed.anyTagGroups,
      excludeTagIDs: parsed.excludeTagIDs,
    });

    if (!created) {
      setSaveMessage("検索条件の保存に失敗しました。");
      return;
    }
    setSaveMessage(`検索条件 ${created.name} (#${created.id}) を保存しました。`);
  };

  const handleClearSnapshots = async () => {
    if (!window.confirm("分析履歴をすべて削除しますか。")) return;
    const ok = await clearSnapshots();
    if (ok) {
      setAnalysisHistory([]);
      setAnalysisMessage("分析履歴を削除しました。");
    }
  };

  const validateWorldSignalInputs = (): { lat: number; lon: number; pDays: number; fDays: number } | null => {
    if (externalDataSource === "e_stat_dashboard") {
      if (!locationKey.trim()) {
        setAnalysisMessage("e-Stat Dashboard では location_key に都道府県コードが必要です。例: 13000");
        return null;
      }
      return {
        lat: Number(latitude) || 0,
        lon: Number(longitude) || 0,
        pDays: Number(pastDays) || 0,
        fDays: Number(forecastDays) || 0,
      };
    }

    if (!locationKey.trim()) {
      setAnalysisMessage("location_key を入力してください。");
      return null;
    }

    const lat = Number(latitude);
    const lon = Number(longitude);
    const pDays = Number(pastDays);
    const fDays = Number(forecastDays);

    if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
      setAnalysisMessage("latitude と longitude は数値で入力してください。");
      return null;
    }
    if (!Number.isFinite(pDays) || !Number.isFinite(fDays)) {
      setAnalysisMessage("past_days と forecast_days は数値で入力してください。");
      return null;
    }

    return { lat, lon, pDays, fDays };
  };

  const handleLoadAnalysisContext = async () => {
    if (!from || !to) {
      setAnalysisMessage("分析コンテキストを取得するには from / to が必要です。");
      return;
    }
    const result = await fetchAnalysisContext(
      externalDataSource,
      locationKey.trim(),
      effectiveExternalSignalType,
      new Date(from).toISOString(),
      new Date(to).toISOString(),
      200
    );
    if (result) {
      setContextResult(result);
      if (result.summary.signal_count > 0 && result.summary.action_count === 0) {
        setDistributionDataset("world_signals");
      }
      setAnalysisMessage("分析コンテキストを更新しました。");
    }
  };

  const handleFetchWorldSignals = async () => {
    const validated = validateWorldSignalInputs();
    if (!validated) return;

    const ok = await fetchOpenMeteo({
      source: externalDataSource,
      locationKey: locationKey.trim(),
      latitude: validated.lat,
      longitude: validated.lon,
      pastDays: validated.pDays,
      forecastDays: validated.fDays,
    });
    if (!ok) return;

    await handleLoadAnalysisContext();
    setDistributionDataset("world_signals");
    setDistributionAxis(externalDataSource === "e_stat_dashboard" ? "signal_value" : "temperature_c");

    const parsed = parseDanbooruStyleQuery(danbooruQuery, tagNameToID);
    const mergedAnd = Array.from(new Set([...filterTagIDs, ...parsed.andTagIDs]));
    const fromISO = from ? new Date(from).toISOString() : undefined;
    const toISO = to ? new Date(to).toISOString() : undefined;
    await runDistributionAnalysis(
      mergedAnd,
      parsed,
      fromISO,
      toISO,
      "world_signals",
      externalDataSource === "e_stat_dashboard" ? "signal_value" : "temperature_c"
    );
    setAnalysisMessage("外部データ取得、表示更新、分布分析の自動実行まで完了しました。");
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm("この行動記録を削除しますか。")) return;
    await deleteAction(id);
  };

  const contextTimePoints = useMemo(
    () => buildContextTimePoints(contextResult, externalDataSource),
    [contextResult, externalDataSource]
  );
  const contextScatterPoints = useMemo(
    () => buildContextScatterPoints(contextResult, scatterXAxis, scatterYAxis),
    [contextResult, scatterXAxis, scatterYAxis]
  );
  const availableWorldAxisEntries = useMemo(
    () => buildAvailableWorldAxisEntries(externalDataSource),
    [externalDataSource]
  );

  const filteredSortedHistory = useMemo(() => {
    const filtered =
      historySeverityFilter === "ALL"
        ? analysisHistory
        : analysisHistory.filter((item) => item.severity === historySeverityFilter);

    const copied = [...filtered];
    if (historySort === "created_desc") {
      copied.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    } else if (historySort === "created_asc") {
      copied.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
    } else {
      copied.sort((a, b) => b.score - a.score);
    }
    return copied;
  }, [analysisHistory, historySeverityFilter, historySort]);

  const actionScatterPoints = useMemo(
    () => buildActionScatterPoints(actions, actionScatterXAxis, actionScatterYAxis, actionAxisValue),
    [actions, actionScatterXAxis, actionScatterYAxis]
  );

  const distributionDatasetLabel = distributionDataset === "action_logs" ? "行動記録" : "外部ビッグデータ";
  const selectedDatasetCount =
    distributionDataset === "action_logs"
      ? contextResult?.summary.action_count ?? actions.length
      : contextResult?.summary.signal_count ?? 0;

  const distributionEmptyHint =
    distributionDataset === "action_logs" && selectedDatasetCount === 0
      ? "行動記録が0件です。検索条件を見直すか、先にデータを追加してください。"
      : distributionDataset === "world_signals" && selectedDatasetCount === 0
        ? "外部ビッグデータが0件です。Open-Meteo取得か分析期間を見直してください。"
        : null;

  const selectedBinStatus = selectedDistributionBin
    ? selectedDistributionBin.bin.gap_count > 0
      ? {
          kind: "shortage" as const,
          label: "不足帯域",
          message: "この帯域には、期待より少ない値しか入っていません。条件追加や群分割の候補を試してください。",
        }
      : selectedDistributionBin.bin.gap_count < 0
        ? {
            kind: "excess" as const,
            label: "過剰帯域",
            message: "この帯域には、期待より値が集まりすぎています。偏りを生む条件が残っている可能性があります。",
          }
        : {
            kind: "balanced" as const,
            label: "均衡帯域",
            message: "この帯域は期待分布に近い状態です。ほかの帯域との差分確認に使ってください。",
          }
    : null;

  const switchDistributionDataset = (dataset: DistributionDataset) => {
    setDistributionDataset(dataset);
    if (dataset === "world_signals") {
      setDistributionAxis(externalDataSource === "e_stat_dashboard" ? "signal_value" : "temperature_c");
    } else {
      setDistributionAxis(defaultAxisByDataset(dataset, actionAxisCandidates));
    }
    setDistributionResult(null);
    setSelectedDistributionBin(null);
  };

  const resetDateRangeSync = () => {
    const synced = syncDateRangeWithOpenDataWindow(pastDays, forecastDays);
    if (!synced) return;
    setIsDateRangeManual(false);
    setFrom(synced.from);
    setTo(synced.to);
    setAnalysisMessage("表示期間を外部データ設定に再同期しました。");
  };

  const contextIsStale = useMemo(
    () =>
      buildContextStaleState(
        contextResult,
        {
          source: externalDataSource,
          signalType: effectiveExternalSignalType,
          locationKey: locationKey.trim(),
          from,
          to,
        },
        isSameDateTimeByMinute
      ),
    [contextResult, externalDataSource, effectiveExternalSignalType, locationKey, from, to]
  );

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>Déconstruction d'Objet機能</h1>
        <div className={styles.actionsRow}>
          <button
            className={styles.secondaryButton}
            onClick={() =>
              navigate("/statistics", {
                state: {
                  dataset: "action_logs",
                  danbooruQuery,
                  from,
                  to,
                  filterTagIDs,
                },
              })
            }
          >
            この検索条件で統計分析
          </button>
          <button className={styles.primaryButton} onClick={() => navigate("/records/actions-form")}>
            新規行動記録
          </button>
        </div>
      </div>

      {error && <div className={styles.errorBox}>エラー: {error}</div>}
      {queryError && <div className={styles.errorBox}>検索条件エラー: {queryError}</div>}
      {snapshotError && <div className={styles.errorBox}>分析履歴エラー: {snapshotError}</div>}
      {saveMessage && <div className={styles.info}>{saveMessage}</div>}
      {analysisMessage && <div className={styles.info}>{analysisMessage}</div>}

      <ActionsSearchSection
        sort={sort}
        order={order}
        from={from}
        to={to}
        isDateRangeManual={isDateRangeManual}
        danbooruQuery={danbooruQuery}
        groupedTags={groupedTags}
        filterTagIDs={filterTagIDs}
        loading={loading}
        snapshotLoading={snapshotLoading}
        targetSaving={targetSaving}
        onSortChange={setSort}
        onOrderChange={setOrder}
        onFromChange={(value) => {
          setIsDateRangeManual(true);
          setFrom(value);
        }}
        onToChange={(value) => {
          setIsDateRangeManual(true);
          setTo(value);
        }}
        onDanbooruQueryChange={setDanbooruQuery}
        onResetDateRangeSync={resetDateRangeSync}
        onToggleFilterTag={toggleFilterTag}
        onSearch={() => void handleSearch()}
        onSaveAsTarget={() => void handleSaveAsTarget()}
        onClear={() => void handleClear()}
      />

      <ActionsOpenDataSection
        open={openDataPanelOpen}
        externalDataSource={externalDataSource}
        externalSignalType={externalSignalType}
        effectiveExternalSignalType={effectiveExternalSignalType}
        locationKey={locationKey}
        latitude={latitude}
        longitude={longitude}
        pastDays={pastDays}
        forecastDays={forecastDays}
        worldSignalLoading={worldSignalLoading}
        worldSignalError={worldSignalError}
        distributionError={distributionError}
        distributionDataset={distributionDataset}
        contextResult={contextResult}
        contextIsStale={contextIsStale}
        contextVizMode={contextVizMode}
        scatterXAxis={scatterXAxis}
        scatterYAxis={scatterYAxis}
        availableWorldAxisEntries={availableWorldAxisEntries}
        contextTimePoints={contextTimePoints}
        contextScatterPoints={contextScatterPoints}
        onToggleOpen={setOpenDataPanelOpen}
        onExternalDataSourceChange={setExternalDataSource}
        onExternalSignalTypeChange={setExternalSignalType}
        onLocationKeyChange={setLocationKey}
        onLatitudeChange={setLatitude}
        onLongitudeChange={setLongitude}
        onPastDaysChange={setPastDays}
        onForecastDaysChange={setForecastDays}
        onRefreshOpenDataView={() => void handleFetchWorldSignals()}
        onLoadAnalysisContext={() => void handleLoadAnalysisContext()}
        onSwitchDistributionDataset={switchDistributionDataset}
        onContextVizModeChange={setContextVizMode}
        onScatterXAxisChange={setScatterXAxis}
        onScatterYAxisChange={setScatterYAxis}
        signalOptions={eStatDashboardSignalOptions}
      />

      <ActionsDistributionSection
        distributionResult={distributionResult}
        distributionDataset={distributionDataset}
        distributionAxis={distributionAxis}
        distributionDatasetLabel={distributionDatasetLabel}
        distributionEmptyHint={distributionEmptyHint}
        selectedDatasetCount={selectedDatasetCount}
        distributionLoading={distributionLoading}
        selectedDistributionBin={selectedDistributionBin}
        selectedBinStatus={selectedBinStatus}
        actionAxisCandidates={actionAxisCandidates}
        availableWorldAxisEntries={availableWorldAxisEntries}
        setDistributionAxis={setDistributionAxis}
        switchDistributionDataset={switchDistributionDataset}
        runDistributionAnalysisFromCurrent={runDistributionAnalysisFromCurrent}
        runDistributionFromCurrentInputs={runDistributionFromCurrentInputs}
        appendSuggestedTagToQuery={appendSuggestedTagToQuery}
        selectDistributionBin={selectDistributionBin}
        analysisState={analysisState}
      />

      <ActionsHistorySection
        analysisHistory={analysisHistory}
        filteredSortedHistory={filteredSortedHistory}
        historySeverityFilter={historySeverityFilter}
        historySort={historySort}
        snapshotLoading={snapshotLoading}
        onHistorySeverityFilterChange={setHistorySeverityFilter}
        onHistorySortChange={setHistorySort}
        onClearSnapshots={() => void handleClearSnapshots()}
      />

      <ActionsResultsSection
        paginationTotal={pagination.total}
        actions={actions}
        actionAxisCandidates={actionAxisCandidates}
        actionScatterXAxis={actionScatterXAxis}
        actionScatterYAxis={actionScatterYAxis}
        actionScatterPoints={actionScatterPoints}
        prototypeNameMap={prototypeNameMap}
        formatTagNames={(tagIDs) => formatTagNames(tagIDs, tagNameMap)}
        loading={loading}
        onActionScatterXAxisChange={setActionScatterXAxis}
        onActionScatterYAxisChange={setActionScatterYAxis}
        onDelete={(id) => void handleDelete(id)}
      />
    </div>
  );
};

export default Actions;
