import { useAuth as useAuthContext } from "../contexts/AuthContext";

// コンテキスト版useAuthを再エクスポートし、既存のインポート経路と互換性を保ちます。
const useAuth = () => useAuthContext();

export default useAuth;
