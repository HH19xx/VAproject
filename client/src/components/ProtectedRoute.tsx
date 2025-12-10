import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import type { ReactNode } from "react";

// プロテクトルートコンポーネントの型定義
interface ProtectedRouteProps {
  children: ReactNode;
}

// 認証が必要なページを保護するコンポーネント
// 未認証の場合はログインページにリダイレクト
const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  const { userID, loading } = useAuth();
  const location = useLocation();

  console.log("[ProtectedRoute] 判定:", {
    path: location.pathname,
    loading,
    userID,
  });

  // 認証状態の読み込み中はローディング表示
  if (loading) {
    console.log("[ProtectedRoute] loading中のため一時表示に留めます");
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <p>読み込み中...</p>
      </div>
    );
  }

  // 未認証の場合はログインページにリダイレクト
  if (!userID) {
    console.log("[ProtectedRoute] 未認証。'/' へリダイレクトします from=", location.pathname);
    return <Navigate to="/" replace />;
  }

  // 認証済みの場合は子コンポーネントを表示
  console.log("[ProtectedRoute] 認証済み。子コンポーネントを表示します path=", location.pathname);
  return <>{children}</>;
};

export default ProtectedRoute;
