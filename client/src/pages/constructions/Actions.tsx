import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useActions } from "../../hooks/useActions";
import { usePrototypes } from "../../hooks/usePrototypes";
import { useTags } from "../../hooks/useTags";
import { useTargets } from "../../hooks/useTargets";
import {
  useAnalysisSnapshots,
} from "../../hooks/useAnalysisSnapshots";
import {
  useDistributionAnalysis,
  type DistributionAnalysisResult,
  type DistributionDataset,
} from "../../hooks/useDistributionAnalysis";
import {
  useWorldSignals,
} from "../../hooks/useWorldSignals";
import { eStatDashboardSignalOptions } from "../../constants/externalDataCatalog";
import styles from "../../assets/styles/Actions.module.scss";
import useSearchTagSuggestions from "../../hooks/useSearchTagSuggestions";
import ActionsSearchSection from "./ActionsSearchSection";
import ActionsOpenDataSection from "./ActionsOpenDataSection";
import ActionsDistributionSection from "./ActionsDistributionSection";
import ActionsHistorySection from "./ActionsHistorySection";
import ActionsResultsSection from "./ActionsResultsSection";
import {
  type ParsedQuery,
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
} from "./actionsSearch/actionsSearchSectionHelpers";
import {
  appendTagToQueryText,
  buildActionSearchOptions,
  buildClearedSearchState,
  buildMergedAndTagIDs,
  buildParsedSearchState,
  buildSavedSearchPayload,
  buildSearchRange,
  confirmDeleteAction,
} from "./actionsSearch/actionsSearchController";
import {
  buildContextStaleState,
} from "./actionsOpenData/actionsOpenDataSectionHelpers";
import { buildAvailableWorldAxisEntries } from "./actionsDistribution/externalAxisLabels";
import { ACTION_AXIS_DEFAULTS, actionAxisValue, defaultAxisByDataset } from "./actionsDistribution/actionsDistributionAxes";
import { calculateDistributionStats, deriveSeverity } from "./actionsDistribution/actionsDistributionScore";
import {
  buildFilteredSortedHistory,
  clearAnalysisHistory,
  loadAnalysisHistory as loadAnalysisHistoryEntries,
} from "./actionsHistory/actionsHistoryController";
import { loadAnalysisContext, validateWorldSignalInputs } from "./actionsOpenData/actionsOpenDataController";
import {
  buildDistributionExecutionState,
  buildResetDateRangeState,
  buildSelectedBinStatus,
  executeDistributionAnalysis,
  getDistributionDatasetLabel,
  getDistributionEmptyHint,
  getSelectedDatasetCount,
  resolveBaselineWindow,
  resolveNextDistributionAxis,
  saveDistributionSnapshot,
  selectDistributionBin as selectDistributionBinEntry,
} from "./actionsDistribution/actionsDistributionController";
import useActionsSearchState from "./state/useActionsSearchState";
import useActionsDistributionState from "./state/useActionsDistributionState";
import useActionsOpenDataState from "./state/useActionsOpenDataState";
import useActionsHistoryState from "./state/useActionsHistoryState";
import {
  useActionsAxisSync,
  useActionsDateRangeSync,
  useActionsExternalSourceSync,
  useActionsInitialLoad,
} from "./effects/useActionsEffects";

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


  const {
    sort,
    setSort,
    order,
    setOrder,
    from,
    setFrom,
    to,
    setTo,
    isDateRangeManual,
    setIsDateRangeManual,
    filterTagIDs,
    setFilterTagIDs,
    danbooruQuery,
    setDanbooruQuery,
    searchedActionLogs,
    setSearchedActionLogs,
    queryError,
    setQueryError,
    saveMessage,
    setSaveMessage,
  } = useActionsSearchState();
  const {
    analysisState,
    setAnalysisState,
    distributionDataset,
    setDistributionDataset,
    distributionAxis,
    setDistributionAxis,
    distributionResult,
    setDistributionResult,
    selectedDistributionBin,
    setSelectedDistributionBin,
    analysisMessage,
    setAnalysisMessage,
  } = useActionsDistributionState();
  const {
    locationKey,
    setLocationKey,
    externalDataSource,
    setExternalDataSource,
    externalSignalType,
    setExternalSignalType,
    latitude,
    setLatitude,
    longitude,
    setLongitude,
    pastDays,
    setPastDays,
    forecastDays,
    setForecastDays,
    contextResult,
    setContextResult,
    openDataPanelOpen,
    setOpenDataPanelOpen,
    actionScatterXAxis,
    setActionScatterXAxis,
    actionScatterYAxis,
    setActionScatterYAxis,
  } = useActionsOpenDataState();
  const {
    analysisHistory,
    setAnalysisHistory,
    historySeverityFilter,
    setHistorySeverityFilter,
    historySort,
    setHistorySort,
  } = useActionsHistoryState();

  useActionsInitialLoad({
    fetchTags,
    fetchPrototypes,
    loadAnalysisHistory,
    setFrom,
    setTo,
    toLocalDateTimeValue,
  });

  useActionsDateRangeSync({
    pastDays,
    forecastDays,
    isDateRangeManual,
    setFrom,
    setTo,
    syncDateRangeWithOpenDataWindow,
  });

  useActionsExternalSourceSync({
    externalDataSource,
    distributionDataset,
    distributionAxis,
    locationKey,
    setDistributionAxis,
    setLocationKey,
  });

  const effectiveExternalSignalType = externalDataSource === "e_stat_dashboard" ? externalSignalType : "";

  async function loadAnalysisHistory() {
    const snapshots = await loadAnalysisHistoryEntries(listSnapshots, 100);
    setAnalysisHistory(snapshots);
  }

  const tagNameMap = useMemo(() => buildTagNameMap(tags), [tags]);
  const tagNameToID = useMemo(() => buildTagNameToID(tags), [tags]);
  const prototypeNameMap = useMemo(() => buildPrototypeNameMap(prototypes), [prototypes]);
  const groupedTags = useMemo(() => buildGroupedTags(tags), [tags]);
  const searchSuggestions = useSearchTagSuggestions(tags, danbooruQuery, filterTagIDs, 12);
  const actionAxisCandidates = useMemo(() => buildActionAxisCandidates(meta, [...ACTION_AXIS_DEFAULTS]), [meta]);

  useActionsAxisSync({
    actionAxisCandidates,
    actionScatterXAxis,
    actionScatterYAxis,
    distributionAxis,
    distributionDataset,
    setActionScatterXAxis,
    setActionScatterYAxis,
    setDistributionAxis,
  });

  const toggleFilterTag = (tagID: number) => {
    setFilterTagIDs((prev) => (prev.includes(tagID) ? prev.filter((id) => id !== tagID) : [...prev, tagID]));
  };

  const appendSearchSuggestion = (tagName: string) => {
    const next = appendTagToQueryText(danbooruQuery, tagName);
    if (!next.alreadyIncluded) {
      setDanbooruQuery(next.updatedQuery);
    }
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

    const baselineWindow = isValidDateRange(from, to) ? resolveBaselineWindow(from, to) : {};
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

  const saveDistributionResultSnapshot = async (result: DistributionAnalysisResult) => {
    const saved = await saveDistributionSnapshot(createSnapshot, danbooruQuery, result);

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
    const result = await executeDistributionAnalysis(analyzeDistribution, {
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
      await saveDistributionResultSnapshot(result);
    }
  };

  const runDistributionAnalysisFromCurrent = async () => {
    const execution = buildDistributionExecutionState(
      danbooruQuery,
      tagNameToID,
      filterTagIDs,
      from,
      to,
      parseDanbooruStyleQuery
    );
    if (!execution.parsed) {
      setQueryError(execution.error);
      return;
    }
    setQueryError(null);
    await runDistributionAnalysis(execution.mergedAnd, execution.parsed, execution.range.from, execution.range.to);
  };

  const runDistributionFromCurrentInputs = async (datasetOverride?: DistributionDataset, axisOverride?: string) => {
    const execution = buildDistributionExecutionState(
      danbooruQuery,
      tagNameToID,
      filterTagIDs,
      from,
      to,
      parseDanbooruStyleQuery
    );
    if (!execution.parsed) {
      setQueryError(execution.error);
      return;
    }
    setQueryError(null);
    await runDistributionAnalysis(
      execution.mergedAnd,
      execution.parsed,
      execution.range.from,
      execution.range.to,
      datasetOverride,
      axisOverride
    );
  };

  const resolveSearchRange = (queryText: string) =>
    queryText.trim() || filterTagIDs.length > 0 ? buildSearchRange(from, to) : {};

  const handleSearch = async (queryOverride?: string) => {
    const activeQuery = queryOverride ?? danbooruQuery;
    const { parsed, error } = buildParsedSearchState(activeQuery, tagNameToID, parseDanbooruStyleQuery);
    if (!parsed) {
      setQueryError(error);
      return;
    }

    setQueryError(null);
    setSaveMessage(null);
    setAnalysisMessage(null);
    setSelectedDistributionBin(null);

    const mergedAnd = buildMergedAndTagIDs(filterTagIDs, parsed);
    const range = resolveSearchRange(activeQuery);

    await fetchActions(1, 10, mergedAnd, buildActionSearchOptions(sort, order, parsed, range));
    const fullSnapshot = await fetchActionsSnapshot(1, 500, mergedAnd, buildActionSearchOptions(sort, order, parsed, range));
    setSearchedActionLogs(fullSnapshot?.logs || []);
    await analyzeCurrentResult(mergedAnd, parsed, range.from, range.to);
    await runDistributionAnalysis(mergedAnd, parsed, range.from, range.to);
  };

  const handleChangeResultsPage = async (page: number) => {
    if (page < 1 || page === pagination.page) return;
    const { parsed, error } = buildParsedSearchState(danbooruQuery, tagNameToID, parseDanbooruStyleQuery);
    if (!parsed) {
      setQueryError(error);
      return;
    }
    const mergedAnd = buildMergedAndTagIDs(filterTagIDs, parsed);
    const range = resolveSearchRange(danbooruQuery);
    const fullSnapshot = await fetchActionsSnapshot(1, 500, mergedAnd, buildActionSearchOptions(sort, order, parsed, range));
    if (fullSnapshot) {
      setSearchedActionLogs(fullSnapshot.logs);
    }
    await fetchActions(page, 10, mergedAnd, buildActionSearchOptions(sort, order, parsed, range));
  };

  const appendSuggestedTagToQuery = async (tag: string, mode: "append" | "search" = "append") => {
    const appended = appendTagToQueryText(danbooruQuery, tag);
    if (appended.alreadyIncluded) {
      setAnalysisMessage(`タグ ${tag} はすでにクエリへ含まれています。`);
      if (mode === "search") {
        await handleSearch();
      }
      return;
    }

    const updated = appended.updatedQuery;
    setDanbooruQuery(updated);

    if (mode === "search") {
      setAnalysisMessage(`タグ ${tag} を追加して再検索します。`);
      await handleSearch(updated);
      return;
    }

    setAnalysisMessage(`タグ ${tag} をクエリに追加しました。`);
  };

  const selectDistributionBin = (index: number) => {
    const selected = selectDistributionBinEntry(distributionResult, index);
    if (!selected) return;
    setSelectedDistributionBin(selected);
    setAnalysisMessage("補完グラフの帯域を選択しました。候補タグや候補軸の試行に使えます。");
  };

  const handleClear = async () => {
    const cleared = buildClearedSearchState();
    setFilterTagIDs(cleared.filterTagIDs);
    setDanbooruQuery(cleared.danbooruQuery);
    setQueryError(cleared.queryError);
    setSaveMessage(cleared.saveMessage);
    setAnalysisMessage(cleared.analysisMessage);
    setAnalysisState(cleared.analysisState);
    setDistributionResult(cleared.distributionResult);
    setSelectedDistributionBin(cleared.selectedDistributionBin);
    setSort(cleared.sort);
    setOrder(cleared.order);
    setIsDateRangeManual(cleared.isDateRangeManual);
    setFrom(cleared.from);
    setTo(cleared.to);
    setSearchedActionLogs(null);
  };

  const handleSaveAsTarget = async () => {
    const { parsed, error } = buildParsedSearchState(danbooruQuery, tagNameToID, parseDanbooruStyleQuery);
    if (!parsed) {
      setQueryError(error);
      return;
    }
    setQueryError(null);

    const mergedAnd = buildMergedAndTagIDs(filterTagIDs, parsed);
    if (mergedAnd.length === 0 && parsed.anyTagIDs.length === 0 && parsed.anyTagGroups.length === 0) {
      setSaveMessage("保存するには、少なくとも1つ以上の条件が必要です。");
      return;
    }

    const suggestedName = `saved-search-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, "-")}`;
    const name = window.prompt("保存名を入力してください", suggestedName)?.trim();
    if (!name) return;

    const savedSearch = buildSavedSearchPayload(danbooruQuery, locationKey);
    const created = await createTarget(name, savedSearch.description, mergedAnd, {
      queryText: savedSearch.queryText,
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
    const ok = await clearAnalysisHistory(clearSnapshots);
    if (ok) {
      setAnalysisHistory([]);
      setAnalysisMessage("分析履歴を削除しました。");
    }
  };


  const handleLoadAnalysisContext = async () => {
    const { result, message } = await loadAnalysisContext({
      fetchAnalysisContext,
      externalDataSource,
      locationKey,
      effectiveExternalSignalType,
      from,
      to,
    });
    if (result) {
      setContextResult(result);
      if (result.summary.signal_count > 0 && result.summary.action_count === 0) {
        setDistributionDataset("world_signals");
      }
    }
    setAnalysisMessage(message);
  };

  const handleFetchWorldSignals = async () => {
    const validated = validateWorldSignalInputs({
      externalDataSource,
      locationKey,
      latitude,
      longitude,
      pastDays,
      forecastDays,
    });
    if (!validated.values) {
      if (validated.message) {
        setAnalysisMessage(validated.message);
      }
      return;
    }

    const ok = await fetchOpenMeteo({
      source: externalDataSource,
      locationKey: locationKey.trim(),
      latitude: validated.values.lat,
      longitude: validated.values.lon,
      pastDays: validated.values.pDays,
      forecastDays: validated.values.fDays,
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
    if (!confirmDeleteAction(window.confirm)) return;
    await deleteAction(id);
  };

  const availableWorldAxisEntries = useMemo(
    () => buildAvailableWorldAxisEntries(externalDataSource),
    [externalDataSource]
  );

  const filteredSortedHistory = useMemo(() => {
    return buildFilteredSortedHistory(analysisHistory, historySeverityFilter, historySort);
  }, [analysisHistory, historySeverityFilter, historySort]);

  const actionScatterPoints = useMemo(
    () => buildActionScatterPoints(searchedActionLogs ?? [], actionScatterXAxis, actionScatterYAxis, actionAxisValue),
    [searchedActionLogs, actionScatterXAxis, actionScatterYAxis]
  );

  const distributionDatasetLabel = getDistributionDatasetLabel(distributionDataset);
  const selectedDatasetCount = getSelectedDatasetCount(distributionDataset, contextResult, actions.length);
  const distributionEmptyHint = getDistributionEmptyHint(distributionDataset, selectedDatasetCount);
  const selectedBinStatus = buildSelectedBinStatus(selectedDistributionBin);

  const switchDistributionDataset = (dataset: DistributionDataset) => {
    setDistributionDataset(dataset);
    setDistributionAxis(
      resolveNextDistributionAxis(dataset, externalDataSource, actionAxisCandidates, defaultAxisByDataset)
    );
    setDistributionResult(null);
    setSelectedDistributionBin(null);
  };

  const resetDateRangeSync = () => {
    const synced = buildResetDateRangeState(syncDateRangeWithOpenDataWindow, pastDays, forecastDays);
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
        <h1 className={styles.title}>Déconstruction機能</h1>
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
        searchSuggestions={searchSuggestions}
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
        onAppendSearchSuggestion={appendSearchSuggestion}
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
        signalOptions={eStatDashboardSignalOptions}
      />

      {searchedActionLogs !== null && (
        <ActionsResultsSection
          currentPage={pagination.page}
          pageSize={pagination.limit}
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
          onPageChange={(page) => void handleChangeResultsPage(page)}
          onDelete={(id) => void handleDelete(id)}
        />
      )}

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
      <div className={styles.footer}>
        <button className={styles.backButton} onClick={() => navigate("/dashboard")}>
          ダッシュボードへ戻る
        </button>
      </div>
    </div>
  );
};

export default Actions;






