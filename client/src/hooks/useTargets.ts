import { useState, useContext } from "react";
import { AuthContext } from "../contexts/AuthContext";

// 観察対象の型定義
export interface Target {
  id: number;
  name: string;
  description: string;
  created_at: string;
  create_user: string;
  updated_at: string;
  update_user: string;
  deleted_at?: string;
}

// ページネーション情報の型定義
interface Pagination {
  page: number;
  limit: number;
  total: number;
}

// APIレスポンスの型定義
interface TargetsResponse {
  targets: Target[];
  pagination: Pagination;
}

// useTargetsフック: 観察対象のCRUD操作を提供
export const useTargets = () => {
  const authContext = useContext(AuthContext);
  const [targets, setTargets] = useState<Target[]>([]);
  const [pagination, setPagination] = useState<Pagination>({ page: 1, limit: 20, total: 0 });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) {
    throw new Error("useTargets must be used within AuthProvider");
  }

  const { token } = authContext;
  const API_BASE_URL = "http://localhost:8080/api/v1";

  // 認証ヘッダーを含むfetchリクエストを実行するヘルパー関数
  const authFetch = async (url: string, options: RequestInit = {}) => {
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
      ...(options.headers as HeadersInit),
    };

    const response = await fetch(url, { ...options, headers });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error?.message || "APIエラーが発生しました");
    }

    return response.json();
  };

  // 観察対象の一覧を取得（ページネーション対応）
  const fetchTargets = async (page: number = 1, limit: number = 20) => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/targets?page=${page}&limit=${limit}`);

      if (data.success) {
        setTargets(data.data.targets || []);
        setPagination(data.data.pagination);
      } else {
        throw new Error(data.message || "観察対象の取得に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "観察対象の取得に失敗しました";
      setError(errorMessage);
      console.error("fetchTargets error:", err);
    } finally {
      setLoading(false);
    }
  };

  // 指定されたIDの観察対象を取得
  const fetchTargetByID = async (id: number): Promise<Target | null> => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/targets/${id}`);

      if (data.success) {
        return data.data as Target;
      } else {
        throw new Error(data.message || "観察対象の取得に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "観察対象の取得に失敗しました";
      setError(errorMessage);
      console.error("fetchTargetByID error:", err);
      return null;
    } finally {
      setLoading(false);
    }
  };

  // 新しい観察対象を作成
  const createTarget = async (name: string, description: string): Promise<Target | null> => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/targets`, {
        method: "POST",
        body: JSON.stringify({ name, description }),
      });

      if (data.success) {
        const newTarget = data.data as Target;
        // 一覧の先頭に新しい観察対象を追加
        setTargets((prev) => [newTarget, ...prev]);
        return newTarget;
      } else {
        throw new Error(data.message || "観察対象の作成に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "観察対象の作成に失敗しました";
      setError(errorMessage);
      console.error("createTarget error:", err);
      return null;
    } finally {
      setLoading(false);
    }
  };

  // 既存の観察対象を更新
  const updateTarget = async (id: number, name: string, description: string): Promise<Target | null> => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/targets/${id}`, {
        method: "PUT",
        body: JSON.stringify({ name, description }),
      });

      if (data.success) {
        const updatedTarget = data.data as Target;
        // 一覧内の該当する観察対象を更新
        setTargets((prev) => prev.map((t) => (t.id === id ? updatedTarget : t)));
        return updatedTarget;
      } else {
        throw new Error(data.message || "観察対象の更新に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "観察対象の更新に失敗しました";
      setError(errorMessage);
      console.error("updateTarget error:", err);
      return null;
    } finally {
      setLoading(false);
    }
  };

  // 観察対象を削除（論理削除）
  const deleteTarget = async (id: number): Promise<boolean> => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/targets/${id}`, {
        method: "DELETE",
      });

      if (data.success) {
        // 一覧から削除された観察対象を除外
        setTargets((prev) => prev.filter((t) => t.id !== id));
        return true;
      } else {
        throw new Error(data.message || "観察対象の削除に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "観察対象の削除に失敗しました";
      setError(errorMessage);
      console.error("deleteTarget error:", err);
      return false;
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
  };
};
