export type HistorySortKey = "created_desc" | "created_asc" | "score_desc";

export const extractSnapshotMeta = (queryText: string): { axis?: string; dataset?: string; query: string } => {
  const axisMatch = queryText.match(/(?:^|\s)axis:([^\s]+)/);
  const datasetMatch = queryText.match(/(?:^|\s)dataset:([^\s]+)/);
  const cleaned = queryText
    .replace(/(?:^|\s)axis:[^\s]+/g, " ")
    .replace(/(?:^|\s)dataset:[^\s]+/g, " ")
    .replace(/\s+/g, " ")
    .trim();

  return {
    axis: axisMatch?.[1],
    dataset: datasetMatch?.[1],
    query: cleaned,
  };
};

export const toDatasetLabel = (dataset?: string): string => {
  if (dataset === "action_logs") return "行動記録";
  if (dataset === "world_signals") return "外部ビッグデータ";
  return dataset || "未指定";
};

export const formatNum = (value: number, digits = 3) => value.toFixed(digits);
