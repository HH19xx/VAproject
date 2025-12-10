import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import styles from "../assets/styles/Profile.module.scss";

interface ProfileData {
  id: number;
  name: string;
  email: string;
  auth_providers: string[];
}

const Profile = () => {
  const navigate = useNavigate();
  const [profile, setProfile] = useState<ProfileData | null>(null);
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // プロフィール情報を取得する
  useEffect(() => {
    const fetchProfile = async () => {
      const token = localStorage.getItem("token");
      if (!token) {
        navigate("/");
        return;
      }

      try {
        const response = await fetch("http://localhost:8080/api/v1/auth/profile", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.error?.message || "プロフィール情報の取得に失敗しました");
        }

        if (data.success && data.data) {
          setProfile(data.data);
          setUsername(data.data.name);
          setEmail(data.data.email);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "エラーが発生しました");
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, [navigate]);

  // プロフィール更新処理
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccessMessage(null);

    if (!username.trim()) {
      setError("ユーザー名を入力してください");
      return;
    }

    setSaving(true);

    try {
      const token = localStorage.getItem("token");
      const response = await fetch("http://localhost:8080/api/v1/auth/profile", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name: username.trim() }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error?.message || "プロフィール更新に失敗しました");
      }

      if (data.success) {
        setSuccessMessage("プロフィールを更新しました");
        // プロフィール情報を再取得する
        setProfile((prev) => (prev ? { ...prev, name: username } : null));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新に失敗しました");
    } finally {
      setSaving(false);
    }
  };

  // OAuth認証かどうかを判定する
  const isOAuthUser = profile?.auth_providers.includes("google") || false;
  // local認証を持っているかどうか
  const hasLocalAuth = profile?.auth_providers.includes("local") || false;

  if (loading) {
    return (
      <div className={styles.container}>
        <div className={styles.card}>
          <p>読み込み中...</p>
        </div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className={styles.container}>
        <div className={styles.card}>
          <p>プロフィール情報の取得に失敗しました</p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h1>プロフィール設定</h1>

        {error && <div className={styles.error}>{error}</div>}
        {successMessage && <div className={styles.success}>{successMessage}</div>}

        <form onSubmit={handleSubmit}>
          <div className={styles.formGroup}>
            <label htmlFor="username">ユーザー名</label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="ユーザー名"
            />
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="email">メールアドレス</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="メールアドレス"
              disabled={isOAuthUser}
              className={isOAuthUser ? styles.disabled : ""}
            />
            {isOAuthUser && (
              <p className={styles.helpText}>
                OAuth認証で登録されたメールアドレスは変更できません
              </p>
            )}
          </div>

          <div className={styles.authProviders}>
            <h3>認証方法</h3>
            <div className={styles.providerList}>
              {hasLocalAuth && (
                <span className={styles.badge}>パスワード認証</span>
              )}
              {isOAuthUser && (
                <span className={`${styles.badge} ${styles.badgeOauth}`}>Google認証</span>
              )}
            </div>
          </div>

          <div className={styles.actions}>
            <button
              type="submit"
              className="btn-primary"
              disabled={saving}
            >
              {saving ? "保存中..." : "保存"}
            </button>
            <button
              type="button"
              onClick={() => navigate("/dashboard")}
              className="btn-secondary"
            >
              戻る
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default Profile;
