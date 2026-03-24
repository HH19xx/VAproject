import type { ActionLog } from "../../../../hooks/useActions";
import type { DistributionDataset } from "../../../../hooks/useDistributionAnalysis";

export const ACTION_AXIS_DEFAULTS = [
  "occurred_at",
  "created_at",
  "updated_at",
  "tag_count",
  "prototype_id",
] as const;

export const actionAxisValue = (action: ActionLog, axis: string): number | null => {
  if (axis === "occurred_at") return new Date(action.occurred_at).getTime();
  if (axis === "created_at") return new Date(action.created_at).getTime();
  if (axis === "updated_at") return new Date(action.updated_at).getTime();
  if (axis === "tag_count") return action.tag_ids.length;
  if (axis === "prototype_id") return action.prototype_id ?? null;

  const attr = action.attributes?.find((item) => item.key === axis);
  return typeof attr?.value_number === "number" ? attr.value_number : null;
};

export const defaultAxisByDataset = (dataset: DistributionDataset, actionAxisCandidates: string[]): string => {
  if (dataset === "world_signals") return "temperature_c";
  return actionAxisCandidates.includes("tag_count") ? "tag_count" : actionAxisCandidates[0] || "tag_count";
};
