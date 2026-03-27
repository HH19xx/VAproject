import { useState } from "react";
import type { ActionLog } from "../../../hooks/useActions";
import type { ExternalDataSource } from "../../../hooks/useWorldSignals";
import type {
  ActionAxisKey,
  CorrelationAnalysisResult,
  DatasetKind,
  RepeatBehaviorResult,
  RepeatBucketUnit,
  StatisticsNavigationState,
  TwoGroupAnalysisResult,
  WorldAxisKey,
} from "../analysis/analysisTypes";

type DateRange = {
  from: string;
  to: string;
};

type UseStatisticsAnalysisStateInput = {
  navigationState: StatisticsNavigationState;
  actionCompareRangeA: DateRange;
  actionCompareRangeB: DateRange;
  actionCorrelationRange: DateRange;
  actionRepeatRange: DateRange;
};

const useStatisticsAnalysisState = ({
  navigationState,
  actionCompareRangeA,
  actionCompareRangeB,
  actionCorrelationRange,
  actionRepeatRange,
}: UseStatisticsAnalysisStateInput) => {
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
  const [correlationFrom, setCorrelationFrom] = useState(
    navigationState.from ?? actionCorrelationRange.from
  );
  const [correlationTo, setCorrelationTo] = useState(
    navigationState.to ?? actionCorrelationRange.to
  );
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
  const [correlationActionXAxis, setCorrelationActionXAxis] =
    useState<ActionAxisKey>("tag_count");
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
  const [pageMessage, setPageMessage] = useState("まだ分析を実行していません。");
  const [pageError, setPageError] = useState<string | null>(null);

  return {
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
  };
};

export default useStatisticsAnalysisState;
