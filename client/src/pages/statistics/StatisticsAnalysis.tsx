import { useMemo } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useActions } from "../../hooks/useActions";
import useSearchTagSuggestions from "../../hooks/useSearchTagSuggestions";
import { useTags } from "../../hooks/useTags";
import {
  useWorldSignals,
  type ExternalDataSource,
} from "../../hooks/useWorldSignals";
import { eStatDashboardSignalOptions } from "../../constants/externalDataCatalog";
import SearchSuggestionRail from "../../components/search/SearchSuggestionRail";
import { buildTagNameToID } from "../constructions/actionsSearch/actionsSearchSectionHelpers";
import { appendTagToQueryText } from "../constructions/actionsSearch/actionsSearchController";
import styles from "../../assets/styles/StatisticsAnalysis.module.scss";
import {
  compareMeansController,
  correlationController,
  fetchExternalDataController,
  repeatBehaviorController,
  searchActionLogsController,
  ACTION_COMPARE_WINDOW_DAYS,
  ACTION_CORRELATION_WINDOW_DAYS,
  ESTAT_COMPARE_YEARS,
  ESTAT_CORRELATION_YEARS,
  actionAxisLabel,
  actionValue,
  createRange,
  createYearRange,
} from "./analysis";
import {
  StatisticsCompareSection,
  StatisticsCorrelationSection,
  StatisticsMessageSection,
  StatisticsRepeatSection,
} from "./sections";
import { useStatisticsAnalysisEffects } from "./effects";
import { useStatisticsAnalysisState } from "./state";
import type { DatasetKind, StatisticsNavigationState } from "./analysis";

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
  const {
    dataset,
    setDataset,
    source,
    setSource,
    locationKey,
    setLocationKey,
    latitude,
    setLatitude,
    longitude,
    setLongitude,
    pastDays,
    setPastDays,
    forecastDays,
    setForecastDays,
    signalType,
    setSignalType,
    actionKeyword,
    setActionKeyword,
    selectedTagIDs,
    compareFromA,
    setCompareFromA,
    compareToA,
    setCompareToA,
    compareFromB,
    setCompareFromB,
    compareToB,
    setCompareToB,
    correlationFrom,
    setCorrelationFrom,
    correlationTo,
    setCorrelationTo,
    repeatKeyword,
    setRepeatKeyword,
    repeatFrom,
    setRepeatFrom,
    repeatTo,
    setRepeatTo,
    repeatUnit,
    setRepeatUnit,
    repeatExpected,
    setRepeatExpected,
    actionAxisOptions,
    setActionAxisOptions,
    searchedActionLogs,
    setSearchedActionLogs,
    hasSearchedActionLogs,
    setHasSearchedActionLogs,
    compareActionAxis,
    setCompareActionAxis,
    correlationActionXAxis,
    setCorrelationActionXAxis,
    correlationActionYAxis,
    setCorrelationActionYAxis,
    compareWorldAxis,
    setCompareWorldAxis,
    correlationWorldXAxis,
    setCorrelationWorldXAxis,
    correlationWorldYAxis,
    setCorrelationWorldYAxis,
    compareResult,
    setCompareResult,
    correlationResult,
    setCorrelationResult,
    repeatResult,
    setRepeatResult,
    pageMessage,
    setPageMessage,
    pageError,
    setPageError,
  } = useStatisticsAnalysisState({
    navigationState,
    actionCompareRangeA,
    actionCompareRangeB,
    actionCorrelationRange,
    actionRepeatRange,
  });
  const tagNameToID = useMemo(() => buildTagNameToID(tags), [tags]);
  const actionSearchSuggestions = useSearchTagSuggestions(tags, actionKeyword, selectedTagIDs, 12);
  useStatisticsAnalysisEffects({
    fetchTags,
    fetchActionsSnapshot,
    compareActionAxis,
    setCompareActionAxis,
    correlationActionXAxis,
    setCorrelationActionXAxis,
    correlationActionYAxis,
    setCorrelationActionYAxis,
    setActionAxisOptions,
    dataset,
    source,
    setPageError,
    setPageMessage,
    setHasSearchedActionLogs,
    setSearchedActionLogs,
    setCompareResult,
    setCorrelationResult,
    setRepeatResult,
    locationKey,
    setLocationKey,
    latitude,
    setLatitude,
    longitude,
    setLongitude,
    setCompareWorldAxis,
    setCorrelationWorldXAxis,
    setCorrelationWorldYAxis,
    setCompareFromA,
    setCompareToA,
    setCompareFromB,
    setCompareToB,
    setCorrelationFrom,
    setCorrelationTo,
    eStatCompareRangeA,
    eStatCompareRangeB,
    eStatCorrelationRange,
    actionCompareRangeA,
    actionCompareRangeB,
    actionCorrelationRange,
  });

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
    setPageMessage("検索しています…");

    const result = await searchActionLogsController({
      actionKeyword,
      tagNameToID,
      selectedTagIDs,
      actionSearchWindow,
      fetchActionsSnapshot,
    });

    if (result.kind === "request_failed") return;
    if (result.kind === "error") {
      setPageError(result.error);
      return;
    }

    setHasSearchedActionLogs(true);
    setSearchedActionLogs(result.logs);
    setActionAxisOptions(result.options);
    if (!result.options.includes(compareActionAxis)) setCompareActionAxis("tag_count");
    if (!result.options.includes(correlationActionXAxis)) setCorrelationActionXAxis("tag_count");
    if (!result.options.includes(correlationActionYAxis)) {
      setCorrelationActionYAxis(
        result.options.includes("prototype_id") ? "prototype_id" : "tag_count"
      );
    }
    setPageMessage(result.message);
  };

  const handleFetchExternalData = async () => {
    setPageError(null);
    setPageMessage("外部データを取得しています…");

    const result = await fetchExternalDataController({
      source,
      locationKey,
      latitude,
      longitude,
      pastDays,
      forecastDays,
      selectedSignalLabel,
      fetchOpenMeteo,
    });

    if (result.kind === "request_failed") return;
    if (result.kind === "error") {
      setPageError(result.error);
      return;
    }

    setPageMessage(result.message);
  };

  const handleCompareMeans = async () => {
    setPageError(null);
    setPageMessage("平均値比較を実行しています…");
    setCompareResult(null);

    const result = await compareMeansController({
      dataset,
      compareFromA,
      compareToA,
      compareFromB,
      compareToB,
      compareActionAxis,
      compareWorldAxis,
      source,
      locationKey,
      signalType: currentSignalType,
      actionKeyword,
      tagNameToID,
      selectedTagIDs,
      fetchActionsSnapshot,
      fetchAnalysisContext,
    });

    if (result.kind === "request_failed") return;
    if (result.kind === "error") {
      setPageError(result.error);
      return;
    }

    setCompareResult(result.result);
    setPageMessage(result.message);
  };

  const handleCorrelation = async () => {
    setPageError(null);
    setPageMessage("相関分析を実行しています…");
    setCorrelationResult(null);

    const result = await correlationController({
      dataset,
      correlationFrom,
      correlationTo,
      correlationActionXAxis,
      correlationActionYAxis,
      correlationWorldXAxis,
      correlationWorldYAxis,
      source,
      locationKey,
      signalType: currentSignalType,
      actionKeyword,
      tagNameToID,
      selectedTagIDs,
      fetchActionsSnapshot,
      fetchAnalysisContext,
    });

    if (result.kind === "request_failed") return;
    if (result.kind === "error") {
      setPageError(result.error);
      return;
    }

    setCorrelationResult(result.result);
    setPageMessage(result.message);
  };

  const handleRepeatBehavior = async () => {
    setPageError(null);
    setPageMessage("反復行動の推定と検定を実行しています…");
    setRepeatResult(null);

    const result = await repeatBehaviorController({
      repeatFrom,
      repeatTo,
      repeatKeyword,
      repeatUnit,
      repeatExpected,
      actionKeyword,
      tagNameToID,
      selectedTagIDs,
      fetchActionsSnapshot,
    });

    if (result.kind === "request_failed") return;
    if (result.kind === "error") {
      setPageError(result.error);
      return;
    }

    setRepeatResult(result.result);
    setPageMessage(result.message);
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
        <StatisticsCompareSection
          dataset={dataset}
          compareFromA={compareFromA}
          setCompareFromA={setCompareFromA}
          compareToA={compareToA}
          setCompareToA={setCompareToA}
          compareFromB={compareFromB}
          setCompareFromB={setCompareFromB}
          compareToB={compareToB}
          setCompareToB={setCompareToB}
          compareActionAxis={compareActionAxis}
          setCompareActionAxis={setCompareActionAxis}
          compareWorldAxis={compareWorldAxis}
          setCompareWorldAxis={setCompareWorldAxis}
          actionAxisOptions={actionAxisOptions}
          source={source}
          selectedSignalLabel={selectedSignalLabel}
          loading={loading}
          handleCompareMeans={handleCompareMeans}
          compareResult={compareResult}
        />
        <StatisticsCorrelationSection
          dataset={dataset}
          correlationFrom={correlationFrom}
          setCorrelationFrom={setCorrelationFrom}
          correlationTo={correlationTo}
          setCorrelationTo={setCorrelationTo}
          correlationActionXAxis={correlationActionXAxis}
          setCorrelationActionXAxis={setCorrelationActionXAxis}
          correlationActionYAxis={correlationActionYAxis}
          setCorrelationActionYAxis={setCorrelationActionYAxis}
          correlationWorldXAxis={correlationWorldXAxis}
          setCorrelationWorldXAxis={setCorrelationWorldXAxis}
          correlationWorldYAxis={correlationWorldYAxis}
          setCorrelationWorldYAxis={setCorrelationWorldYAxis}
          actionAxisOptions={actionAxisOptions}
          source={source}
          selectedSignalLabel={selectedSignalLabel}
          loading={loading}
          handleCorrelation={handleCorrelation}
          correlationResult={correlationResult}
        />
      </div>
      <StatisticsRepeatSection
        repeatKeyword={repeatKeyword}
        setRepeatKeyword={setRepeatKeyword}
        repeatFrom={repeatFrom}
        setRepeatFrom={setRepeatFrom}
        repeatTo={repeatTo}
        setRepeatTo={setRepeatTo}
        repeatUnit={repeatUnit}
        setRepeatUnit={setRepeatUnit}
        repeatExpected={repeatExpected}
        setRepeatExpected={setRepeatExpected}
        handleRepeatBehavior={handleRepeatBehavior}
        loading={loading}
        repeatResult={repeatResult}
      />
      <StatisticsMessageSection
        title={"実行メモ"}
        pageMessage={pageMessage}
        pageError={pageError}
        error={error}
        showSuccess={pageMessage.includes("しました")}
      />
    </div>
  );
};

export default StatisticsAnalysis;












