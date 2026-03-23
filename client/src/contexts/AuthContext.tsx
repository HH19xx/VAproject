import { createContext, useContext, useEffect, useState } from "react";
import type { ReactNode } from "react";

// 認証コンテキストの型定義
interface AuthContextType {
  token: string | null;
  refreshToken: string | null;
  userID: number | null;
  login: (name: string, password: string) => Promise<void>;
  logout: () => void;
  authFetch: (url: string, options?: RequestInit) => Promise<any>;
  loading: boolean;
}

// 認証コンテキストを作成
export const AuthContext = createContext<AuthContextType | undefined>(undefined);

// 認証プロバイダーコンポーネント
export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [token, setToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [userID, setUserID] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [initialized, setInitialized] = useState(false);

  // アクセストークンとリフレッシュトークンをlocalStorageと状態に保存
  const persistTokens = (accessToken: string, newRefreshToken: string) => {

    localStorage.setItem("token", accessToken);
    localStorage.setItem("refresh_token", newRefreshToken);
    setToken(accessToken);
    setRefreshToken(newRefreshToken);
  };

  // 保存済みトークンを削除し、認証状態を初期化
  const clearTokens = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("refresh_token");
    setToken(null);
    setRefreshToken(null);
    setUserID(null);
  };

  // リフレッシュトークンでアクセストークンを再取得
  const attemptRefresh = async () => {
    if (!refreshToken) {
      throw new Error("no refresh token");
    }

    const res = await fetch("http://localhost:8080/api/v1/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });


    if (!res.ok) {
      throw new Error("refresh failed");
    }

    const body = await res.json();

    const data = body?.data;
    if (!data?.access_token || !data?.refresh_token) {
      throw new Error("invalid refresh response");
    }
    persistTokens(data.access_token, data.refresh_token);
    return data.access_token as string;
  };

  // ★ 初回マウント時にlocalStorageからトークンを読み込み（ここではloadingをfalseにしない）
  useEffect(() => {
    const storedToken = localStorage.getItem("token");
    const storedRefresh = localStorage.getItem("refresh_token");


    if (storedToken) setToken(storedToken);
    if (storedRefresh) setRefreshToken(storedRefresh);

    setInitialized(true); // ★ ここで「localStorage読み込み完了」を宣言
  }, []);

  // トークンの変化に応じてセッション状態を確認
  useEffect(() => {
    let cancelled = false;

    const ensureSession = async () => {
      // ★ localStorage復元前なら何もしない
      if (!initialized) {
        return;
      }


      // token・refreshToken 両方ないなら「素の未ログイン」と見なして終了
      if (!token && !refreshToken) {
        if (!cancelled) {
          setUserID(null);
          setLoading(false);
        }
        return;
      }

      setLoading(true);

      try {
        // token がないが refreshToken がある場合はリフレッシュを試行
        if (!token && refreshToken) {
          await attemptRefresh();
          return;
        }

        if (!token) {
          clearTokens();
          if (!cancelled) setLoading(false);
          return;
        }

        // /api/v1/auth/meエンドポイントでユーザー情報を取得
        const res = await fetch("http://localhost:8080/api/v1/auth/me", {
          method: "GET",
          headers: { Authorization: `Bearer ${token}` },
        });


        // 401エラーの場合はトークンリフレッシュを試行
        if (res.status === 401 && refreshToken) {
          await attemptRefresh();
          return;
        }

        if (!res.ok) {
          throw new Error("unauthorized");
        }

        const body = await res.json();

        const id = body?.data?.id ?? null;
        if (!cancelled) {
          setUserID(id);
        }
      } catch (err) {
        clearTokens();
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    ensureSession();

    return () => {
      cancelled = true;
    };
  }, [initialized, token, refreshToken]); // ★ initialized も依存に追加

  // ユーザー名・パスワードでログイン
  const login = async (name: string, password: string) => {

    const res = await fetch("http://localhost:8080/api/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, password }),
    });


    if (!res.ok) {
      throw new Error("login failed");
    }

    const body = await res.json();
    const data = body?.data;
    if (!data?.access_token || !data?.refresh_token) {
      throw new Error("invalid login response");
    }
    persistTokens(data.access_token, data.refresh_token);
  };

  // ログアウト処理
  const authFetch = async (url: string, options: RequestInit = {}) => {
    const buildHeaders = (accessToken: string | null): HeadersInit => {
      const headers: HeadersInit = {
        ...(options.body ? { "Content-Type": "application/json" } : {}),
        ...(options.headers as HeadersInit),
      };
      if (accessToken) {
        return {
          ...headers,
          Authorization: `Bearer ${accessToken}`,
        };
      }
      return headers;
    };

    const run = async (accessToken: string | null) =>
      fetch(url, {
        ...options,
        headers: buildHeaders(accessToken),
      });

    let currentToken = token ?? localStorage.getItem("token");
    let response = await run(currentToken);

    if (response.status === 401) {
      try {
        currentToken = await attemptRefresh();
        response = await run(currentToken);
      } catch (err) {
        clearTokens();
        throw err instanceof Error ? err : new Error("refresh failed");
      }
    }

    if (!response.ok) {
      const errBody = await response.json().catch(() => ({}));
      throw new Error(errBody?.error?.message || "API request failed");
    }

    return response.json();
  };
  const logout = () => {
    clearTokens();
  };


  return (
    <AuthContext.Provider
      value={{ token, refreshToken, userID, login, logout, authFetch, loading }}
    >
      {children}
    </AuthContext.Provider>
  );
};

// 認証コンテキストを使用するカスタムフック
export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
  return ctx;
};

