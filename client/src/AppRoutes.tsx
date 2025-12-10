// src/AppRoutes.tsx
import { Routes, Route, useNavigate, useLocation } from "react-router-dom";
import { useEffect } from "react";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";
import Profile from "./pages/Profile";
import Targets from "./pages/Targets";
import TargetForm from "./pages/TargetForm";
import ProtectedRoute from "./components/ProtectedRoute";

function AppRoutes() {
  const navigate = useNavigate();
  const location = useLocation();

  // OAuth コールバックのパラメータを拾う処理
  useEffect(() => {
    console.log("[AppRoutes] 初期化: 現在URL =", window.location.href);

    const params = new URLSearchParams(window.location.search);
    const token = params.get("token");
    const refresh = params.get("refresh");
    const expires = params.get("expires");

    console.log(
      "[AppRoutes] OAuthパラメータ取得:",
      "token =", token,
      "refresh =", refresh,
      "expires =", expires
    );

    if (token && refresh) {
      console.log("[AppRoutes] localStorage にトークンを保存します");
      localStorage.setItem("token", token);
      localStorage.setItem("refresh_token", refresh);
      if (expires) {
        localStorage.setItem("token_expires_at", expires);
      }

      // URL クエリを消す
      console.log("[AppRoutes] URL クエリを削除します");
      window.history.replaceState({}, document.title, window.location.pathname);

      console.log("[AppRoutes] /dashboard へ navigate します");
      navigate("/dashboard", { replace: true });
    } else {
      console.log("[AppRoutes] OAuthパラメータなし（通常のページ表示）");
    }
  }, [navigate]);

  // ルーティングが変わるたびのログ（確認用）
  useEffect(() => {
    console.log("[AppRoutes] ルート変更:", location.pathname, location.search);
  }, [location]);

  return (
    <Routes>
      <Route path="/" element={<Login />} />
      <Route path="/register" element={<Register />} />

      <Route
        path="/dashboard"
        element={
          <ProtectedRoute>
            <Dashboard />
          </ProtectedRoute>
        }
      />
      <Route
        path="/profile"
        element={
          <ProtectedRoute>
            <Profile />
          </ProtectedRoute>
        }
      />
      <Route
        path="/targets"
        element={
          <ProtectedRoute>
            <Targets />
          </ProtectedRoute>
        }
      />
      <Route
        path="/targets/new"
        element={
          <ProtectedRoute>
            <TargetForm />
          </ProtectedRoute>
        }
      />
      <Route
        path="/targets/:id/edit"
        element={
          <ProtectedRoute>
            <TargetForm />
          </ProtectedRoute>
        }
      />
    </Routes>
  );
}

export default AppRoutes;
