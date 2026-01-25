import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import styles from "../assets/styles/Register.module.scss";

// ユーザー登録ページ
const Register = () => {
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // ユーザー登録処理
  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // クライアント側バリデーション
    if (!username.trim()) {
      setError("ユーザー名を入力してください");
      return;
    }

    if (password.length < 6) {
      setError("パスワードは6文字以上で入力してください");
      return;
    }

    if (password !== confirmPassword) {
      setError("パスワードが一致しません");
      return;
    }

    setLoading(true);

    try {
      // 登録APIを呼び出し
      const response = await fetch("http://localhost:8080/api/v1/auth/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: username.trim(),
          password: password,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error?.message || "登録に失敗しました");
      }

      // 登録成功後、自動的にログイン処理を実行（バックエンドがトークンを返す）
      if (data.success && data.data.access_token && data.data.refresh_token) {
        // AuthContextのloginメソッドは使わず、直接localStorageに保存してリダイレクト
        localStorage.setItem("token", data.data.access_token);
        localStorage.setItem("refresh_token", data.data.refresh_token);

        alert("ユーザー登録が完了しました");
        navigate("/dashboard");
        // ページリロードで認証状態を反映
        window.location.reload();
      } else {
        throw new Error("トークンの取得に失敗しました");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "登録に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h1>ユーザー登録</h1>

        <form onSubmit={handleRegister} className={styles.form}>
          <div className={styles.formGroup}>
            <label htmlFor="username">ユーザー名</label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={loading}
              placeholder="ユーザー名を入力"
            />
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="password">パスワード</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={loading}
              placeholder="パスワードを入力（6文字以上）"
            />
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="confirmPassword">パスワード（確認）</label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              disabled={loading}
              placeholder="パスワードを再入力"
            />
          </div>

          {error && <p className={styles.error}>{error}</p>}

          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? "登録中..." : "登録"}
          </button>

          <p className={styles.linkText}>
            <Link to="/">ログイン画面に戻る</Link>
          </p>
        </form>
      </div>
    </div>
  );
};

export default Register;
