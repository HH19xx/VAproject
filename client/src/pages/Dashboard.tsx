import { useAuth } from "../contexts/AuthContext";
import { useNavigate } from "react-router-dom";
import styles from "../assets/styles/Dashboard.module.scss";

// ダッシュボードページコンポーネント
const Dashboard = () => {
  const { userID, logout } = useAuth();
  const navigate = useNavigate();

  // ログアウト処理
  const handleLogout = () => {
    logout();
    navigate("/");
  };

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <h1>ダッシュボード</h1>
        <div className={styles.userInfo}>
          <span className={styles.userId}>ユーザーID: {userID}</span>
          <button onClick={() => navigate("/profile")} className="btn-secondary">
            プロフィール
          </button>
          <button onClick={handleLogout} className="btn-secondary">
            ログアウト
          </button>
        </div>
      </header>

      <main className={styles.main}>
        <div className={styles.grid}>
          <div className="card">
            <h2>ようこそ</h2>
            <p>ユーザーID {userID} さん、ようこそダッシュボードへ！</p>
            <p className="info-text">認証機能が正常に動作しています。</p>
          </div>

          <div className="card">
            <h2>統計情報</h2>
            <div className={styles.stats}>
              <div className={styles.statItem}>
                <span className={styles.statLabel}>観察対象</span>
                <span className={styles.statValue}>0</span>
              </div>
              <div className={styles.statItem}>
                <span className={styles.statLabel}>行動記録</span>
                <span className={styles.statValue}>0</span>
              </div>
              <div className={styles.statItem}>
                <span className={styles.statLabel}>行動種別</span>
                <span className={styles.statValue}>0</span>
              </div>
            </div>
          </div>

          <div className="card">
            <h2>クイックアクション</h2>
            <div className={styles.actions}>
              <button className="btn-primary" onClick={() => navigate("/targets")}>
                観察対象管理
              </button>
              <button className="btn-primary" onClick={() => navigate("/targets/new")}>
                観察対象を追加
              </button>
              <button className="btn-primary" disabled>
                行動を記録
              </button>
              <button className="btn-primary" disabled>
                統計を表示
              </button>
            </div>
            <p className="info-text">※行動記録と統計機能は今後実装予定です</p>
          </div>

          <div className="card">
            <h2>最近の活動</h2>
            <p className="info-text">まだ活動がありません</p>
          </div>
        </div>
      </main>
    </div>
  );
};

export default Dashboard;
