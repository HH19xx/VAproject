import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useActions, type ActionLog } from "../hooks/useActions";
import { usePrototypes } from "../hooks/usePrototypes";
import { useTags } from "../hooks/useTags";
import { useTargets } from "../hooks/useTargets";
import {
  useAnalysisSnapshots,
  type AnalysisSeverity,
  type AnalysisSnapshot,
  type DistributionStats,
} from "../hooks/useAnalysisSnapshots";
import { useWorldSignals, type AnalysisContextResult } from "../hooks/useWorldSignals";
import styles from "../assets/styles/Actions.module.scss";

type SortKey = "occurred_at" | "created_at" | "updated_at" | "title" | "tag_count";
type SortOrder = "asc" | "desc";

interface ParsedQuery {
  andTagIDs: number[];
  anyTagIDs: number[];
  anyTagGroups: number[][];
  excludeTagIDs: number[];
  unknownTokens: string[];
}

interface AnalysisViewState {
  severity: AnalysisSeverity;
  score: number;
  pValue?: number;
  significant?: boolean;
  deltaAvgTag: number;
  deltaVarTag: number;
  deltaPrototypeRate: number;
  current: DistributionStats;
  baseline?: DistributionStats;
}

type ContextVizMode = "time" | "scatter";
type ScatterAxisKey = "temperature_c" | "precipitation_mm" | "wind_speed_ms" | "weather_code";
type HistorySortKey = "created_desc" | "created_asc" | "score_desc";
type BuiltInActionAxisKey = "occurred_at" | "created_at" | "updated_at" | "tag_count" | "prototype_id";

const normalizeTagToken = (token: string): string => token.trim().toLowerCase().replace(/\s+/g, "_");

const formatNum = (value: number, digits = 3) => value.toFixed(digits);

const isValidDateRange = (from: string, to: string): boolean => {
  if (!from || !to) return false;
  const fromDate = new Date(from);
  const toDate = new Date(to);
  return Number.isFinite(fromDate.getTime()) && Number.isFinite(toDate.getTime()) && fromDate < toDate;
};

const parseDanbooruStyleQuery = (query: string, tagNameToID: Map<string, number>): ParsedQuery => {
  const andTagIDs: number[] = [];
  const anyTagIDs: number[] = [];
  const anyTagGroups: number[][] = [];
  const excludeTagIDs: number[] = [];
  const unknownTokens: string[] = [];

  const seenAnd = new Set<number>();
  const seenAny = new Set<number>();
  const seenExclude = new Set<number>();

  const rawTokens = query
    .split(/\s+/)
    .map((token) => token.trim())
    .filter((token) => token.length > 0);

  if (!rawTokens.includes("|")) {
    for (const tokenRaw of rawTokens) {
      let mode: "and" | "any" | "exclude" = "and";
      let body = tokenRaw;

      if (body.startsWith("-")) {
        mode = "exclude";
        body = body.slice(1);
      } else if (body.startsWith("~")) {
        mode = "any";
        body = body.slice(1);
      }

      if (!body) continue;

      const tagID = tagNameToID.get(normalizeTagToken(body));
      if (!tagID) {
        unknownTokens.push(tokenRaw);
        continue;
      }

      if (mode === "and" && !seenAnd.has(tagID)) {
        seenAnd.add(tagID);
        andTagIDs.push(tagID);
      }
      if (mode === "any" && !seenAny.has(tagID)) {
        seenAny.add(tagID);
        anyTagIDs.push(tagID);
      }
      if (mode === "exclude" && !seenExclude.has(tagID)) {
        seenExclude.add(tagID);
        excludeTagIDs.push(tagID);
      }
    }
    return { andTagIDs, anyTagIDs, anyTagGroups, excludeTagIDs, unknownTokens };
  }

  const groups: string[][] = [];
  let current: string[] = [];
  for (const token of rawTokens) {
    if (token === "|") {
      if (current.length > 0) groups.push(current);
      current = [];
      continue;
    }
    current.push(token);
  }
  if (current.length > 0) groups.push(current);

  for (const group of groups) {
    const groupAnd: number[] = [];
    const seenGroupAnd = new Set<number>();
    for (const tokenRaw of group) {
      let body = tokenRaw;
      if (body.startsWith("-")) {
        body = body.slice(1);
        if (!body) continue;
        const tagID = tagNameToID.get(normalizeTagToken(body));
        if (!tagID) {
          unknownTokens.push(tokenRaw);
          continue;
        }
        if (!seenExclude.has(tagID)) {
          seenExclude.add(tagID);
          excludeTagIDs.push(tagID);
        }
        continue;
      }

      if (body.startsWith("~")) {
        body = body.slice(1);
        if (!body) continue;
        const tagID = tagNameToID.get(normalizeTagToken(body));
        if (!tagID) {
          unknownTokens.push(tokenRaw);
          continue;
        }
        if (!seenAny.has(tagID)) {
          seenAny.add(tagID);
          anyTagIDs.push(tagID);
        }
        continue;
      }

      if (body.startsWith("+")) {
        body = body.slice(1);
      }
      if (!body) continue;
      const tagID = tagNameToID.get(normalizeTagToken(body));
      if (!tagID) {
        unknownTokens.push(tokenRaw);
        continue;
      }
      if (!seenGroupAnd.has(tagID)) {
        seenGroupAnd.add(tagID);
        groupAnd.push(tagID);
      }
    }
    if (groupAnd.length > 0) anyTagGroups.push(groupAnd);
  }

  return { andTagIDs, anyTagIDs, anyTagGroups, excludeTagIDs, unknownTokens };
};

