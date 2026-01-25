import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useActions } from "../hooks/useActions";
import { useTargets } from "../hooks/useTargets";
import styles from "../assets/styles/ActionForm.module.scss";

// 行動記録の追加・編集フォーム
const ActionForm = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditMode = !!id;

  const { loading, error, fetchActionByID, createAction, updateAction } = useActions();
  const { targets, fetchTargets } = useTargets();
  const [actionTypes, setActionTypes] = useState<{ id: number; action_name: string }[]>([]);

  const [targetId, setTargetId] = useState<number | null>(null);
  const [actionTypeId, setActionTypeId] = useState<number>(0);
  const [timestamp, setTimestamp] = useState<string>("");
  const [notes, setNotes] = useState("");
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // セレクト用に target / action_types をロード
  useEffect(() => {
    fetchTargets();
    fetch(`${import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1"}/action_types`)
      .then((res) => res.json())
      .then((data) => {
        if (data?.success && Array.isArray(data.data?.action_types)) {
          setActionTypes(data.data.action_types);
        }
      })
      .catch(() => setActionTypes([]));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 編集時は既存データを取得
  useEffect(() => {
    if (isEditMode && id) {
      const load = async () => {
        const action = await fetchActionByID(parseInt(id, 10));
        if (action) {
          setTargetId(action.target_id);
          setActionTypeId(action.action_type);
          setTimestamp(action.timestamp.slice(0, 16)); // datetime-local 用に整形
          setNotes(action.notes);
        } else {
          setFormError("行動記録が見つかりません");
        }
      };
      load();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, isEditMode]);

  const targetOptions = useMemo(() => targets, [targets]);
  const actionTypeOptions = useMemo(() => actionTypes, [actionTypes]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    if (!targetId || !actionTypeId || !timestamp) {
      setFormError("対象・種別・日時は必須です");
      return;
    }

    setSubmitting(true);
    try {
      if (isEditMode && id) {
        const result = await updateAction(
          parseInt(id, 10),
          targetId,
          actionTypeId,
          new Date(timestamp).toISOString(),
          notes.trim()
        );
        if (result) {
          alert("行動記録を更新しました");
          navigate("/actions");
        } else {
          setFormError("更新に失敗しました");
        }
      } else {
        const result = await createAction(
          targetId,
          actionTypeId,
          new Date(timestamp).toISOString(),
          notes.trim()
        );
        if (result) {
          alert("行動記録を作成しました");
          navigate("/actions");
        } else {
          setFormError("作成に失敗しました");
        }
      }
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "エラーが発生しました");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading && isEditMode) {
    return (
      <div className={styles.loading}>
        <p>読み込み中...</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>{isEditMode ? "行動記録を編集" : "行動記録を追加"}</h1>

      {(formError || error) && <div className={styles.errorBox}>エラー: {formError || error}</div>}

      <form onSubmit={handleSubmit}>
        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="target">
            観察対象 <span className={styles.required}>*</span>
          </label>
          <select
            id="target"
            className={styles.select}
            value={targetId}
            onChange={(e) => setTargetId(parseInt(e.target.value, 10))}
            disabled={submitting}
          >
            <option value={0}>選択してください</option>
            {targetOptions.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name}
              </option>
            ))}
          </select>
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="actionType">
            行動種別 <span className={styles.required}>*</span>
          </label>
          <select
            id="actionType"
            className={styles.select}
            value={actionTypeId}
            onChange={(e) => setActionTypeId(parseInt(e.target.value, 10))}
            disabled={submitting}
          >
            <option value={0}>選択してください</option>
            {actionTypeOptions.map((t) => (
              <option key={t.id} value={t.id}>
                {t.action_name}
              </option>
            ))}
          </select>
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="timestamp">
            実施日時 <span className={styles.required}>*</span>
          </label>
          <input
            id="timestamp"
            className={styles.input}
            type="datetime-local"
            value={timestamp}
            onChange={(e) => setTimestamp(e.target.value)}
            disabled={submitting}
          />
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="notes">
            メモ
          </label>
          <textarea
            id="notes"
            className={styles.textarea}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={5}
            placeholder="例）ユーザーからのヒアリング内容など"
            disabled={submitting}
          />
        </div>

        <div className={styles.actions}>
          <button className={styles.primaryButton} type="submit" disabled={submitting}>
            {submitting ? "処理中..." : isEditMode ? "更新" : "作成"}
          </button>
          <button className={styles.secondaryButton} type="button" onClick={() => navigate("/actions")} disabled={submitting}>
            キャンセル
          </button>
        </div>
      </form>
    </div>
  );
};

export default ActionForm;
