import { useContext, useState } from "react";
import { AuthContext } from "../contexts/AuthContext";

// 行動記録（action_logs）型定義
export interface ActionLog {
  id: number;
  target_id: number | null;
  action_type: number;
  timestamp: string;
  notes: string;
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

// useActions: 行動記録のCRUD
export const useActions = () => {
  const authContext = useContext(AuthContext);
  const [actions, setActions] = useState<ActionLog[]>([]);
  const [pagination, setPagination] = useState<Pagination>({ page: 1, limit: 20, total: 0 });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) throw new Error("useActions must be used within AuthProvider");
  const { token } = authContext;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

  const authFetch = async (url: string, options: RequestInit = {}) => {
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
      ...(options.headers as HeadersInit),
    };
    const res = await fetch(url, { ...options, headers });
    if (!res.ok) {
      const errBody = await res.json().catch(() => ({}));
      throw new Error(errBody?.error?.message || "APIエラーが発生しました");
    }
    return res.json();
  };

  // 一覧
  const fetchActions = async (page = 1, limit = 20) => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/action_logs?page=${page}&limit=${limit}`);
      if (data.success) {
        setActions(data.data.logs || []);
        setPagination(data.data.pagination);
      } else {
        throw new Error(data.message || "行動記録の取得に失敗しました");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の取得に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  // ID指定取得
  const fetchActionByID = async (id: number): Promise<ActionLog | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/action_logs/${id}`);
      if (data.success) return data.data as ActionLog;
      throw new Error(data.message || "行動記録の取得に失敗しました");
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の取得に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  // 作成（target_id は任意）
  const createAction = async (
    target_id: number | null,
    action_type: number,
    timestamp: string,
    notes: string
  ): Promise<ActionLog | null> => {
    setLoading(true);
    setError(null);
    try {
      const payload: any = { action_type, timestamp, notes };
      if (target_id) payload.target_id = target_id;

      const data = await authFetch(`${API_BASE_URL}/action_logs`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      if (data.success) {
        const created = data.data as ActionLog;
        setActions((prev) => [created, ...prev]);
        return created;
      }
      throw new Error(data.message || "行動記録の作成に失敗しました");
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の作成に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  // 更新
  const updateAction = async (
    id: number,
    target_id: number | null,
    action_type: number,
    timestamp: string,
    notes: string
  ): Promise<ActionLog | null> => {
    setLoading(true);
    setError(null);
    try {
      const payload: any = { action_type, timestamp, notes };
      if (target_id) payload.target_id = target_id;

      const data = await authFetch(`${API_BASE_URL}/action_logs/${id}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      });
      if (data.success) {
        const updated = data.data as ActionLog;
        setActions((prev) => prev.map((a) => (a.id === id ? updated : a)));
        return updated;
      }
      throw new Error(data.message || "行動記録の更新に失敗しました");
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の更新に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  // 削除
  const deleteAction = async (id: number): Promise<boolean> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/action_logs/${id}`, { method: "DELETE" });
      if (data.success) {
        setActions((prev) => prev.filter((a) => a.id !== id));
        return true;
      }
      throw new Error(data.message || "行動記録の削除に失敗しました");
    } catch (err) {
      setError(err instanceof Error ? err.message : "行動記録の削除に失敗しました");
      return false;
    } finally {
      setLoading(false);
    }
  };

  return { actions, pagination, loading, error, fetchActions, fetchActionByID, createAction, updateAction, deleteAction };
};