const calculateDistributionStats = (logs: ActionLog[]): DistributionStats => {
  if (logs.length === 0) {
    return {
      count: 0,
      avg_tag_count: 0,
      var_tag_count: 0,
      prototype_rate: 0,
    };
  }

  const tagCounts = logs.map((log) => log.tag_ids.length);
  const mean = tagCounts.reduce((sum, value) => sum + value, 0) / tagCounts.length;
  const variance = tagCounts.reduce((sum, value) => sum + (value - mean) ** 2, 0) / tagCounts.length;
  const prototypeCount = logs.filter((log) => Boolean(log.prototype_id)).length;

  return {
    count: logs.length,
    avg_tag_count: mean,
    var_tag_count: variance,
    prototype_rate: prototypeCount / logs.length,
  };
};

const computeStdDev = (variance: number): number => Math.sqrt(Math.max(0, variance));

const approxNormalCDF = (x: number): number => {
  const absX = Math.abs(x);
  const t = 1 / (1 + 0.2316419 * absX);
  const d = 0.3989423 * Math.exp((-absX * absX) / 2);
  const prob =
    1 -
    d *
      t *
      (0.3193815 + t * (-0.3565638 + t * (1.781478 + t * (-1.821256 + t * 1.330274))));
  return x >= 0 ? prob : 1 - prob;
};

const estimatePValueByMeanDiff = (current: DistributionStats, baseline?: DistributionStats): number | undefined => {
  if (!baseline) return undefined;
  if (current.count < 2 || baseline.count < 2) return undefined;

  const stdCurrent = computeStdDev(current.var_tag_count);
  const stdBaseline = computeStdDev(baseline.var_tag_count);
  const se = Math.sqrt((stdCurrent ** 2) / current.count + (stdBaseline ** 2) / baseline.count);
  if (se <= 0) return undefined;

  const z = Math.abs(current.avg_tag_count - baseline.avg_tag_count) / se;
  return Math.max(0, Math.min(1, 2 * (1 - approxNormalCDF(z))));
};

const deriveSeverity = (
  current: DistributionStats,
  baseline?: DistributionStats
): Omit<AnalysisViewState, "current" | "baseline"> => {
  const base = baseline ?? {
    count: 0,
    avg_tag_count: 0,
    var_tag_count: 0,
    prototype_rate: 0,
  };

  const deltaAvgTag = current.avg_tag_count - base.avg_tag_count;
  const deltaVarTag = current.var_tag_count - base.var_tag_count;
  const deltaPrototypeRate = current.prototype_rate - base.prototype_rate;
  const score = Math.abs(deltaAvgTag) * 0.5 + Math.abs(deltaVarTag) * 0.3 + Math.abs(deltaPrototypeRate) * 100 * 0.2;
  const pValue = estimatePValueByMeanDiff(current, baseline);
  const significant = pValue !== undefined ? pValue < 0.05 : undefined;

  let severity: AnalysisSeverity = "OK";
  if (score >= 12 || significant === true) {
    severity = "ALERT";
  } else if (score >= 5) {
    severity = "NOTICE";
  }

  return {
    severity,
    score,
    pValue,
    significant,
    deltaAvgTag,
    deltaVarTag,
    deltaPrototypeRate,
  };
};

const toBadgeClass = (severity: AnalysisSeverity): string => {
  if (severity === "ALERT") return styles.badgeAlert;
  if (severity === "NOTICE") return styles.badgeNotice;
  return styles.badgeOk;
};

const toLocalDateTimeValue = (value: Date): string => {
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  const hour = String(value.getHours()).padStart(2, "0");
  const min = String(value.getMinutes()).padStart(2, "0");
  return `${year}-${month}-${day}T${hour}:${min}`;
};

const axisLabelMap: Record<ScatterAxisKey, string> = {
  temperature_c: "気温 (C)",
  precipitation_mm: "降水量 (mm)",
  wind_speed_ms: "風速 (m/s)",
  weather_code: "天気コード",
};

