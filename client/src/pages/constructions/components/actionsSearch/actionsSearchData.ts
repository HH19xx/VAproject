import type { ActionLog, ActionLogListMeta } from "../../../../hooks/useActions";

export type SearchTagSuggestion = {
  id: number;
  name: string;
  source: "frequent" | "kana" | "alpha";
};

const ASCII_LEADING_RE = /^[A-Za-z0-9]/;
const SEARCH_TOKEN_SPLIT_RE = /\s+/;
const ACTIVE_TOKEN_RE = /^(.*?)(\S*)$/s;

const normalizeSearchText = (value: string) => value.trim().toLowerCase();

const isAsciiLeadingTag = (name: string) => ASCII_LEADING_RE.test(name.trim());
const isStructuralToken = (value: string) => value.trim() === "|";

const compareJa = (left: string, right: string) => left.localeCompare(right, "ja");
const toComparableTagToken = (value: string) => normalizeSearchText(value).replace(/^[~-]/, "");
const extractActiveToken = (query: string) => query.match(ACTIVE_TOKEN_RE)?.[2] ?? "";

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
  tags.forEach((tag) => map.set(tag.name.trim().toLowerCase().replace(/\s+/g, "_"), tag.id));
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

export const buildSearchTagSuggestions = (
  tags: Array<{ id: number; name: string }>,
  query: string,
  selectedTagIDs: number[],
  limit: number
): SearchTagSuggestion[] => {
  const selected = new Set(selectedTagIDs);
  const activeToken = extractActiveToken(query);
  const normalizedQuery = toComparableTagToken(activeToken);
  const usedTokens = new Set(
    query
      .split(SEARCH_TOKEN_SPLIT_RE)
      .filter((token) => !isStructuralToken(token))
      .map((token) => toComparableTagToken(token))
      .filter(Boolean)
  );
  if (normalizedQuery) {
    usedTokens.delete(normalizedQuery);
  }
  const availableTags = tags.filter((tag) => !selected.has(tag.id));

  const pushUnique = (
    target: SearchTagSuggestion[],
    seen: Set<number>,
    candidates: Array<{ id: number; name: string }>,
    source: SearchTagSuggestion["source"]
  ) => {
    for (const candidate of candidates) {
      if (target.length >= limit) break;
      if (seen.has(candidate.id)) continue;
      seen.add(candidate.id);
      target.push({ ...candidate, source });
    }
  };

  if (!normalizedQuery) {
    const kana = [...availableTags]
      .filter((tag) => !usedTokens.has(toComparableTagToken(tag.name)))
      .filter((tag) => !isAsciiLeadingTag(tag.name))
      .sort((left, right) => compareJa(left.name, right.name));

    const alpha = [...availableTags]
      .filter((tag) => !usedTokens.has(toComparableTagToken(tag.name)))
      .filter((tag) => isAsciiLeadingTag(tag.name))
      .sort((left, right) => compareJa(left.name, right.name));

    const suggestions: SearchTagSuggestion[] = [];
    const seen = new Set<number>();
    pushUnique(suggestions, seen, kana, "kana");
    pushUnique(suggestions, seen, alpha, "alpha");
    return suggestions.slice(0, limit);
  }

  const matched = availableTags.filter((tag) => {
    const comparable = toComparableTagToken(tag.name);
    return !usedTokens.has(comparable) && comparable.includes(normalizedQuery);
  });
  const kanaMatched = matched
    .filter((tag) => !isAsciiLeadingTag(tag.name))
    .sort((left, right) => compareJa(left.name, right.name));
  const alphaMatched = matched
    .filter((tag) => isAsciiLeadingTag(tag.name))
    .sort((left, right) => compareJa(left.name, right.name));

  const suggestions: SearchTagSuggestion[] = [];
  const seen = new Set<number>();
  const preferAlpha = ASCII_LEADING_RE.test(activeToken.trim().replace(/^[~-]/, ""));
  if (preferAlpha) {
    pushUnique(suggestions, seen, alphaMatched, "alpha");
    pushUnique(suggestions, seen, kanaMatched, "kana");
  } else {
    pushUnique(suggestions, seen, kanaMatched, "kana");
    pushUnique(suggestions, seen, alphaMatched, "alpha");
  }
  return suggestions.slice(0, limit);
};
