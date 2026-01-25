import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useActions, type ActionLog } from "../hooks/useActions";
import { useTargets } from "../hooks/useTargets";
import styles from "../assets/styles/Actions.module.scss";

// 行動記録一覧ページ
const Actions = () => {
  const navigate = useNavigate();
  const { actions, pagination, loading, error, fetchActions, deleteAction } = useActions();
  const { targets, fetchTargets } = useTargets();

  const [actionTypes, setActionTypes] = useState<{ id: number; action_name: string }[]>([]);

  useEffect(() => {
    fetchActions();
    fetchTargets();
    // 行動種別一覧取得
    fetch(`${import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1"}/action_types`)
      .then((res) => res.json())
      .then((data) => {
        if (data?.success && Array.isArray(data.data?.action_types)) {
          setActionTypes(data.data.action_types);
        }
      })
      .catch(() => {
        setActionTypes([]);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const targetMap = useMemo(() => {
    const map = new Map<number, string>();
    targets.forEach((t) => map.set(t.id, t.name));
    return map;
  }, [targets]);

  const actionTypeMap = useMemo(() => {
    const map = new Map<number, string>();
    actionTypes.forEach((t) => map.set(t.id, t.action_name));
    return map;
  }, [actionTypes]);

  const renderRow = (action: ActionLog) => {
    const targetName =
      action.target_id == null ? "未紐づけ" : targetMap.get(action.target_id) ?? `ID:${action.target_id}`;
    const actionTypeName = actionTypeMap.get(action.action_type) ?? `Type:${action.action_type}`;
    return (
      <tr key={action.id}>
        <td className={styles.idCol}>{action.id}</td>
        <td className={styles.targetCol}>{targetName}</td>
        <td className={styles.actionTypeCol}>{actionTypeName}</td>
        <td className={styles.notesCol}>{action.notes || <span style={{ fontStyle: "italic" }}>メモなし</span>}</td>
        <td className={styles.timestampCol}>{new Date(action.timestamp).toLocaleString("ja-JP")}</td>
        <td className={styles.actionsCol}>
          <button className={styles.successButton} onClick={() => navigate(`/actions/${action.id}/edit`)}>
            編集
          </button>
          <button
            className={styles.dangerButton}
            onClick={() => handleDelete(action.id, action.notes || actionTypeName)}
            disabled={loading}
          >
            削除
          </button>
        </td>
      </tr>
    );
  };

  const handleDelete = async (id: number, label: string) => {
    if (!window.confirm(`「${label}」を削除してもよろしいですか？`)) return;
    const success = await deleteAction(id);
    if (success) alert("行動記録を削除しました");
  };

  if (loading && actions.length === 0) {
    return (
      <div className={styles.container}>
        <p>読み込み中...</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>行動記録</h1>
        <button className={styles.primaryButton} onClick={() => navigate("/actions/new")}>
          新規追加
        </button>
      </div>

      {error && <div className={styles.errorBox}>エラー: {error}</div>}

      <div className={styles.info}>
        全 {pagination.total} 件中 {actions.length} 件を表示
      </div>

      {actions.length === 0 ? (
        <div className={styles.empty}>
          <p>行動記録が登録されていません</p>
          <button className={styles.primaryButton} onClick={() => navigate("/actions/new")}>
            最初の行動を追加
          </button>
        </div>
      ) : (
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.idCol}>ID</th>
                <th className={styles.targetCol}>観察対象</th>
                <th className={styles.actionTypeCol}>行動種別</th>
                <th>メモ</th>
                <th className={styles.timestampCol}>発生日</th>
                <th className={styles.actionsHead}>操作</th>
              </tr>
            </thead>
            <tbody>{actions.map(renderRow)}</tbody>
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

export default Actions;
