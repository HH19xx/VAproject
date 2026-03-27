import { useEffect } from "react";
import type { ActionLog } from "../../../hooks/useActions";
import type { ExternalDataSource } from "../../../hooks/useWorldSignals";
import {
  RECENT_ACTION_LOOKBACK_DAYS,
  collectActionAxisOptions,
  createRange,
} from "../analysis/analysisHelpers";
import type {
  ActionAxisKey,
  CorrelationAnalysisResult,
  DatasetKind,
  RepeatBehaviorResult,
  TwoGroupAnalysisResult,
  WorldAxisKey,
} from "../analysis/analysisTypes";

type FetchActionsSnapshotFn = (
  page?: number,
  limit?: number,
  tagIDs?: number[],
  options?: {
    sort?: "occurred_at" | "created_at" | "updated_at" | "title" | "tag_count";
    order?: "asc" | "desc";
    from?: string;
    to?: string;
    anyTagIDs?: number[];
    anyTagGroups?: number[][];
    excludeTagIDs?: number[];
  }
) => Promise<{ logs: ActionLog[] } | null>;

type StatisticsEffectsInput = {
  fetchTags: () => Promise<unknown>;
  fetchActionsSnapshot: FetchActionsSnapshotFn;
  compareActionAxis: ActionAxisKey;
  setCompareActionAxis: (value: ActionAxisKey) => void;
  correlationActionXAxis: ActionAxisKey;
  setCorrelationActionXAxis: (value: ActionAxisKey) => void;
  correlationActionYAxis: ActionAxisKey;
  setCorrelationActionYAxis: (value: ActionAxisKey) => void;
  setActionAxisOptions: (value: ActionAxisKey[]) => void;
  dataset: DatasetKind;
  source: ExternalDataSource;
  setPageError: (value: string | null) => void;
  setPageMessage: (value: string) => void;
  setHasSearchedActionLogs: (value: boolean) => void;
  setSearchedActionLogs: (value: ActionLog[]) => void;
  setCompareResult: (value: TwoGroupAnalysisResult | null) => void;
  setCorrelationResult: (value: CorrelationAnalysisResult | null) => void;
  setRepeatResult: (value: RepeatBehaviorResult | null) => void;
  locationKey: string;
  setLocationKey: (value: string) => void;
  latitude: string;
  setLatitude: (value: string) => void;
  longitude: string;
  setLongitude: (value: string) => void;
  setCompareWorldAxis: (value: WorldAxisKey) => void;
  setCorrelationWorldXAxis: (value: WorldAxisKey) => void;
  setCorrelationWorldYAxis: (value: WorldAxisKey) => void;
  setCompareFromA: (value: string) => void;
  setCompareToA: (value: string) => void;
  setCompareFromB: (value: string) => void;
  setCompareToB: (value: string) => void;
  setCorrelationFrom: (value: string) => void;
  setCorrelationTo: (value: string) => void;
  eStatCompareRangeA: { from: string; to: string };
  eStatCompareRangeB: { from: string; to: string };
  eStatCorrelationRange: { from: string; to: string };
  actionCompareRangeA: { from: string; to: string };
  actionCompareRangeB: { from: string; to: string };
  actionCorrelationRange: { from: string; to: string };
};

const useStatisticsAnalysisEffects = ({
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
}: StatisticsEffectsInput) => {
  useEffect(() => {
    void fetchTags();
  }, [fetchTags]);

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
  }, [
    fetchActionsSnapshot,
    compareActionAxis,
    setCompareActionAxis,
    correlationActionXAxis,
    setCorrelationActionXAxis,
    correlationActionYAxis,
    setCorrelationActionYAxis,
    setActionAxisOptions,
  ]);

  useEffect(() => {
    setPageError(null);
    setPageMessage("まだ分析を実行していません。");
    setHasSearchedActionLogs(false);
    setSearchedActionLogs([]);
    setCompareResult(null);
    setCorrelationResult(null);
    setRepeatResult(null);
  }, [
    dataset,
    source,
    setPageError,
    setPageMessage,
    setHasSearchedActionLogs,
    setSearchedActionLogs,
    setCompareResult,
    setCorrelationResult,
    setRepeatResult,
  ]);

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
  }, [
    source,
    locationKey,
    latitude,
    longitude,
    setLocationKey,
    setLatitude,
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
  ]);
};

export default useStatisticsAnalysisEffects;
