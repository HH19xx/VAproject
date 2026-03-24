import type { ActionLog, ActionLogListMeta } from "../../../../hooks/useActions";

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
