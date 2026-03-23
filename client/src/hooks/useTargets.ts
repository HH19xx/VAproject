import { useContext, useState } from "react";
import { AuthContext } from "../contexts/AuthContext";

export interface Target {
  id: number;
  user_id: number;
  name: string;
  description: string;
  match_mode: "AND";
  query_text: string;
  tag_ids: number[];
  any_tag_ids: number[];
  any_tag_groups: number[][];
  exclude_tag_ids: number[];
  created_at: string;
  create_user: string;
  updated_at: string;
  update_user: string;
  deleted_at?: string;
}

export interface ActionLog {
  id: number;
  user_id: number;
  occurred_at: string;
  notes: string;
  tag_ids: number[];
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

interface TargetActionLogsResponse {
  logs: ActionLog[];
  pagination: Pagination;
}

interface TargetQueryPayload {
  queryText?: string;
  anyTagIDs?: number[];
  anyTagGroups?: number[][];
  excludeTagIDs?: number[];
}

export const useTargets = () => {
  const authContext = useContext(AuthContext);
  const [targets, setTargets] = useState<Target[]>([]);
  const [pagination, setPagination] = useState<Pagination>({ page: 1, limit: 20, total: 0 });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) {
    throw new Error("useTargets は AuthProvider 内で使用してください");
  }

  const { authFetch } = authContext;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";


  const fetchTargets = async (page = 1, limit = 20) => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/targets?page=${page}&limit=${limit}`);
      if (!data.success) {
        throw new Error(data.message || "観察対象の取得に失敗しました");
      }
      setTargets(data.data.targets || []);
      setPagination(data.data.pagination);
    } catch (err) {
      setError(err instanceof Error ? err.message : "観察対象の取得に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  const fetchTargetByID = async (id: number): Promise<Target | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/targets/${id}`);
      if (!data.success) {
        throw new Error(data.message || "観察対象の取得に失敗しました");
      }
      return data.data as Target;
    } catch (err) {
      setError(err instanceof Error ? err.message : "観察対象の取得に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  const createTarget = async (
    name: string,
    description: string,
    tagIDs: number[],
    query: TargetQueryPayload = {}
  ): Promise<Target | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/targets`, {
        method: "POST",
        body: JSON.stringify({
          name,
          description,
          match_mode: "AND",
          query_text: query.queryText || "",
          tag_ids: tagIDs,
          any_tag_ids: query.anyTagIDs || [],
          any_tag_groups: query.anyTagGroups || [],
          exclude_tag_ids: query.excludeTagIDs || [],
        }),
      });
      if (!data.success) {
        throw new Error(data.message || "観察対象の作成に失敗しました");
      }
      const created = data.data as Target;
      setTargets((prev) => [created, ...prev]);
      return created;
    } catch (err) {
      setError(err instanceof Error ? err.message : "観察対象の作成に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  const updateTarget = async (
    id: number,
    name: string,
    description: string,
    tagIDs: number[],
    query: TargetQueryPayload = {}
  ): Promise<Target | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/targets/${id}`, {
        method: "PUT",
        body: JSON.stringify({
          name,
          description,
          match_mode: "AND",
          query_text: query.queryText || "",
          tag_ids: tagIDs,
          any_tag_ids: query.anyTagIDs || [],
          any_tag_groups: query.anyTagGroups || [],
          exclude_tag_ids: query.excludeTagIDs || [],
        }),
      });
      if (!data.success) {
        throw new Error(data.message || "観察対象の更新に失敗しました");
      }
      const updated = data.data as Target;
      setTargets((prev) => prev.map((t) => (t.id === id ? updated : t)));
      return updated;
    } catch (err) {
      setError(err instanceof Error ? err.message : "観察対象の更新に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  const deleteTarget = async (id: number): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/targets/${id}`, {
        method: "DELETE",
      });
      if (!data.success) {
        throw new Error(data.message || "観察対象の削除に失敗しました");
      }
      setTargets((prev) => prev.filter((t) => t.id !== id));
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : "観察対象の削除に失敗しました");
      return false;
    } finally {
      setLoading(false);
    }
  };

  const fetchActionLogsByTarget = async (targetID: number, page = 1, limit = 20): Promise<TargetActionLogsResponse | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/targets/${targetID}/action_logs?page=${page}&limit=${limit}`);
      if (!data.success) {
        throw new Error(data.message || "観察対象による行動記録の取得に失敗しました");
      }
      return data.data as TargetActionLogsResponse;
    } catch (err) {
      setError(err instanceof Error ? err.message : "観察対象による行動記録の取得に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  return {
    targets,
    pagination,
    loading,
    error,
    fetchTargets,
    fetchTargetByID,
    createTarget,
    updateTarget,
    deleteTarget,
    fetchActionLogsByTarget,
  };
};

