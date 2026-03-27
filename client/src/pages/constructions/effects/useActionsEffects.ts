import { useEffect } from "react";
import type { DistributionDataset } from "../../../hooks/useDistributionAnalysis";
import type { ExternalDataSource } from "../../../hooks/useWorldSignals";

type UseActionsInitialLoadInput = {
  fetchTags: () => Promise<unknown>;
  fetchPrototypes: () => Promise<unknown>;
  loadAnalysisHistory: () => Promise<void>;
  setFrom: (value: string) => void;
  setTo: (value: string) => void;
  toLocalDateTimeValue: (value: Date) => string;
};

export const useActionsInitialLoad = ({
  fetchTags,
  fetchPrototypes,
  loadAnalysisHistory,
  setFrom,
  setTo,
  toLocalDateTimeValue,
}: UseActionsInitialLoadInput) => {
  useEffect(() => {
    void fetchTags();
    void fetchPrototypes();
    void loadAnalysisHistory();
    const now = new Date();
    const past = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
    setFrom(toLocalDateTimeValue(past));
    setTo(toLocalDateTimeValue(now));
  }, []);
};

type UseActionsDateRangeSyncInput = {
  pastDays: string;
  forecastDays: string;
  isDateRangeManual: boolean;
  setFrom: (value: string) => void;
  setTo: (value: string) => void;
  syncDateRangeWithOpenDataWindow: (pastDays: string, forecastDays: string) => { from: string; to: string } | null;
};

export const useActionsDateRangeSync = ({
  pastDays,
  forecastDays,
  isDateRangeManual,
  setFrom,
  setTo,
  syncDateRangeWithOpenDataWindow,
}: UseActionsDateRangeSyncInput) => {
  useEffect(() => {
    const synced = syncDateRangeWithOpenDataWindow(pastDays, forecastDays);
    if (!synced || isDateRangeManual) return;
    setFrom(synced.from);
    setTo(synced.to);
  }, [pastDays, forecastDays, isDateRangeManual, setFrom, setTo, syncDateRangeWithOpenDataWindow]);
};

type UseActionsExternalSourceSyncInput = {
  externalDataSource: ExternalDataSource;
  distributionDataset: DistributionDataset;
  distributionAxis: string;
  locationKey: string;
  setDistributionAxis: (value: string) => void;
  setLocationKey: (value: string) => void;
};

export const useActionsExternalSourceSync = ({
  externalDataSource,
  distributionDataset,
  distributionAxis,
  locationKey,
  setDistributionAxis,
  setLocationKey,
}: UseActionsExternalSourceSyncInput) => {
  useEffect(() => {
    if (externalDataSource === "e_stat_dashboard") {
      if (distributionDataset === "world_signals") {
        setDistributionAxis("signal_value");
      }
      if (locationKey === "tokyo_shinjuku") {
        setLocationKey("13000");
      }
      return;
    }

    if (distributionDataset === "world_signals" && distributionAxis === "signal_value") {
      setDistributionAxis("temperature_c");
    }
    if (locationKey === "13000") {
      setLocationKey("tokyo_shinjuku");
    }
  }, [
    externalDataSource,
    distributionDataset,
    distributionAxis,
    locationKey,
    setDistributionAxis,
    setLocationKey,
  ]);
};

type UseActionsAxisSyncInput = {
  actionAxisCandidates: string[];
  actionScatterXAxis: string;
  actionScatterYAxis: string;
  distributionAxis: string;
  distributionDataset: DistributionDataset;
  setActionScatterXAxis: (value: string) => void;
  setActionScatterYAxis: (value: string) => void;
  setDistributionAxis: (value: string) => void;
};

export const useActionsAxisSync = ({
  actionAxisCandidates,
  actionScatterXAxis,
  actionScatterYAxis,
  distributionAxis,
  distributionDataset,
  setActionScatterXAxis,
  setActionScatterYAxis,
  setDistributionAxis,
}: UseActionsAxisSyncInput) => {
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
  }, [
    actionAxisCandidates,
    actionScatterXAxis,
    actionScatterYAxis,
    distributionAxis,
    distributionDataset,
    setActionScatterXAxis,
    setActionScatterYAxis,
    setDistributionAxis,
  ]);
};
