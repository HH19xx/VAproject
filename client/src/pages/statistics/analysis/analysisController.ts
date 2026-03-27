import type { ActionLog } from "../../../hooks/useActions";
import type {
  AnalysisContextResult,
  ExternalDataSource,
  FetchWorldSignalInput,
} from "../../../hooks/useWorldSignals";
import { parseDanbooruStyleQuery } from "../../constructions/actionsSearch/actionsSearchSectionHelpers";
import {
  actionValue,
  collectActionAxisOptions,
  computeCorrelation,
  computeRepeatBehavior,
  computeWelchTest,
  worldValue,
} from "./analysisHelpers";
import type {
  ActionAxisKey,
  CorrelationAnalysisResult,
  RepeatBehaviorResult,
  RepeatBucketUnit,
  TwoGroupAnalysisResult,
  WorldAxisKey,
} from "./analysisTypes";

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

type FetchAnalysisContextFn = (
  source: ExternalDataSource,
  locationKey: string,
  signalType: string,
  fromISO: string,
  toISO: string,
  limit?: number
) => Promise<AnalysisContextResult | null>;

type FetchActionValuesResult =
  | { kind: "request_failed" }
  | { kind: "error"; error: string }
  | { kind: "success"; logs: ActionLog[]; values: number[]; options: ActionAxisKey[] };

type FetchWorldValuesResult =
  | { kind: "request_failed" }
  | { kind: "success"; values: number[] };

type ControllerErrorResult = { kind: "error"; error: string };
type ControllerRequestFailedResult = { kind: "request_failed" };

type SearchActionLogsResult =
  | ControllerErrorResult
  | ControllerRequestFailedResult
  | { kind: "success"; logs: ActionLog[]; options: ActionAxisKey[]; message: string };

type CompareMeansResult =
  | ControllerErrorResult
  | ControllerRequestFailedResult
  | { kind: "success"; result: TwoGroupAnalysisResult; message: string };

type CorrelationResult =
  | ControllerErrorResult
  | ControllerRequestFailedResult
  | { kind: "success"; result: CorrelationAnalysisResult; message: string };

type RepeatBehaviorControllerResult =
  | ControllerErrorResult
  | ControllerRequestFailedResult
  | { kind: "success"; result: RepeatBehaviorResult; message: string };

type FetchExternalDataResult =
  | ControllerErrorResult
  | { kind: "success"; message: string }
  | ControllerRequestFailedResult;

type ParsedActionSearchInput = {
  actionKeyword: string;
  tagNameToID: Map<string, number>;
  selectedTagIDs: number[];
  fetchActionsSnapshot: FetchActionsSnapshotFn;
};

