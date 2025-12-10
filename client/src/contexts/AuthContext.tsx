import { createContext, useContext, useEffect, useState } from "react";
import type { ReactNode } from "react";

// 認証コンテキストの型定義
interface AuthContextType {
  token: string | null;
  refreshToken: string | null;
  userID: number | null;
  login: (name: string, password: string) => Promise<void>;
  logout: () => void;
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
  const [initialized, setInitialized] = useState(false); // ★ localStorage復元完了フラグ

  // アクセストークンとリフレッシュトークンをlocalStorageと状態に保存
  const persistTokens = (accessToken: string, newRefreshToken: string) => {
    console.log("[AuthContext] persistTokens 呼び出し:", {
      accessToken: accessToken?.slice(0, 16) + "...",
      newRefreshToken: newRefreshToken?.slice(0, 16) + "...",
    });

    localStorage.setItem("token", accessToken);
    localStorage.setItem("refresh_token", newRefreshToken);
    setToken(accessToken);
    setRefreshToken(newRefreshToken);
  };

  // 保存済みトークンを削除し、認証状態を初期化
  const clearTokens = () => {
    console.log("[AuthContext] clearTokens 呼び出し: トークンと userID をクリアします");
    localStorage.removeItem("token");
    localStorage.removeItem("refresh_token");
    setToken(null);
    setRefreshToken(null);
    setUserID(null);
  };

  // リフレッシュトークンでアクセストークンを再取得
  const attemptRefresh = async () => {
    console.log("[AuthContext] attemptRefresh 開始: refreshToken =", refreshToken?.slice(0, 16) + "...");
    if (!refreshToken) {
      console.log("[AuthContext] attemptRefresh 中止: refreshToken がありません");
      throw new Error("no refresh token");
    }

    const res = await fetch("http://localhost:8080/api/v1/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    console.log("[AuthContext] /auth/refresh レスポンスステータス:", res.status);

    if (!res.ok) {
      console.log("[AuthContext] /auth/refresh 失敗");
      throw new Error("refresh failed");
    }

    const body = await res.json();
    console.log("[AuthContext] /auth/refresh レスポンスボディ:", body);

    const data = body?.data;
    if (!data?.access_token || !data?.refresh_token) {
      console.log("[AuthContext] /auth/refresh レスポンス不正: data:", data);
      throw new Error("invalid refresh response");
    }
    persistTokens(data.access_token, data.refresh_token);
    return data.access_token as string;
  };

  // ★ 初回マウント時にlocalStorageからトークンを読み込み（ここではloadingをfalseにしない）
  useEffect(() => {
    console.log("[AuthContext] 初期化: localStorage からトークンを読み込みます");
    const storedToken = localStorage.getItem("token");
    const storedRefresh = localStorage.getItem("refresh_token");

    console.log("[AuthContext] localStorage 読み取り:", {
      storedToken: storedToken ? storedToken.slice(0, 16) + "..." : null,
      storedRefresh: storedRefresh ? storedRefresh.slice(0, 16) + "..." : null,
    });

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
        console.log("[AuthContext] ensureSession スキップ: initialized=false");
        return;
      }

      console.log("[AuthContext] ensureSession 開始:", {
        token: token ? token.slice(0, 16) + "..." : null,
        refreshToken: refreshToken ? refreshToken.slice(0, 16) + "..." : null,
      });

      // token・refreshToken 両方ないなら「素の未ログイン」と見なして終了
      if (!token && !refreshToken) {
        console.log(
          "[AuthContext] token / refreshToken どちらも存在せず → 未ログインとして loading=false にします"
        );
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
          console.log("[AuthContext] token なし。refreshToken でリフレッシュを試みます");
          await attemptRefresh();
          console.log("[AuthContext] リフレッシュ成功。再度 ensureSession が走る想定です");
          return;
        }

        if (!token) {
          console.log("[AuthContext] token も refreshToken も不正 → 未ログイン扱い");
          clearTokens();
          if (!cancelled) setLoading(false);
          return;
        }

        // /api/v1/auth/meエンドポイントでユーザー情報を取得
        console.log("[AuthContext] /auth/me でユーザー確認を行います");
        const res = await fetch("http://localhost:8080/api/v1/auth/me", {
          method: "GET",
          headers: { Authorization: `Bearer ${token}` },
        });

        console.log("[AuthContext] /auth/me レスポンスステータス:", res.status);

        // 401エラーの場合はトークンリフレッシュを試行
        if (res.status === 401 && refreshToken) {
          console.log("[AuthContext] /auth/me が 401。refreshToken でリフレッシュを試みます");
          await attemptRefresh();
          return;
        }

        if (!res.ok) {
          console.log("[AuthContext] /auth/me がエラー応答:", res.status);
          throw new Error("unauthorized");
        }

        const body = await res.json();
        console.log("[AuthContext] /auth/me レスポンスボディ:", body);

        const id = body?.data?.id ?? null;
        if (!cancelled) {
          console.log("[AuthContext] userID をセット:", id);
          setUserID(id);
        }
      } catch (err) {
        console.error("[AuthContext] セッション確認エラー:", err);
        clearTokens();
      } finally {
        if (!cancelled) {
          setLoading(false);
          console.log("[AuthContext] ensureSession 完了: loading=false, userID=", userID);
        }
      }
    };

    ensureSession();

    return () => {
      console.log("[AuthContext] ensureSession cleanup: cancelled=true");
      cancelled = true;
    };
  }, [initialized, token, refreshToken]); // ★ initialized も依存に追加

  // ユーザー名・パスワードでログイン
  const login = async (name: string, password: string) => {
    console.log("[AuthContext] login 開始:", { name });

    const res = await fetch("http://localhost:8080/api/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, password }),
    });

    console.log("[AuthContext] /auth/login レスポンスステータス:", res.status);

    if (!res.ok) {
      console.log("[AuthContext] /auth/login 失敗");
      throw new Error("login failed");
    }

    const body = await res.json();
    console.log("[AuthContext] /auth/login レスポンスボディ:", body);
    const data = body?.data;
    if (!data?.access_token || !data?.refresh_token) {
      console.log("[AuthContext] /auth/login レスポンス不正: data=", data);
      throw new Error("invalid login response");
    }
    persistTokens(data.access_token, data.refresh_token);
  };

  // ログアウト処理
  const logout = () => {
    console.log("[AuthContext] logout 呼び出し");
    clearTokens();
  };

  console.log("[AuthContext] Provider レンダリング: state =", {
    token: token ? token.slice(0, 16) + "..." : null,
    refreshToken: refreshToken ? refreshToken.slice(0, 16) + "..." : null,
    userID,
    loading,
    initialized,
  });

  return (
    <AuthContext.Provider
      value={{ token, refreshToken, userID, login, logout, loading }}
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
