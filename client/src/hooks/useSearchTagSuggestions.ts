import { useMemo } from "react";
import { buildSearchTagSuggestions, type SearchTagSuggestion } from "../pages/constructions/actionsSearch/actionsSearchSectionHelpers";

const DEFAULT_SEARCH_SUGGESTION_LIMIT = 12;

const useSearchTagSuggestions = (
  tags: Array<{ id: number; name: string }>,
  query: string,
  selectedTagIDs: number[],
  limit = DEFAULT_SEARCH_SUGGESTION_LIMIT
): SearchTagSuggestion[] =>
  useMemo(
    () => buildSearchTagSuggestions(tags, query, selectedTagIDs, limit),
    [tags, query, selectedTagIDs, limit]
  );

export default useSearchTagSuggestions;
