import { createContext, useContext, useEffect, useState } from "react";

interface AuthContextType {
  token: string | null;
  refreshToken: string | null;
  userID: number | null;
  login: (name: string, password: string) => Promise<void>;
  logout: () => void;
  loading: boolean;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [token, setToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [userID, setUserID] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);

  // localStorageへ保存するヘルパーをまとめる
  // アクセストークンとリフレッシュトークンをlocalStorageと状態に保存します。
  const persistTokens = (accessToken: string, newRefreshToken: string) => {
    localStorage.setItem("token", accessToken);
    localStorage.setItem("refresh_token", newRefreshToken);
    setToken(accessToken);
    setRefreshToken(newRefreshToken);
  };

  // 保存済みトークンを削除し、認証状態を初期化します。
  const clearTokens = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("refresh_token");
    setToken(null);
    setRefreshToken(null);
    setUserID(null);
  };

  // リフレッシュトークンでアクセストークンを再取得します。
  const attemptRefresh = async () => {
    if (!refreshToken) throw new Error("no refresh token");

    const res = await fetch("http://localhost:8080/api/v1/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!res.ok) throw new Error("refresh failed");

    const body = await res.json();
    const data = body?.data;
    if (!data?.access_token || !data?.refresh_token) {
      throw new Error("invalid refresh response");
    }
    persistTokens(data.access_token, data.refresh_token);
    return data.access_token as string;
  };

  useEffect(() => {
    const storedToken = localStorage.getItem("token");
    const storedRefresh = localStorage.getItem("refresh_token");
    if (storedToken) setToken(storedToken);
    if (storedRefresh) setRefreshToken(storedRefresh);
    setLoading(false);
  }, []);

  // トークンの変化に応じてセッション状態を確認します。
  useEffect(() => {
    let cancelled = false;

    const ensureSession = async () => {
      setLoading(true);

      try {
        if (!token) {
          if (refreshToken) {
            await attemptRefresh();
          } else {
            clearTokens();
          }
          return;
        }

        const res = await fetch("http://localhost:8080/api/v1/auth/me", {
          method: "GET",
          headers: { Authorization: `Bearer ${token}` },
        });

        if (res.status === 401 && refreshToken) {
          await attemptRefresh();
          return;
        }

        if (!res.ok) throw new Error("unauthorized");

        const body = await res.json();
        const id = body?.data?.id ?? null;
        if (!cancelled) setUserID(id);
      } catch (err) {
        clearTokens();
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    ensureSession();

    return () => {
      cancelled = true;
    };
  }, [token, refreshToken]);

  const login = async (name: string, password: string) => {
    // ユーザ名・パスワードでのログイン（新APIパス）
    const res = await fetch("http://localhost:8080/api/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, password }),
    });

    if (!res.ok) throw new Error("login failed");
    const body = await res.json();
    const data = body?.data;
    if (!data?.access_token || !data?.refresh_token) {
      throw new Error("invalid login response");
    }
    persistTokens(data.access_token, data.refresh_token);
  };

  const logout = () => {
    clearTokens();
  };

  return (
    <AuthContext.Provider value={{ token, refreshToken, userID, login, logout, loading }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
  return ctx;
};
