import { useContext, useState } from "react";
import { AuthContext } from "../contexts/AuthContext";

export interface Tag {
  id: number;
  user_id: number;
  name: string;
  group_name?: string;
  description?: string;
}

export const useTags = () => {
  const authContext = useContext(AuthContext);
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!authContext) {
    throw new Error("useTags must be used within AuthProvider");
  }

  const { authFetch } = authContext;
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";


  const fetchTags = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/tags`);
      if (!data.success) {
        throw new Error(data.message || "タグの取得に失敗しました");
      }
      setTags(data.data.tags || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "タグの取得に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  const createTag = async (name: string, group = "", description = ""): Promise<Tag | null> => {
    setLoading(true);
    setError(null);
    try {
      const data = await authFetch(`${API_BASE_URL}/tags`, {
        method: "POST",
        body: JSON.stringify({
          name,
          group,
          description,
        }),
      });
      if (!data.success) {
        throw new Error(data.message || "タグの作成に失敗しました");
      }
      const created = data.data as Tag;
      setTags((prev) => {
        if (prev.some((t) => t.id === created.id)) {
          return prev;
        }
        return [...prev, created];
      });
      return created;
    } catch (err) {
      setError(err instanceof Error ? err.message : "タグの作成に失敗しました");
      return null;
    } finally {
      setLoading(false);
    }
  };

  return {
    tags,
    loading,
    error,
    fetchTags,
    createTag,
  };
};

