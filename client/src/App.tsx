import { BrowserRouter } from "react-router-dom";
import { AuthProvider } from "./contexts/AuthContext";
import "./assets/styles/main.scss";
import AppRoutes from "./AppRoutes";

// アプリケーションのルート設定とプロバイダー設定
function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
