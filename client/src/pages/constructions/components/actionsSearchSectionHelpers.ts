import type { ActionLog, ActionLogListMeta } from "../../../hooks/useActions";

export type SortKey = "occurred_at" | "created_at" | "updated_at" | "title" | "tag_count";
export type SortOrder = "asc" | "desc";

export interface ParsedQuery {
  andTagIDs: number[];
  anyTagIDs: number[];
  anyTagGroups: number[][];
  excludeTagIDs: number[];
  unknownTokens: string[];
}

export const normalizeTagToken = (token: string): string =>
  token.trim().toLowerCase().replace(/\s+/g, "_");

export const parseDanbooruStyleQuery = (query: string, tagNameToID: Map<string, number>): ParsedQuery => {
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
    for (const rawToken of rawTokens) {
      let mode: "and" | "any" | "exclude" = "and";
      let body = rawToken;

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
        unknownTokens.push(rawToken);
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
  let currentGroup: string[] = [];
  for (const token of rawTokens) {
    if (token === "|") {
      if (currentGroup.length > 0) groups.push(currentGroup);
      currentGroup = [];
      continue;
    }
    currentGroup.push(token);
  }
  if (currentGroup.length > 0) groups.push(currentGroup);

  for (const group of groups) {
    const groupAnd: number[] = [];
    const seenGroupAnd = new Set<number>();

    for (const rawToken of group) {
      let body = rawToken;

      if (body.startsWith("-")) {
        body = body.slice(1);
        if (!body) continue;
        const tagID = tagNameToID.get(normalizeTagToken(body));
        if (!tagID) {
          unknownTokens.push(rawToken);
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
          unknownTokens.push(rawToken);
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
        unknownTokens.push(rawToken);
        continue;
      }
      if (!seenGroupAnd.has(tagID)) {
        seenGroupAnd.add(tagID);
        groupAnd.push(tagID);
      }
    }

    if (groupAnd.length > 0) {
      anyTagGroups.push(groupAnd);
    }
  }

  return { andTagIDs, anyTagIDs, anyTagGroups, excludeTagIDs, unknownTokens };
};

export const toLocalDateTimeValue = (value: Date): string => {
  const local = new Date(value.getTime() - value.getTimezoneOffset() * 60 * 1000);
  return local.toISOString().slice(0, 16);
};

export const isValidDateRange = (from: string, to: string): boolean => {
  if (!from || !to) return false;
  const fromDate = new Date(from);
  const toDate = new Date(to);
  return Number.isFinite(fromDate.getTime()) && Number.isFinite(toDate.getTime()) && fromDate < toDate;
};

const OPEN_DATA_DAY_MS = 24 * 60 * 60 * 1000;

export const syncDateRangeWithOpenDataWindow = (
  pastDaysValue: string,
  forecastDaysValue: string
): { from: string; to: string } | null => {
  const past = Number(pastDaysValue);
  const forecast = Number(forecastDaysValue);
  if (!Number.isFinite(past) || !Number.isFinite(forecast)) return null;

  const now = new Date();
  const from = new Date(now.getTime() - past * OPEN_DATA_DAY_MS);
  const to = new Date(now.getTime() + forecast * OPEN_DATA_DAY_MS);
  return {
    from: toLocalDateTimeValue(from),
    to: toLocalDateTimeValue(to),
  };
};

export const buildGroupedTags = (tags: Array<{ id: number; name: string; group_name?: string }>) => {
  const groups = new Map<string, Array<{ id: number; name: string }>>();
  tags.forEach((tag) => {
    const key = tag.group_name || "未分類";
    if (!groups.has(key)) {
      groups.set(key, []);
    }
    groups.get(key)?.push({ id: tag.id, name: tag.name });
  });
  return Array.from(groups.entries());
};

export const buildActionAxisCandidates = (meta: ActionLogListMeta | undefined, defaults: string[]) => {
  const seen = new Set<string>();
  const merged: string[] = [];
  for (const axis of [...defaults, ...(meta?.axis_candidates || [])]) {
    const key = axis.trim();
    if (!key || seen.has(key)) continue;
    seen.add(key);
    merged.push(key);
  }
  return merged;
};

export const buildTagNameMap = (tags: Array<{ id: number; name: string }>) => {
  const map = new Map<number, string>();
  tags.forEach((tag) => map.set(tag.id, tag.name));
  return map;
};

export const buildTagNameToID = (tags: Array<{ id: number; name: string }>) => {
  const map = new Map<string, number>();
  tags.forEach((tag) => map.set(normalizeTagToken(tag.name), tag.id));
  return map;
};

export const buildPrototypeNameMap = (prototypes: Array<{ id: number; name: string }>) => {
  const map = new Map<number, string>();
  prototypes.forEach((prototype) => map.set(prototype.id, prototype.name));
  return map;
};

export const formatTagNames = (tagIDs: number[], tagNameMap: Map<number, string>) =>
  tagIDs.map((id) => tagNameMap.get(id) || `#${id}`).join(", ");

export const buildActionScatterPoints = (
  actions: ActionLog[],
  xAxis: string,
  yAxis: string,
  axisValueResolver: (action: ActionLog, axis: string) => number | null
) =>
  actions
    .map((action) => {
      const x = axisValueResolver(action, xAxis);
      const y = axisValueResolver(action, yAxis);
      if (x === null || y === null) return null;
      return { id: action.id, title: action.title, x, y };
    })
    .filter((point): point is { id: number; title: string; x: number; y: number } => point !== null);

export const isSameDateTimeByMinute = (left: string, right: string): boolean => {
  const leftTime = new Date(left).getTime();
  const rightTime = new Date(right).getTime();
  if (!Number.isFinite(leftTime) || !Number.isFinite(rightTime)) return false;
  return Math.floor(leftTime / 60000) === Math.floor(rightTime / 60000);
};
