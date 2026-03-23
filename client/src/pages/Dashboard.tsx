import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import styles from "../assets/styles/Dashboard.module.scss";

const Dashboard = () => {
  const { userID, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/");
  };

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <div>
          <h1>ダッシュボード</h1>
          <p className={styles.subTitle}>使いたい画面を選んでください。</p>
        </div>
        <div className={styles.userInfo}>
          <span className={styles.userId}>{`ユーザーID: ${userID}`}</span>
          <button className="btn-secondary" onClick={() => navigate("/profile")}>プロフィール</button>
          <button className="btn-secondary" onClick={handleLogout}>ログアウト</button>
        </div>
      </header>

      <main className={styles.main}>
        <section className={styles.block}>
          <div className={styles.blockHeader}>
            <h2>記録機能</h2>
          </div>
          <div className={styles.grid}>
            <button className={`${styles.actionCard} btn-primary`} onClick={() => navigate("/records/actions-form")}>
            <strong>行動記録を追加</strong>
            <span>タグと属性を持つ記録を作成します。</span>
          </button>
          <button className={`${styles.actionCard} btn-primary`} onClick={() => navigate("/statistics/external-data-sources")}>
            <strong>外部データを追加</strong>
            <span>Open-Meteo と e-Stat を取得して保存します。</span>
          </button>
          </div>
        </section>
        <section className={styles.block}>
          <div className={styles.blockHeader}>
            <h2>Déconstruction機能</h2>
            <p>記録を作り、タグ検索と部分集合比較を行います。</p>
          </div>
          <div className={styles.grid}>
            <button className={`${styles.actionCard} btn-primary`} onClick={() => navigate("/constructions/actions")}>
              <strong>行動記録を見る</strong>
              <span>検索、比較、可視化を行います。</span>
            </button>
          </div>
        </section>

        <section className={styles.block}>
          <div className={styles.blockHeader}>
            <h2>Quantification機能</h2>
            <p>推定と検定を行います。</p>
          </div>
          <div className={styles.gridSingle}>
            <button className={`${styles.actionCard} btn-primary`} onClick={() => navigate("/statistics")}>
              <strong>統計分析を開く</strong>
              <span>平均差比較、相関分析、反復推定と検定を行います。</span>
            </button>
          </div>
        </section>
      </main>
    </div>
  );
};

export default Dashboard;
