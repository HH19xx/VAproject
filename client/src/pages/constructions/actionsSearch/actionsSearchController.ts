import type { SortKey, SortOrder } from "./actionsSearchTypes";
import type { ParsedQuery } from "./actionsSearchQuery";

export interface ActionSearchOptions {
  sort?: "occurred_at" | "created_at" | "updated_at" | "title" | "tag_count";
  order?: "asc" | "desc";
  from?: string;
  to?: string;
  anyTagIDs?: number[];
  anyTagGroups?: number[][];
  excludeTagIDs?: number[];
}

export const buildParsedSearchState = (
  danbooruQuery: string,
  tagNameToID: Map<string, number>,
  parseQuery: (query: string, tagNameToID: Map<string, number>) => ParsedQuery
): { parsed: ParsedQuery | null; error: string | null } => {
  const parsed = parseQuery(danbooruQuery, tagNameToID);
  if (parsed.unknownTokens.length > 0) {
    return {
      parsed: null,
      error: `未知のタグがあります: ${parsed.unknownTokens.join(", ")}`,
    };
  }
  return { parsed, error: null };
};

export const buildMergedAndTagIDs = (filterTagIDs: number[], parsed: ParsedQuery): number[] =>
  Array.from(new Set([...filterTagIDs, ...parsed.andTagIDs]));

export const buildSearchRange = (from: string, to: string): { from?: string; to?: string } => ({
  from: from ? new Date(from).toISOString() : undefined,
  to: to ? new Date(to).toISOString() : undefined,
});

export const buildActionSearchOptions = (
  sort: SortKey,
  order: SortOrder,
  parsed: ParsedQuery,
  range: { from?: string; to?: string }
): ActionSearchOptions => ({
  sort,
  order,
  from: range.from,
  to: range.to,
  anyTagIDs: parsed.anyTagIDs,
  anyTagGroups: parsed.anyTagGroups,
  excludeTagIDs: parsed.excludeTagIDs,
});

export const appendTagToQueryText = (
  danbooruQuery: string,
  tag: string
): { updatedQuery: string; alreadyIncluded: boolean } => {
  const comparableTag = tag.trim().toLowerCase();
  const tokens = danbooruQuery
    .split(/\s+/)
    .filter((token) => token.trim() !== "|")
    .map((token) => token.trim().toLowerCase().replace(/^[~-]/, ""))
    .filter(Boolean);
  if (tokens.includes(comparableTag)) {
    return { updatedQuery: danbooruQuery.trim(), alreadyIncluded: true };
  }

  const match = danbooruQuery.match(/^(.*?)(\S*)$/s);
  const base = match?.[1] ?? "";
  const activeToken = match?.[2] ?? "";
  const prefix = activeToken.startsWith("~") ? "~" : activeToken.startsWith("-") ? "-" : "";
  const replacement = `${prefix}${tag}`;
  const updatedQuery = `${base}${replacement}`.trim();

  return {
    updatedQuery,
    alreadyIncluded: false,
  };
};

export const buildSavedSearchPayload = (
  danbooruQuery: string,
  locationKey: string
): { queryText: string; description: string } => {
  const normalizedLocationKey = locationKey.trim();
  const locationToken = normalizedLocationKey ? `location_key:${normalizedLocationKey}` : "";
  const queryText = [danbooruQuery.trim(), locationToken].filter(Boolean).join(" ");
  return {
    queryText,
    description: queryText || "保存済み検索条件",
  };
};

export const buildClearedSearchState = () => ({
  filterTagIDs: [] as number[],
  danbooruQuery: "",
  queryError: null as string | null,
  saveMessage: null as string | null,
  analysisMessage: null as string | null,
  analysisState: null,
  distributionResult: null,
  selectedDistributionBin: null,
  sort: "occurred_at" as SortKey,
  order: "desc" as SortOrder,
  isDateRangeManual: false,
  from: "",
  to: "",
});

export const confirmDeleteAction = (confirmFn: (message?: string) => boolean): boolean =>
  confirmFn("この行動記録を削除しますか。");
