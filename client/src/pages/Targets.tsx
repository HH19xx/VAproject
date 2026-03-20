import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTargets, type ActionLog } from "../hooks/useTargets";
import { useTags } from "../hooks/useTags";
import styles from "../assets/styles/Targets.module.scss";

const Targets = () => {
  const navigate = useNavigate();
  const { targets, pagination, loading, error, fetchTargets, deleteTarget, fetchActionLogsByTarget } = useTargets();
  const { tags, fetchTags } = useTags();

  const [selectedTargetID, setSelectedTargetID] = useState<number | null>(null);
  const [matchedLogs, setMatchedLogs] = useState<ActionLog[]>([]);
  const [matchedTotal, setMatchedTotal] = useState(0);
  const [matchedLoading, setMatchedLoading] = useState(false);

  useEffect(() => {
    fetchTargets();
    fetchTags();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const tagNameMap = useMemo(() => {
    const map = new Map<number, string>();
    tags.forEach((tag) => map.set(tag.id, tag.name));
    return map;
  }, [tags]);

  const handleDelete = async (id: number, name: string) => {
    if (!window.confirm(`「${name}」を削除しますか？`)) return;
    const success = await deleteTarget(id);
    if (!success) return;

    if (selectedTargetID === id) {
      setSelectedTargetID(null);
      setMatchedLogs([]);
      setMatchedTotal(0);
    }
  };

  const handleShowLogs = async (targetID: number) => {
    setSelectedTargetID(targetID);
    setMatchedLoading(true);
    const result = await fetchActionLogsByTarget(targetID, 1, 20);
    setMatchedLoading(false);
    if (!result) {
      setMatchedLogs([]);
      setMatchedTotal(0);
      return;
    }
    setMatchedLogs(result.logs);
    setMatchedTotal(result.pagination.total);
  };

  const formatTagNames = (tagIDs: number[]) =>
    tagIDs.map((id) => tagNameMap.get(id) || `#${id}`).join(", ");

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>観察対象一覧</h1>
        <button className={styles.primaryButton} onClick={() => navigate("/targets/new")}>
          観察対象を作成
        </button>
      </div>

      {error && <div className={styles.errorBox}>エラー: {error}</div>}

      <div className={styles.info}>
        合計 {pagination.total} 件 / 現在 {targets.length} 件表示
      </div>

      {targets.length === 0 ? (
        <div className={styles.empty}>
          <p>観察対象がありません。</p>
          <button className={styles.primaryButton} onClick={() => navigate("/targets/new")}>
            最初の観察対象を作成
          </button>
        </div>
      ) : (
        <div className={styles.tableWrapper}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th style={{ width: "80px" }}>ID</th>
                <th style={{ width: "180px" }}>名前</th>
                <th style={{ width: "130px" }}>match_mode</th>
                <th>タグ条件（AND）</th>
                <th style={{ width: "220px", textAlign: "center" }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {targets.map((target) => (
                <tr key={target.id}>
                  <td>{target.id}</td>
                  <td style={{ fontWeight: 700 }}>{target.name}</td>
                  <td>{target.match_mode}</td>
                  <td>{formatTagNames(target.tag_ids)}</td>
                  <td className={styles.actionsCell}>
                    <button className={styles.successButton} onClick={() => navigate(`/targets/${target.id}/edit`)}>
                      編集
                    </button>
                    <button className={styles.secondaryButton} onClick={() => handleShowLogs(target.id)} disabled={loading}>
                      一致ログ
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

      {selectedTargetID && (
        <div className={styles.logPanel}>
          <h2 className={styles.logPanelTitle}>対象ID {selectedTargetID} の一致ログ</h2>
          {matchedLoading ? (
            <p>読み込み中...</p>
          ) : matchedLogs.length === 0 ? (
            <p>一致するログがありません。</p>
          ) : (
            <>
              <p className={styles.info}>合計 {matchedTotal} 件（上位 20 件）</p>
              <ul className={styles.logList}>
                {matchedLogs.map((log) => (
                  <li key={log.id} className={styles.logItem}>
                    <div>
                      <strong>#{log.id}</strong> {new Date(log.occurred_at).toLocaleString("ja-JP")}
                    </div>
                    <div className={styles.logTags}>タグ: {formatTagNames(log.tag_ids)}</div>
                    {log.notes && <div className={styles.logNote}>メモ: {log.notes}</div>}
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}

      <div className={styles.footer}>
        <button className={styles.backButton} onClick={() => navigate("/dashboard")}>
          ダッシュボードへ戻る
        </button>
      </div>
    </div>
  );
};

export default Targets;
