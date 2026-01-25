// src/AppRoutes.tsx
import { Routes, Route, useNavigate, useLocation } from "react-router-dom";
import { useEffect } from "react";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";
import Profile from "./pages/Profile";
import Targets from "./pages/Targets";
import TargetForm from "./pages/TargetForm";
import Actions from "./pages/Actions";
import ActionForm from "./pages/ActionForm";
import ProtectedRoute from "./components/ProtectedRoute";

function AppRoutes() {
  const navigate = useNavigate();
  const location = useLocation();

  // OAuth コールバックから戻った場合にクエリを読み取り、トークンを保存する
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const token = params.get("token");
    const refresh = params.get("refresh");
    const expires = params.get("expires");

    if (token && refresh) {
      try {
        localStorage.setItem("token", token);
        localStorage.setItem("refresh_token", refresh);
        if (expires) {
          localStorage.setItem("token_expires_at", expires);
        }
      } catch {
        // localStorage が利用できない場合は無視
      }

      // クエリを除去
      window.history.replaceState({}, document.title, window.location.pathname);
      navigate("/dashboard", { replace: true });
    }
  }, [navigate]);

  // ページ遷移時の副作用が必要ならここで実装（現状は何もしない）
  useEffect(() => {
    void location;
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
      <Route
        path="/actions"
        element={
          <ProtectedRoute>
            <Actions />
          </ProtectedRoute>
        }
      />
      <Route
        path="/actions/new"
        element={
          <ProtectedRoute>
            <ActionForm />
          </ProtectedRoute>
        }
      />
      <Route
        path="/actions/:id/edit"
        element={
          <ProtectedRoute>
            <ActionForm />
          </ProtectedRoute>
        }
      />
    </Routes>
  );
}

export default AppRoutes;
