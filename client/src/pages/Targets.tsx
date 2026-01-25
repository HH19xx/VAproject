import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useTargets } from "../hooks/useTargets";
import styles from "../assets/styles/Targets.module.scss";

// 観察対象一覧ページ
const Targets = () => {
  const navigate = useNavigate();
  const { targets, pagination, loading, error, fetchTargets, deleteTarget } = useTargets();

  useEffect(() => {
    fetchTargets();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleDelete = async (id: number, name: string) => {
    if (!window.confirm(`「${name}」を削除してもよろしいですか？`)) return;
    const success = await deleteTarget(id);
    if (success) alert("観察対象を削除しました");
  };

  if (loading && targets.length === 0) {
    return (
      <div className={styles.container}>
        <p>読み込み中...</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>観察対象管理</h1>
        <button className={styles.primaryButton} onClick={() => navigate("/targets/new")}>
          新規追加
        </button>
      </div>

      {error && <div className={styles.errorBox}>エラー: {error}</div>}

      <div className={styles.info}>
        全 {pagination.total} 件中 {targets.length} 件を表示
      </div>

      {targets.length === 0 ? (
        <div className={styles.empty}>
          <p>観察対象が登録されていません</p>
          <button className={styles.primaryButton} onClick={() => navigate("/targets/new")}>
            最初の観察対象を追加
          </button>
        </div>
      ) : (
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th style={{ width: "80px" }}>ID</th>
                <th style={{ width: "220px" }}>名前</th>
                <th>説明</th>
                <th style={{ width: "170px" }}>作成日時</th>
                <th style={{ width: "200px", textAlign: "center" }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {targets.map((target) => (
                <tr key={target.id}>
                  <td>{target.id}</td>
                  <td style={{ fontWeight: 700 }}>{target.name}</td>
                  <td style={{ color: "#666" }}>{target.description || <span style={{ fontStyle: "italic" }}>説明なし</span>}</td>
                  <td style={{ fontSize: "13px", color: "#666" }}>{new Date(target.created_at).toLocaleString("ja-JP")}</td>
                  <td className={styles.actionsCell}>
                    <button className={styles.successButton} onClick={() => navigate(`/targets/${target.id}/edit`)}>
                      編集
                    </button>
                    <button className={styles.dangerButton} onClick={() => handleDelete(target.id, target.name)} disabled={loading}>
                      削除
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className={styles.footer}>
        <button className={styles.backButton} onClick={() => navigate("/dashboard")}>
          ダッシュボードに戻る
        </button>
      </div>
    </div>
  );
};

export default Targets;
