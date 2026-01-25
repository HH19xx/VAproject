import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import type { ReactNode } from "react";

// 認証が必要なルートを保護するコンポーネント
interface ProtectedRouteProps {
  children: ReactNode;
}

const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  const { userID, loading } = useAuth();
  const location = useLocation();

  // セッション確認中はローディング表示
  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <p>読み込み中...</p>
      </div>
    );
  }

  // 未認証ならログイン画面へリダイレクト
  if (!userID) {
    return <Navigate to="/" replace state={{ from: location.pathname }} />;
  }

  // 認証済みなら子要素を表示
  return <>{children}</>;
};

export default ProtectedRoute;
