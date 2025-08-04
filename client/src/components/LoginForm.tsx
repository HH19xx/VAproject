import { useState } from "react";
import { useAuth } from "../contexts/AuthContext";

const LoginForm = () => {
  const { login } = useAuth();
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login(name, password);
    } catch {
      setError("ログイン失敗");
    }
  };

  const handleGoogleLogin = () => {
    window.location.href = "http://localhost:8080/auth/google/login";
  };

  return (
    <div>
      <form onSubmit={handleSubmit}>
        <h2>ログイン</h2>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="名前"
        />
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="パスワード"
        />
        <button type="submit">ログイン</button>
        {error && <p style={{ color: "red" }}>{error}</p>}
      </form>

      <hr />

      <button onClick={handleGoogleLogin}>Googleでログイン</button>
    </div>
  );
};

export default LoginForm;