const fetchActionValuesController = async ({
  actionKeyword,
  tagNameToID,
  selectedTagIDs,
  fetchActionsSnapshot,
  from,
  to,
  axis,
}: ParsedActionSearchInput & {
  from: string;
  to: string;
  axis: ActionAxisKey;
}): Promise<FetchActionValuesResult> => {
  const parsed = parseDanbooruStyleQuery(actionKeyword, tagNameToID);
  if (parsed.unknownTokens.length > 0) {
    return {
      kind: "error",
      error: "未登録タグがあります: " + parsed.unknownTokens.join(", "),
    };
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
  if (!snapshot) return { kind: "request_failed" };

  const values = snapshot.logs
    .map((log) => actionValue(log, axis))
    .filter((value): value is number => typeof value === "number" && Number.isFinite(value));

  return {
    kind: "success",
    logs: snapshot.logs,
    values,
    options: collectActionAxisOptions(snapshot.logs),
  };
};

const fetchWorldValuesController = async ({
  source,
  locationKey,
  signalType,
  from,
  to,
  axis,
  fetchAnalysisContext,
}: {
  source: ExternalDataSource;
  locationKey: string;
  signalType: string;
  from: string;
  to: string;
  axis: WorldAxisKey;
  fetchAnalysisContext: FetchAnalysisContextFn;
}): Promise<FetchWorldValuesResult> => {
  const context = await fetchAnalysisContext(
    source,
    locationKey.trim(),
    signalType,
    new Date(from).toISOString(),
    new Date(to).toISOString(),
    500
  );
  if (!context) return { kind: "request_failed" };

  const values = context.world_signals
    .map((signal) => worldValue(signal, axis))
    .filter((value): value is number => typeof value === "number" && Number.isFinite(value));

  return {
    kind: "success",
    values,
  };
};

export const searchActionLogsController = async ({
  actionKeyword,
  tagNameToID,
  selectedTagIDs,
  actionSearchWindow,
  fetchActionsSnapshot,
}: ParsedActionSearchInput & {
  actionSearchWindow: { from: string; to: string } | null;
}): Promise<SearchActionLogsResult> => {
  const parsed = parseDanbooruStyleQuery(actionKeyword, tagNameToID);
  if (parsed.unknownTokens.length > 0) {
    return {
      kind: "error",
      error: "未登録タグがあります: " + parsed.unknownTokens.join(", "),
    };
  }
  if (!actionSearchWindow) {
    return {
      kind: "error",
      error: "検索に必要な期間が不正です。",
    };
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
  if (!snapshot) return { kind: "request_failed" };

  return {
    kind: "success",
    logs: snapshot.logs,
    options: collectActionAxisOptions(snapshot.logs),
    message: "行動記録の検索結果を更新しました。件数=" + snapshot.logs.length,
  };
};

export const fetchExternalDataController = async ({
  source,
  locationKey,
  latitude,
  longitude,
  pastDays,
  forecastDays,
  selectedSignalLabel,
  fetchOpenMeteo,
}: {
  source: ExternalDataSource;
  locationKey: string;
  latitude: string;
  longitude: string;
  pastDays: string;
  forecastDays: string;
  selectedSignalLabel: string;
  fetchOpenMeteo: (input: FetchWorldSignalInput) => Promise<boolean>;
}): Promise<FetchExternalDataResult> => {
  const payload: FetchWorldSignalInput = {
    source,
    locationKey: locationKey.trim(),
    latitude: source === "open_meteo" ? Number(latitude) : 0,
    longitude: source === "open_meteo" ? Number(longitude) : 0,
    pastDays: source === "open_meteo" ? Number(pastDays) : 0,
    forecastDays: source === "open_meteo" ? Number(forecastDays) : 0,
  };

  if (!payload.locationKey) {
    return {
      kind: "error",
      error:
        source === "e_stat_dashboard"
          ? "e-Stat Dashboard では都道府県コードを入力してください。例: 13000"
          : "location_key を入力してください。",
    };
  }

  if (
    source === "open_meteo" &&
    (!Number.isFinite(payload.latitude) ||
      !Number.isFinite(payload.longitude) ||
      !Number.isFinite(payload.pastDays) ||
      !Number.isFinite(payload.forecastDays))
  ) {
    return {
      kind: "error",
      error: "Open-Meteo では緯度・経度・past_days・forecast_days を数値で入力してください。",
    };
  }

  const ok = await fetchOpenMeteo(payload);
  if (!ok) {
    return {
      kind: "error",
      error: "外部データの取得に失敗しました。API設定と入力値を確認してください。",
    };
  }

  return {
    kind: "success",
    message:
      source === "e_stat_dashboard"
        ? "e-Stat Dashboard の " + selectedSignalLabel + " を取得しました。"
        : "Open-Meteo の外部データを取得しました。",
  };
};

export const compareMeansController = async ({
  dataset,
  compareFromA,
  compareToA,
  compareFromB,
  compareToB,
  compareActionAxis,
  compareWorldAxis,
  source,
  locationKey,
  signalType,
  actionKeyword,
  tagNameToID,
  selectedTagIDs,
  fetchActionsSnapshot,
  fetchAnalysisContext,
}: {
  dataset: "action_logs" | "world_signals";
  compareFromA: string;
  compareToA: string;
  compareFromB: string;
  compareToB: string;
  compareActionAxis: ActionAxisKey;
  compareWorldAxis: WorldAxisKey;
  source: ExternalDataSource;
  locationKey: string;
  signalType: string;
  actionKeyword: string;
  tagNameToID: Map<string, number>;
  selectedTagIDs: number[];
  fetchActionsSnapshot: FetchActionsSnapshotFn;
  fetchAnalysisContext: FetchAnalysisContextFn;
}): Promise<CompareMeansResult> => {
  if (dataset === "action_logs") {
    const groupA = await fetchActionValuesController({
      actionKeyword,
      tagNameToID,
      selectedTagIDs,
      fetchActionsSnapshot,
      from: compareFromA,
      to: compareToA,
      axis: compareActionAxis,
    });
    if (groupA.kind !== "success") return groupA;

    const groupB = await fetchActionValuesController({
      actionKeyword,
      tagNameToID,
      selectedTagIDs,
      fetchActionsSnapshot,
      from: compareFromB,
      to: compareToB,
      axis: compareActionAxis,
    });
    if (groupB.kind !== "success") return groupB;

    const result = computeWelchTest(groupA.values, groupB.values);
    if (!result) {
      return {
        kind: "error",
        error: "2期間比較には、各期間で2件以上の数値データが必要です。",
      };
    }

    return {
      kind: "success",
      result,
      message:
        "行動記録の平均値比較を更新しました。対象件数 A=" +
        groupA.values.length +
        " / B=" +
        groupB.values.length,
    };
  }

  const groupA = await fetchWorldValuesController({
    source,
    locationKey,
    signalType,
    from: compareFromA,
    to: compareToA,
    axis: compareWorldAxis,
    fetchAnalysisContext,
  });
  if (groupA.kind !== "success") return groupA;

  const groupB = await fetchWorldValuesController({
    source,
    locationKey,
    signalType,
    from: compareFromB,
    to: compareToB,
    axis: compareWorldAxis,
    fetchAnalysisContext,
  });
  if (groupB.kind !== "success") return groupB;

  const result = computeWelchTest(groupA.values, groupB.values);
  if (!result) {
    return {
      kind: "error",
      error:
        source === "e_stat_dashboard"
          ? "e-Stat は年次データなので、各期間に複数年が入るよう期間を広げてください。"
          : "2期間比較には、各期間で2件以上の数値データが必要です。",
    };
  }

  return {
    kind: "success",
    result,
    message:
      "外部データの平均値比較を更新しました。対象件数 A=" +
      groupA.values.length +
      " / B=" +
      groupB.values.length,
  };
};

export const correlationController = async ({
  dataset,
  correlationFrom,
  correlationTo,
  correlationActionXAxis,
  correlationActionYAxis,
  correlationWorldXAxis,
  correlationWorldYAxis,
  source,
  locationKey,
  signalType,
  actionKeyword,
  tagNameToID,
  selectedTagIDs,
  fetchActionsSnapshot,
  fetchAnalysisContext,
}: {
  dataset: "action_logs" | "world_signals";
  correlationFrom: string;
  correlationTo: string;
  correlationActionXAxis: ActionAxisKey;
  correlationActionYAxis: ActionAxisKey;
  correlationWorldXAxis: WorldAxisKey;
  correlationWorldYAxis: WorldAxisKey;
  source: ExternalDataSource;
  locationKey: string;
  signalType: string;
  actionKeyword: string;
  tagNameToID: Map<string, number>;
  selectedTagIDs: number[];
  fetchActionsSnapshot: FetchActionsSnapshotFn;
  fetchAnalysisContext: FetchAnalysisContextFn;
}): Promise<CorrelationResult> => {
  if (dataset === "action_logs") {
    const actionData = await fetchActionValuesController({
      actionKeyword,
      tagNameToID,
      selectedTagIDs,
      fetchActionsSnapshot,
      from: correlationFrom,
      to: correlationTo,
      axis: correlationActionXAxis,
    });
    if (actionData.kind !== "success") return actionData;

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
      return {
        kind: "error",
        error: "相関分析には、同一期間内で3件以上の数値ペアが必要です。",
      };
    }

    return {
      kind: "success",
      result,
      message: "行動記録の相関分析を更新しました。対象ペア数=" + result.count,
    };
  }

  const context = await fetchAnalysisContext(
    source,
    locationKey.trim(),
    signalType,
    new Date(correlationFrom).toISOString(),
    new Date(correlationTo).toISOString(),
    500
  );
  if (!context) return { kind: "request_failed" };

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
    return {
      kind: "error",
      error: "相関分析には、同一期間内で3件以上の数値ペアが必要です。",
    };
  }

  return {
    kind: "success",
    result,
    message: "外部データの相関分析を更新しました。対象ペア数=" + result.count,
  };
};

