import { useState, useContext } from "react";
import { AuthContext } from "../contexts/AuthContext";

// 行動種別の型定義
export interface ActionType {
  id: number;
  action_name: string;
  description: string;
  created_at: string;
  create_user: string;
  updated_at: string;
  update_user: string;
  deleted_at?: string;
}

// useActionTypesフック: 行動種別のCRUD操作を提供
export const useActionTypes = () => {
  const authContext = useContext(AuthContext);
  const [actionTypes, setActionTypes] = useState<ActionType[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) {
    throw new Error("useActionTypes must be used within AuthProvider");
  }

  const { token } = authContext;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

  // 認証ヘッダーを含むfetchリクエスト
  const authFetch = async (url: string, options: RequestInit = {}) => {
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
      ...(options.headers as HeadersInit),
    };

    const response = await fetch(url, { ...options, headers });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error?.message || "APIエラーが発生しました");
    }

    return response.json();
  };

  // 行動種別の一覧を取得
  const fetchActionTypes = async () => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/action_types`);

      if (data.success) {
        setActionTypes(data.data.action_types || []);
      } else {
        throw new Error(data.message || "行動種別の取得に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "行動種別の取得に失敗しました";
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  // 指定されたIDの行動種別を取得
  const fetchActionTypeByID = async (id: number): Promise<ActionType | null> => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/action_types/${id}`);

      if (data.success) {
        return data.data as ActionType;
      } else {
        throw new Error(data.message || "行動種別の取得に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "行動種別の取得に失敗しました";
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  // 新しい行動種別を作成
  const createActionType = async (action_name: string, description: string): Promise<ActionType | null> => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/action_types`, {
        method: "POST",
        body: JSON.stringify({ action_name, description }),
      });

      if (data.success) {
        const newActionType = data.data as ActionType;
        setActionTypes((prev) => [newActionType, ...prev]);
        return newActionType;
      } else {
        throw new Error(data.message || "行動種別の作成に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "行動種別の作成に失敗しました";
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  // 既存の行動種別を更新
  const updateActionType = async (id: number, action_name: string, description: string): Promise<ActionType | null> => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/action_types/${id}`, {
        method: "PUT",
        body: JSON.stringify({ action_name, description }),
      });

      if (data.success) {
        const updatedActionType = data.data as ActionType;
        setActionTypes((prev) => prev.map((t) => (t.id === id ? updatedActionType : t)));
        return updatedActionType;
      } else {
        throw new Error(data.message || "行動種別の更新に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "行動種別の更新に失敗しました";
      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  // 行動種別を削除（論理削除）
  const deleteActionType = async (id: number): Promise<boolean> => {
    setLoading(true);
    setError(null);

    try {
      const data = await authFetch(`${API_BASE_URL}/action_types/${id}`, {
        method: "DELETE",
      });

      if (data.success) {
        setActionTypes((prev) => prev.filter((t) => t.id !== id));
        return true;
      } else {
        throw new Error(data.message || "行動種別の削除に失敗しました");
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "行動種別の削除に失敗しました";
      setError(errorMessage);
      return false;
    } finally {
      setLoading(false);
    }
  };

  return {
    actionTypes,
    loading,
    error,
    fetchActionTypes,
    fetchActionTypeByID,
    createActionType,
    updateActionType,
    deleteActionType,
  };
};
