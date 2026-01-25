import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useActionTypes } from "../hooks/useActionTypes";
import styles from "../assets/styles/Targets.module.scss";

// 行動種別一覧ページ
const ActionTypes = () => {
  const navigate = useNavigate();
  const { actionTypes, loading, error, fetchActionTypes, deleteActionType } = useActionTypes();

  useEffect(() => {
    fetchActionTypes();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleDelete = async (id: number, name: string) => {
    if (!window.confirm(`「${name}」を削除しますか？`)) return;

    const success = await deleteActionType(id);
    if (success) {
      alert("行動種別を削除しました");
    }
  };

  if (loading && actionTypes.length === 0) {
    return (
      <div className={styles.container}>
        <p>読み込み中...</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>行動種別管理</h1>
        <button className={styles.primaryButton} onClick={() => navigate("/action-types/new")}>
          新規作成
        </button>
      </div>

      {error && <div className={styles.errorBox}>エラー: {error}</div>}

      <p className={styles.info}>登録件数: {actionTypes.length}件</p>

      {actionTypes.length === 0 ? (
        <div className={styles.empty}>
          <p>行動種別が登録されていません。</p>
          <p>「新規作成」ボタンから行動種別を追加してください。</p>
        </div>
      ) : (
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>ID</th>
                <th>種別名</th>
                <th>説明</th>
                <th>作成日時</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {actionTypes.map((actionType) => (
                <tr key={actionType.id}>
                  <td>{actionType.id}</td>
                  <td>{actionType.action_name}</td>
                  <td>{actionType.description || "-"}</td>
                  <td>{new Date(actionType.created_at).toLocaleString("ja-JP")}</td>
                  <td className={styles.actionsCell}>
                    <button
                      className={styles.successButton}
                      onClick={() => navigate(`/action-types/${actionType.id}/edit`)}
                    >
                      編集
                    </button>
                    <button
                      className={styles.dangerButton}
                      onClick={() => handleDelete(actionType.id, actionType.action_name)}
                    >
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

export default ActionTypes;