const worldSignalAxisValue = (signal: AnalysisContextResult["world_signals"][number], axis: ScatterAxisKey): number | null => {
  const value = signal[axis];
  return typeof value === "number" ? value : null;
};

const ACTION_AXIS_DEFAULTS: BuiltInActionAxisKey[] = [
  "occurred_at",
  "created_at",
  "updated_at",
  "tag_count",
  "prototype_id",
];

const toActionAxisLabel = (axis: string): string => {
  if (axis === "occurred_at") return "発生日時";
  if (axis === "created_at") return "作成日時";
  if (axis === "updated_at") return "更新日時";
  if (axis === "tag_count") return "タグ数";
  if (axis === "prototype_id") return "プロトタイプID";
  return axis;
};

const actionAxisValue = (action: ActionLog, axis: string): number | null => {
  if (axis === "occurred_at") {
    const ts = new Date(action.occurred_at).getTime();
    return Number.isFinite(ts) ? ts : null;
  }
  if (axis === "created_at") {
    const ts = new Date(action.created_at).getTime();
    return Number.isFinite(ts) ? ts : null;
  }
  if (axis === "updated_at") {
    const ts = new Date(action.updated_at).getTime();
    return Number.isFinite(ts) ? ts : null;
  }
  if (axis === "tag_count") return action.tag_ids.length;
  if (axis === "prototype_id") return action.prototype_id ?? 0;

  const attribute = action.attributes?.find((item) => item.key === axis);
  if (!attribute) return null;
  return typeof attribute.value_number === "number" && Number.isFinite(attribute.value_number)
    ? attribute.value_number
    : null;
};

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

  const [sort, setSort] = useState<SortKey>("occurred_at");
  const [order, setOrder] = useState<SortOrder>("desc");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [filterTagIDs, setFilterTagIDs] = useState<number[]>([]);
  const [danbooruQuery, setDanbooruQuery] = useState("");
  const [queryError, setQueryError] = useState<string | null>(null);
  const [saveMessage, setSaveMessage] = useState<string | null>(null);
  const [analysisState, setAnalysisState] = useState<AnalysisViewState | null>(null);
  const [analysisHistory, setAnalysisHistory] = useState<AnalysisSnapshot[]>([]);
  const [analysisMessage, setAnalysisMessage] = useState<string | null>(null);
  const [locationKey, setLocationKey] = useState("tokyo_shinjuku");
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

  useEffect(() => {
    fetchActions();
    fetchTags();
    fetchPrototypes();
    void loadAnalysisHistory();
    const now = new Date();
    const past = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
    setFrom(toLocalDateTimeValue(past));
    setTo(toLocalDateTimeValue(now));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const loadAnalysisHistory = async () => {
    const snapshots = await listSnapshots(100);
    setAnalysisHistory(snapshots);
  };

  const tagNameMap = useMemo(() => {
    const map = new Map<number, string>();
    tags.forEach((tag) => map.set(tag.id, tag.name));
    return map;
  }, [tags]);

  const tagNameToID = useMemo(() => {
    const map = new Map<string, number>();
    tags.forEach((tag) => map.set(normalizeTagToken(tag.name), tag.id));
    return map;
  }, [tags]);

  const prototypeNameMap = useMemo(() => {
    const map = new Map<number, string>();
    prototypes.forEach((prototype) => map.set(prototype.id, prototype.name));
    return map;
  }, [prototypes]);

  const groupedTags = useMemo(() => {
    const groups = new Map<string, typeof tags>();
    tags.forEach((tag) => {
      const key = tag.group_name || "未分類";
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key)?.push(tag);
    });
    return Array.from(groups.entries());
  }, [tags]);

  const actionAxisCandidates = useMemo(() => {
    const seen = new Set<string>();
    const merged: string[] = [];
    for (const axis of [...ACTION_AXIS_DEFAULTS, ...(meta.axis_candidates || [])]) {
      const key = axis.trim();
      if (!key || seen.has(key)) continue;
      seen.add(key);
      merged.push(key);
    }
    return merged;
  }, [meta.axis_candidates]);

  useEffect(() => {
    if (actionAxisCandidates.length === 0) return;
    if (!actionAxisCandidates.includes(actionScatterXAxis)) {
      setActionScatterXAxis(actionAxisCandidates[0]);
    }
    if (!actionAxisCandidates.includes(actionScatterYAxis)) {
      setActionScatterYAxis(actionAxisCandidates[1] || actionAxisCandidates[0]);
    }
  }, [actionAxisCandidates, actionScatterXAxis, actionScatterYAxis]);

  const toggleFilterTag = (tagID: number) => {
    setFilterTagIDs((prev) => (prev.includes(tagID) ? prev.filter((id) => id !== tagID) : [...prev, tagID]));
  };

  const resolveBaselineWindow = (): { from?: string; to?: string } => {
    if (isValidDateRange(from, to)) {
      const fromDate = new Date(from);
      const toDate = new Date(to);
      const duration = toDate.getTime() - fromDate.getTime();
      const baselineTo = fromDate;
      const baselineFrom = new Date(fromDate.getTime() - duration);
      return {
        from: baselineFrom.toISOString(),
        to: baselineTo.toISOString(),
      };
    }
    return {};
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

    const nextState: AnalysisViewState = {
      ...summary,
      current: currentStats,
      baseline: baselineStats,
    };
    setAnalysisState(nextState);

    const saved = await createSnapshot({
      query_text: danbooruQuery,
      severity: summary.severity,
      score: Math.round(summary.score),
      delta_avg_tag: summary.deltaAvgTag,
      delta_var_tag: summary.deltaVarTag,
      delta_prototype_rate: summary.deltaPrototypeRate,
      p_value: summary.pValue,
      significant: summary.significant,
      current: currentStats,
      baseline: baselineStats,
    });
    if (saved) {
      setAnalysisHistory((prev) => [saved, ...prev].slice(0, 100));
      setAnalysisMessage("現状分析を保存しました。");
    }
  };

  const handleSearch = async () => {
    const parsed = parseDanbooruStyleQuery(danbooruQuery, tagNameToID);
    if (parsed.unknownTokens.length > 0) {
      setQueryError(`未知のタグがあります: ${parsed.unknownTokens.join(", ")}`);
      return;
    }
    setQueryError(null);
    setSaveMessage(null);
    setAnalysisMessage(null);

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
  };

  const handleClear = async () => {
    setFilterTagIDs([]);
    setDanbooruQuery("");
    setQueryError(null);
    setSaveMessage(null);
    setAnalysisMessage(null);
    setAnalysisState(null);
    setSort("occurred_at");
    setOrder("desc");
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
      setSaveMessage("保存対象にするには、タグ条件を1つ以上指定してください。");
      return;
    }

    const suggestedName = `saved-search-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, "-")}`;
    const name = window.prompt("検索条件の名前", suggestedName)?.trim();
    if (!name) return;

    const normalizedLocationKey = locationKey.trim();
    const locationToken = normalizedLocationKey ? `location_key:${normalizedLocationKey}` : "";
    const queryTextToSave = [danbooruQuery.trim(), locationToken].filter(Boolean).join(" ");
    const descriptionToSave = queryTextToSave || "行動記録検索条件";

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
    setSaveMessage(`検索条件を保存しました: ${created.name} (#${created.id})`);
  };

  const handleClearSnapshots = async () => {
    if (!window.confirm("分析履歴をすべて削除します。よろしいですか？")) return;
    const ok = await clearSnapshots();
    if (ok) {
      setAnalysisHistory([]);
      setAnalysisMessage("分析履歴をクリアしました。");
    }
  };

  const handleFetchWorldSignals = async () => {
    const lat = Number(latitude);
    const lon = Number(longitude);
    const pDays = Number(pastDays);
    const fDays = Number(forecastDays);
    if (!locationKey.trim()) {
      setAnalysisMessage("location_key を入力してください。");
      return;
    }
    if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
      setAnalysisMessage("緯度・経度は数値で入力してください。");
      return;
    }
    if (!Number.isInteger(pDays) || !Number.isInteger(fDays)) {
      setAnalysisMessage("past_days / forecast_days は整数で入力してください。");
      return;
    }

    const ok = await fetchOpenMeteo({
      locationKey: locationKey.trim(),
      latitude: lat,
      longitude: lon,
      pastDays: pDays,
      forecastDays: fDays,
    });
    if (ok) {
      setAnalysisMessage("Open-Meteo データを取得して保存しました。");
    }
  };

  const handleLoadAnalysisContext = async () => {
    if (!from || !to) {
      setAnalysisMessage("分析期間（from/to）を指定してください。");
      return;
    }
    const fromISO = new Date(from).toISOString();
    const toISO = new Date(to).toISOString();
    const result = await fetchAnalysisContext(locationKey.trim(), fromISO, toISO, 200);
    if (result) {
      setContextResult(result);
      setAnalysisMessage("行動記録とオープンデータの統合コンテキストを取得しました。");
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm("この行動記録を削除しますか？")) return;
    await deleteAction(id);
  };

  const formatTagNames = (tagIDs: number[]) => tagIDs.map((id) => tagNameMap.get(id) || `#${id}`).join(", ");

  const contextTimePoints = useMemo(() => {
    if (!contextResult || contextResult.world_signals.length === 0) return [];
    return contextResult.world_signals
      .filter((signal) => typeof signal.temperature_c === "number")
      .map((signal) => ({
        t: new Date(signal.observed_at).getTime(),
        temperature: signal.temperature_c as number,
      }));
  }, [contextResult]);

  const contextScatterPoints = useMemo(() => {
    if (!contextResult || contextResult.world_signals.length === 0) return [];
    return contextResult.world_signals
      .filter((signal) => worldSignalAxisValue(signal, scatterXAxis) !== null && worldSignalAxisValue(signal, scatterYAxis) !== null)
      .map((signal) => ({
        x: worldSignalAxisValue(signal, scatterXAxis) as number,
        y: worldSignalAxisValue(signal, scatterYAxis) as number,
      }));
  }, [contextResult, scatterXAxis, scatterYAxis]);

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

  const actionScatterPoints = useMemo(() => {
    if (actions.length === 0) return [];
    return actions
      .map((action) => {
        const x = actionAxisValue(action, actionScatterXAxis);
        const y = actionAxisValue(action, actionScatterYAxis);
        if (x === null || y === null) return null;
        return { id: action.id, title: action.title, x, y };
      })
      .filter((point): point is { id: number; title: string; x: number; y: number } => point !== null);
  }, [actions, actionScatterXAxis, actionScatterYAxis]);

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>行動記録</h1>
        <button className={styles.primaryButton} onClick={() => navigate("/actions/new")}>
          新規行動記録
        </button>
      </div>

      {error && <div className={styles.errorBox}>エラー: {error}</div>}
      {queryError && <div className={styles.errorBox}>検索条件エラー: {queryError}</div>}
      {snapshotError && <div className={styles.errorBox}>分析履歴エラー: {snapshotError}</div>}
      {saveMessage && <div className={styles.info}>{saveMessage}</div>}
      {analysisMessage && <div className={styles.info}>{analysisMessage}</div>}

      <div className={styles.filterPanel}>
        <div className={styles.filterTitle}>検索条件</div>
        <div className={styles.controlGrid}>
          <label>
            並び順
            <select className={styles.select} value={sort} onChange={(e) => setSort(e.target.value as SortKey)}>
              <option value="occurred_at">occurred_at</option>
              <option value="created_at">created_at</option>
              <option value="updated_at">updated_at</option>
              <option value="title">title</option>
              <option value="tag_count">tag_count</option>
            </select>
          </label>
          <label>
            順序
            <select className={styles.select} value={order} onChange={(e) => setOrder(e.target.value as SortOrder)}>
              <option value="desc">desc</option>
              <option value="asc">asc</option>
            </select>
          </label>
          <label>
            期間開始
            <input className={styles.input} type="datetime-local" value={from} onChange={(e) => setFrom(e.target.value)} />
          </label>
          <label>
            期間終了
            <input className={styles.input} type="datetime-local" value={to} onChange={(e) => setTo(e.target.value)} />
          </label>
        </div>

        <div className={styles.controlGrid}>
          <label>
            Danbooru形式検索クエリ
            <input
              className={styles.input}
              type="text"
              value={danbooruQuery}
              onChange={(e) => setDanbooruQuery(e.target.value)}
              placeholder="例: 平日 通勤 ~混雑 -遅延 | 山手線 新宿"
            />
          </label>
        </div>

        <div className={styles.info}>
          記法: `tag`=AND, `~tag`=OR, `-tag`=NOT, `A B | C D`=(A AND B) OR (C AND D), `+tag`=Group内AND
        </div>

        <div className={styles.filterTagList}>
          {groupedTags.map(([groupName, groupTags]) => (
            <div key={groupName} className={styles.filterTagGroup}>
              <div className={styles.filterTagGroupTitle}>{groupName}</div>
              <div className={styles.filterTagItems}>
                {groupTags.map((tag) => (
                  <label key={tag.id} className={styles.filterTagItem}>
                    <input type="checkbox" checked={filterTagIDs.includes(tag.id)} onChange={() => toggleFilterTag(tag.id)} />
                    <span>{tag.name}</span>
                  </label>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className={styles.filterActions}>
          <button className={styles.secondaryButton} onClick={handleSearch} disabled={loading || snapshotLoading}>
            検索
          </button>
          <button className={styles.successButton} onClick={handleSaveAsTarget} disabled={loading || targetSaving}>
            検索条件として保存
          </button>
          <button className={styles.backButton} onClick={handleClear} disabled={loading}>
            クリア
          </button>
        </div>
      </div>

      <div className={styles.vizPanel}>
        <div className={styles.vizHeader}>
          <div className={styles.filterTitle}>オープンデータ連携（世界側データ）</div>
        </div>
        {worldSignalError && <div className={styles.errorBox}>オープンデータエラー: {worldSignalError}</div>}
        <div className={styles.controlGrid}>
          <label>
            location_key
            <input className={styles.input} value={locationKey} onChange={(e) => setLocationKey(e.target.value)} />
          </label>
          <label>
            latitude
            <input className={styles.input} value={latitude} onChange={(e) => setLatitude(e.target.value)} />
          </label>
          <label>
            longitude
            <input className={styles.input} value={longitude} onChange={(e) => setLongitude(e.target.value)} />
          </label>
          <label>
            past_days
            <input className={styles.input} value={pastDays} onChange={(e) => setPastDays(e.target.value)} />
          </label>
          <label>
            forecast_days
            <input className={styles.input} value={forecastDays} onChange={(e) => setForecastDays(e.target.value)} />
          </label>
        </div>
        <div className={styles.filterActions}>
          <button className={styles.secondaryButton} onClick={handleFetchWorldSignals} disabled={worldSignalLoading}>
            Open-Meteo取得
          </button>
          <button className={styles.successButton} onClick={handleLoadAnalysisContext} disabled={worldSignalLoading}>
            分析コンテキスト取得
          </button>
        </div>

        {contextResult && (
          <>
            <div className={styles.analysisPanel}>
              <div className={styles.analysisTitle}>世界側コンテキスト要約</div>
              <div className={styles.analysisRow}>action_count: {contextResult.summary.action_count}</div>
              <div className={styles.analysisRow}>signal_count: {contextResult.summary.signal_count}</div>
              <div className={styles.analysisRow}>
                avg_tags_per_action: {formatNum(contextResult.summary.avg_tags_per_action, 2)}
              </div>
              <div className={styles.analysisRow}>
                avg_temperature_c:{" "}
                {contextResult.summary.avg_temperature_c !== undefined
                  ? formatNum(contextResult.summary.avg_temperature_c, 2)
                  : "-"}
              </div>
              <div className={styles.analysisRow}>
                total_precipitation_mm: {formatNum(contextResult.summary.total_precipitation_mm, 2)}
              </div>
            </div>

            <div className={styles.vizModeSwitch}>
              <button
                className={contextVizMode === "time" ? styles.modeActive : styles.modeButton}
                onClick={() => setContextVizMode("time")}
              >
                1D 時系列
              </button>
              <button
                className={contextVizMode === "scatter" ? styles.modeActive : styles.modeButton}
                onClick={() => setContextVizMode("scatter")}
              >
                2D 散布図
              </button>
            </div>

            {contextVizMode === "time" ? (
              <div className={styles.scatterWrapper}>
                <svg className={styles.scatterSvg} viewBox="0 0 900 260" preserveAspectRatio="none">
                  {contextTimePoints.length > 1 &&
                    (() => {
                      const minT = Math.min(...contextTimePoints.map((p) => p.t));
                      const maxT = Math.max(...contextTimePoints.map((p) => p.t));
                      const minY = Math.min(...contextTimePoints.map((p) => p.temperature));
                      const maxY = Math.max(...contextTimePoints.map((p) => p.temperature));
                      const safeDx = maxT - minT || 1;
                      const safeDy = maxY - minY || 1;
                      return contextTimePoints.map((p, idx, arr) => {
                        if (idx === 0) return null;
                        const prev = arr[idx - 1];
                        const x1 = 40 + ((prev.t - minT) / safeDx) * 820;
                        const y1 = 220 - ((prev.temperature - minY) / safeDy) * 180;
                        const x2 = 40 + ((p.t - minT) / safeDx) * 820;
                        const y2 = 220 - ((p.temperature - minY) / safeDy) * 180;
                        return <line key={`${idx}-${p.t}`} x1={x1} y1={y1} x2={x2} y2={y2} stroke="#007bff" strokeWidth="2" />;
                      });
                    })()}
                  <text x="40" y="20" className={styles.axisLabel}>
                    時系列（気温）
                  </text>
                </svg>
              </div>
            ) : (
              <>
                <div className={styles.controlGrid}>
                  <label>
                    X軸
                    <select
                      className={styles.select}
                      value={scatterXAxis}
                      onChange={(e) => setScatterXAxis(e.target.value as ScatterAxisKey)}
                    >
                      {Object.entries(axisLabelMap).map(([axisKey, axisLabel]) => (
                        <option key={`x-${axisKey}`} value={axisKey}>
                          {axisLabel}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label>
                    Y軸
                    <select
                      className={styles.select}
                      value={scatterYAxis}
                      onChange={(e) => setScatterYAxis(e.target.value as ScatterAxisKey)}
                    >
                      {Object.entries(axisLabelMap).map(([axisKey, axisLabel]) => (
                        <option key={`y-${axisKey}`} value={axisKey}>
                          {axisLabel}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>

                <div className={styles.scatterWrapper}>
                  <svg className={styles.scatterSvg} viewBox="0 0 900 260" preserveAspectRatio="none">
                    {contextScatterPoints.length > 0 &&
                      (() => {
                        const minX = Math.min(...contextScatterPoints.map((p) => p.x));
                        const maxX = Math.max(...contextScatterPoints.map((p) => p.x));
                        const minY = Math.min(...contextScatterPoints.map((p) => p.y));
                        const maxY = Math.max(...contextScatterPoints.map((p) => p.y));
                        const safeDx = maxX - minX || 1;
                        const safeDy = maxY - minY || 1;
                        return contextScatterPoints.map((p, idx) => {
                          const x = 40 + ((p.x - minX) / safeDx) * 820;
                          const y = 220 - ((p.y - minY) / safeDy) * 180;
                          return <circle key={`${idx}-${p.x}-${p.y}`} className={styles.scatterPoint} cx={x} cy={y} r={4} />;
                        });
                      })()}
                    <text x="40" y="20" className={styles.axisLabel}>
                      {`散布図（x:${axisLabelMap[scatterXAxis]} / y:${axisLabelMap[scatterYAxis]}）`}
                    </text>
                  </svg>
                </div>
              </>
            )}
          </>
        )}
      </div>

      {analysisState && (
        <div className={styles.analysisPanel}>
          <div className={styles.analysisTitle}>
            現状分析
            <span className={toBadgeClass(analysisState.severity)}>{analysisState.severity}</span>
          </div>
          <div className={styles.analysisRow}>スコア: {formatNum(analysisState.score, 2)}</div>
          <div className={styles.analysisRow}>平均タグ数差分: {formatNum(analysisState.deltaAvgTag)}</div>
          <div className={styles.analysisRow}>タグ数分散差分: {formatNum(analysisState.deltaVarTag)}</div>
          <div className={styles.analysisRow}>
            プロトタイプ率差分: {formatNum(analysisState.deltaPrototypeRate * 100, 2)}%
          </div>
          <div className={styles.analysisRow}>
            現在分布: 件数={analysisState.current.count}, 平均タグ数={formatNum(analysisState.current.avg_tag_count)}, 分散=
            {formatNum(analysisState.current.var_tag_count)}, プロトタイプ率=
            {formatNum(analysisState.current.prototype_rate * 100, 1)}%
          </div>
          {analysisState.baseline && (
            <div className={styles.analysisRow}>
              基準分布: 件数={analysisState.baseline.count}, 平均タグ数={formatNum(analysisState.baseline.avg_tag_count)}, 分散=
              {formatNum(analysisState.baseline.var_tag_count)}, プロトタイプ率=
              {formatNum(analysisState.baseline.prototype_rate * 100, 1)}%
            </div>
          )}
          {analysisState.pValue !== undefined && (
            <div className={styles.analysisRow}>
              p値(平均タグ数差分): {analysisState.pValue.toExponential(3)} / 有意={String(analysisState.significant)}
            </div>
          )}
        </div>
      )}

      <div className={styles.trendPanel}>
        <div className={styles.analysisTitle}>推移履歴（分析スナップショット）</div>
        <div className={styles.controlGrid}>
          <label>
            重要度フィルタ
            <select
              className={styles.select}
              value={historySeverityFilter}
              onChange={(e) => setHistorySeverityFilter(e.target.value as AnalysisSeverity | "ALL")}
            >
              <option value="ALL">ALL</option>
              <option value="OK">OK</option>
              <option value="NOTICE">NOTICE</option>
              <option value="ALERT">ALERT</option>
            </select>
          </label>
          <label>
            並び順
            <select className={styles.select} value={historySort} onChange={(e) => setHistorySort(e.target.value as HistorySortKey)}>
              <option value="created_desc">新しい順</option>
              <option value="created_asc">古い順</option>
              <option value="score_desc">スコア順</option>
            </select>
          </label>
        </div>
        <svg className={styles.trendSvg} viewBox="0 0 600 120" preserveAspectRatio="none">
          {filteredSortedHistory.length > 1 &&
            filteredSortedHistory.slice(0, 40).map((entry, index, arr) => {
              if (index === 0) return null;
              const maxScore = Math.max(1, ...arr.map((v) => v.score));
              const x1 = ((arr.length - index) / (arr.length - 1)) * 590 + 5;
              const x2 = ((arr.length - (index - 1)) / (arr.length - 1)) * 590 + 5;
              const y1 = 110 - (entry.score / maxScore) * 100;
              const y2 = 110 - (arr[index - 1].score / maxScore) * 100;
              return <line key={entry.id} x1={x1} y1={y1} x2={x2} y2={y2} stroke="#007bff" strokeWidth="2" />;
            })}
        </svg>
        <div className={styles.analysisRow}>
          履歴件数: {filteredSortedHistory.length}件 / 全体 {analysisHistory.length}件（最新100件まで）
        </div>
        <div className={styles.analysisRow}>
          直近5件:
          {filteredSortedHistory
            .slice(0, 5)
            .map((item) => `${new Date(item.created_at).toLocaleString("ja-JP")} [${item.severity}] score=${formatNum(item.score, 2)}`)
            .join(" / ")}
        </div>
        <div className={styles.filterActions}>
          <button className={styles.dangerButton} onClick={handleClearSnapshots} disabled={snapshotLoading}>
            分析履歴を削除
          </button>
        </div>
      </div>

      <div className={styles.info}>検索結果: 全 {pagination.total} 件 / 現在表示 {actions.length} 件</div>

      <div className={styles.vizPanel}>
        <div className={styles.vizHeader}>
          <div className={styles.filterTitle}>検索結果の2D可視化</div>
        </div>
        <div className={styles.controlGrid}>
          <label>
            X軸
            <select
              className={styles.select}
              value={actionScatterXAxis}
              onChange={(e) => setActionScatterXAxis(e.target.value)}
            >
              {actionAxisCandidates.map((axis) => (
                <option key={`action-x-${axis}`} value={axis}>
                  {toActionAxisLabel(axis)}
                </option>
              ))}
            </select>
          </label>
          <label>
            Y軸
            <select
              className={styles.select}
              value={actionScatterYAxis}
              onChange={(e) => setActionScatterYAxis(e.target.value)}
            >
              {actionAxisCandidates.map((axis) => (
                <option key={`action-y-${axis}`} value={axis}>
                  {toActionAxisLabel(axis)}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className={styles.analysisRow}>
          描画点: {actionScatterPoints.length} 件（数値軸を持つレコードのみ）
        </div>
        <div className={styles.analysisRow}>
          軸キー: x={actionScatterXAxis} / y={actionScatterYAxis}
        </div>
        {actionScatterPoints.length === 0 && (
          <div className={styles.info}>選択した軸の組み合わせで数値を持つ記録がありません。軸を変更してください。</div>
        )}
        <div className={styles.scatterWrapper}>
          <svg className={styles.scatterSvg} viewBox="0 0 900 260" preserveAspectRatio="none">
            {actionScatterPoints.length > 0 &&
              (() => {
                const minX = Math.min(...actionScatterPoints.map((p) => p.x));
                const maxX = Math.max(...actionScatterPoints.map((p) => p.x));
                const minY = Math.min(...actionScatterPoints.map((p) => p.y));
                const maxY = Math.max(...actionScatterPoints.map((p) => p.y));
                const safeDx = maxX - minX || 1;
                const safeDy = maxY - minY || 1;
                return actionScatterPoints.map((point) => {
                  const x = 40 + ((point.x - minX) / safeDx) * 820;
                  const y = 220 - ((point.y - minY) / safeDy) * 180;
                  return (
                    <circle key={`action-point-${point.id}`} className={styles.scatterPoint} cx={x} cy={y} r={4}>
                      <title>{`${point.title} / x=${point.x.toFixed(2)} / y=${point.y.toFixed(2)}`}</title>
                    </circle>
                  );
                });
              })()}
            <text x="40" y="20" className={styles.axisLabel}>
              {`散布図（x:${toActionAxisLabel(actionScatterXAxis)} / y:${toActionAxisLabel(actionScatterYAxis)}）`}
            </text>
          </svg>
        </div>
      </div>

      {actions.length === 0 ? (
        <div className={styles.empty}>行動記録がありません。</div>
      ) : (
        <div className={styles.timeline}>
          {actions.map((action, index) => (
            <div key={action.id} className={styles.timelineItem}>
              <div className={styles.timelineIndex}>{index + 1}</div>
              <div className={styles.timelineBody}>
                <div className={styles.timelineHeader}>
                  <div className={styles.timelineTitle}>{action.title}</div>
                  <div className={styles.timelineTime}>{new Date(action.occurred_at).toLocaleString("ja-JP")}</div>
                </div>
                <div className={styles.timelineMeta}>タグ: {formatTagNames(action.tag_ids)}</div>
                <div className={styles.timelineMeta}>
                  プロトタイプ: {action.prototype_id ? prototypeNameMap.get(action.prototype_id) || `#${action.prototype_id}` : "-"}
                </div>
                {action.notes && <div className={styles.timelineNote}>{action.notes}</div>}
                <div className={styles.timelineActions}>
                  <button className={styles.successButton} onClick={() => navigate(`/actions/${action.id}/edit`)}>
                    編集
                  </button>
                  <button className={styles.dangerButton} onClick={() => handleDelete(action.id)} disabled={loading}>
                    削除
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <div className={styles.footer}>
        <button className={styles.backButton} onClick={() => navigate("/dashboard")}>
          ダッシュボードへ戻る
        </button>
      </div>
    </div>
  );
};

export default Actions;