export const repeatBehaviorController = async ({
  repeatFrom,
  repeatTo,
  repeatKeyword,
  repeatUnit,
  repeatExpected,
  actionKeyword,
  tagNameToID,
  selectedTagIDs,
  fetchActionsSnapshot,
}: {
  repeatFrom: string;
  repeatTo: string;
  repeatKeyword: string;
  repeatUnit: RepeatBucketUnit;
  repeatExpected: string;
  actionKeyword: string;
  tagNameToID: Map<string, number>;
  selectedTagIDs: number[];
  fetchActionsSnapshot: FetchActionsSnapshotFn;
}): Promise<RepeatBehaviorControllerResult> => {
  const actionData = await fetchActionValuesController({
    actionKeyword,
    tagNameToID,
    selectedTagIDs,
    fetchActionsSnapshot,
    from: repeatFrom,
    to: repeatTo,
    axis: "tag_count",
  });
  if (actionData.kind !== "success") return actionData;

  const expected = Number(repeatExpected);
  if (!Number.isFinite(expected)) {
    return {
      kind: "error",
      error: "期待値には数値を入力してください。",
    };
  }

  const result = computeRepeatBehavior(actionData.logs, repeatKeyword, repeatUnit, expected);
  if (!result) {
    return {
      kind: "error",
      error: "対象キーワードに一致する行動記録が見つかりませんでした。",
    };
  }

  return {
    kind: "success",
    result,
    message:
      "反復推定を更新しました。対象件数=" +
      result.matchedActionCount +
      " / バケット数=" +
      result.bucketCount,
  };
};
