import type { ActionLog } from "../../../hooks/useActions";

export const actionAxisValue = (action: ActionLog, axis: string): number | null => {
  if (axis === "occurred_at") return new Date(action.occurred_at).getTime();
  if (axis === "created_at") return new Date(action.created_at).getTime();
  if (axis === "updated_at") return new Date(action.updated_at).getTime();
  if (axis === "tag_count") return action.tag_ids.length;
  if (axis === "prototype_id") return action.prototype_id ?? null;

  const attr = action.attributes?.find((item) => item.key === axis);
  return typeof attr?.value_number === "number" ? attr.value_number : null;
};
