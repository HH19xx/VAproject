import { useContext, useState } from "react";
import { AuthContext } from "../contexts/AuthContext";

export interface Prototype {
  id: number;
  user_id: number;
  name: string;
  description: string;
  parent_prototype_id?: number;
  tag_ids: number[];
  created_at: string;
  create_user: string;
  updated_at: string;
  update_user: string;
  deleted_at?: string;
}

export const usePrototypes = () => {
  const authContext = useContext(AuthContext);
  const [prototypes, setPrototypes] = useState<Prototype[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) {
    throw new Error("usePrototypes must be used within AuthProvider");
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

  const fetchPrototypes = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/prototypes`);
      if (!data.success) {
        throw new Error(data.message || "プロトタイプの取得に失敗しました");
      }
      setPrototypes(data.data.prototypes || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "プロトタイプの取得に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  return {
    prototypes,
    loading,
    error,
    fetchPrototypes,
  };
};
