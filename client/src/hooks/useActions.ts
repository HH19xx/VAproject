import { useContext, useState } from "react";
import { AuthContext } from "../contexts/AuthContext";

export interface ActionAttribute {
  key: string;
  value_number: number;
}

export interface ActionLog {
  id: number;
  user_id: number;
  title: string;
  prototype_id?: number;
  occurred_at: string;
  notes: string;
  tag_ids: number[];
  attributes?: ActionAttribute[];
  created_at: string;
  create_user: string;
  updated_at: string;
  update_user: string;
  deleted_at?: string;
}

interface Pagination {
  page: number;
  limit: number;
  total: number;
}

export interface ActionLogListMeta {
  from?: string;
  to?: string;
  sort: string;
  order: string;
  used_tag_ids: number[];
  used_any_tag_ids: number[];
  used_any_tag_groups?: number[][];
  used_exclude_tag_ids: number[];
  axis_candidates: string[];
}

interface FetchActionsOptions {
  sort?: "occurred_at" | "created_at" | "updated_at" | "title" | "tag_count";
  order?: "asc" | "desc";
  from?: string;
  to?: string;
  anyTagIDs?: number[];
  anyTagGroups?: number[][];
  excludeTagIDs?: number[];
}

interface ActionLogSnapshot {
  logs: ActionLog[];
  pagination: Pagination;
  meta?: ActionLogListMeta;
}

export const useActions = () => {
  const authContext = useContext(AuthContext);
  const [actions, setActions] = useState<ActionLog[]>([]);
  const [pagination, setPagination] = useState<Pagination>({ page: 1, limit: 20, total: 0 });
  const [meta, setMeta] = useState<ActionLogListMeta>({
    sort: "occurred_at",
    order: "desc",
    used_tag_ids: [],
    used_any_tag_ids: [],
    used_exclude_tag_ids: [],
    axis_candidates: ["occurred_at", "created_at", "updated_at", "tag_count", "prototype_id"],
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) {
    throw new Error("useActions は AuthProvider 内で使用してください");
  }

  const { token } = authContext;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

  const authFetch = async (url: string, options: RequestInit = {}) => {
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
      ...(options.headers as HeadersInit),
    };
    const response = await fetch(url, { ...options, headers });
    if (!response.ok) {
      const errBody = await response.json().catch(() => ({}));
      throw new Error(errBody?.error?.message || "APIリクエストに失敗しました");
    }
    return response.json();
  };

  const buildSearchParams = (page: number, limit: number, tagIDs: number[], options: FetchActionsOptions) => {
    const params = new URLSearchParams();
    params.set("page", String(page));
    params.set("limit", String(limit));
    if (tagIDs.length > 0) {
      params.set("tag_ids", tagIDs.join(","));
    }
    if (options.sort) {
      params.set("sort", options.sort);
    }
    if (options.order) {
      params.set("order", options.order);
    }
    if (options.from) {
      params.set("from", options.from);
    }
    if (options.to) {
      params.set("to", options.to);
    }
    if (options.anyTagIDs && options.anyTagIDs.length > 0) {
      params.set("any_tag_ids", options.anyTagIDs.join(","));
    }
    if (options.anyTagGroups && options.anyTagGroups.length > 0) {
      const encodedGroups = options.anyTagGroups.map((group) => group.join("+")).join(",");
      params.set("any_tag_groups", encodedGroups);
    }
    if (options.excludeTagIDs && options.excludeTagIDs.length > 0) {
      params.set("exclude_tag_ids", options.excludeTagIDs.join(","));
    }
    return params;
  };

  const fetchActions = async (page = 1, limit = 20, tagIDs: number[] = [], options: FetchActionsOptions = {}) => {
    setLoading(true);
    setError(null);
    try {
      const params = buildSearchParams(page, limit, tagIDs, options);

      const data = await authFetch(`${API_BASE_URL}/action_logs?${params.toString()}`);
      if (!data.success) {
        throw new Error(data.message || "行動記録の取得に失敗しました");
      }
      setActions(data.data.logs || []);
      setPagination(data.data.pagination);
      if (data.data.meta) {
        setMeta(data.data.meta as ActionLogListMeta);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の取得に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  const fetchActionsSnapshot = async (
    page = 1,
    limit = 200,
    tagIDs: number[] = [],
    options: FetchActionsOptions = {}
  ): Promise<ActionLogSnapshot | null> => {
    try {
      const params = buildSearchParams(page, limit, tagIDs, options);
      const data = await authFetch(`${API_BASE_URL}/action_logs?${params.toString()}`);
      if (!data.success) {
        throw new Error(data.message || "行動記録スナップショットの取得に失敗しました");
      }
      return {
        logs: data.data.logs || [],
        pagination: data.data.pagination,
        meta: data.data.meta as ActionLogListMeta | undefined,
      };
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録スナップショットの取得に失敗しました");
      return null;
    }
  };

  const fetchActionByID = async (id: number): Promise<ActionLog | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/action_logs/${id}`);
      if (!data.success) {
        throw new Error(data.message || "行動記録の取得に失敗しました");
      }
      return data.data as ActionLog;
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の取得に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  const createAction = async (
    title: string,
    parentPrototypeID: number | null,
    occurredAt: string,
    notes: string,
    tagIDs: number[],
    attributes: ActionAttribute[] = []
  ): Promise<ActionLog | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/action_logs`, {
        method: "POST",
        body: JSON.stringify({
          title,
          parent_prototype_id: parentPrototypeID ?? null,
          occurred_at: occurredAt,
          notes,
          tag_ids: tagIDs,
          attributes,
        }),
      });
      if (!data.success) {
        throw new Error(data.message || "行動記録の作成に失敗しました");
      }
      const created = data.data as ActionLog;
      setActions((prev) => [created, ...prev]);
      return created;
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の作成に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  const updateAction = async (
    id: number,
    title: string,
    parentPrototypeID: number | null,
    occurredAt: string,
    notes: string,
    tagIDs: number[],
    attributes: ActionAttribute[] = []
  ): Promise<ActionLog | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/action_logs/${id}`, {
        method: "PUT",
        body: JSON.stringify({
          title,
          parent_prototype_id: parentPrototypeID ?? null,
          occurred_at: occurredAt,
          notes,
          tag_ids: tagIDs,
          attributes,
        }),
      });
      if (!data.success) {
        throw new Error(data.message || "行動記録の更新に失敗しました");
      }
      const updated = data.data as ActionLog;
      setActions((prev) => prev.map((a) => (a.id === id ? updated : a)));
      return updated;
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の更新に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  const deleteAction = async (id: number): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/action_logs/${id}`, { method: "DELETE" });
      if (!data.success) {
        throw new Error(data.message || "行動記録の削除に失敗しました");
      }
      setActions((prev) => prev.filter((a) => a.id !== id));
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の削除に失敗しました");
      return false;
    } finally {
      setLoading(false);
    }
  };

  return {
    actions,
    pagination,
    meta,
    loading,
    error,
    fetchActions,
    fetchActionsSnapshot,
    fetchActionByID,
    createAction,
    updateAction,
    deleteAction,
  };
};
