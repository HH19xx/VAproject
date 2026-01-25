import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useActionTypes } from "../hooks/useActionTypes";
import styles from "../assets/styles/TargetForm.module.scss";

// 行動種別の追加・編集フォーム
const ActionTypeForm = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditMode = !!id;

  const { loading, error, fetchActionTypeByID, createActionType, updateActionType } = useActionTypes();

  const [actionName, setActionName] = useState("");
  const [description, setDescription] = useState("");
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // 編集時は既存データを取得
  useEffect(() => {
    if (isEditMode && id) {
      const loadActionType = async () => {
        const actionType = await fetchActionTypeByID(parseInt(id, 10));
        if (actionType) {
          setActionName(actionType.action_name);
          setDescription(actionType.description || "");
        } else {
          setFormError("行動種別が見つかりません");
        }
      };
      loadActionType();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, isEditMode]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    const trimmedName = actionName.trim();
    if (!trimmedName) {
      setFormError("種別名は必須です");
      return;
    }
    if (trimmedName.length > 64) {
      setFormError("種別名は64文字以内で入力してください");
      return;
    }

    setSubmitting(true);

    try {
      if (isEditMode && id) {
        const result = await updateActionType(parseInt(id, 10), trimmedName, description.trim());
        if (result) {
          alert("行動種別を更新しました");
          navigate("/action-types");
        } else {
          setFormError("更新に失敗しました");
        }
      } else {
        const result = await createActionType(trimmedName, description.trim());
        if (result) {
          alert("行動種別を作成しました");
          navigate("/action-types");
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
      <div className={styles.container}>
        <p>読み込み中...</p>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>{isEditMode ? "行動種別を編集" : "行動種別を追加"}</h1>

      {(formError || error) && <div className={styles.errorBox}>エラー: {formError || error}</div>}

      <form onSubmit={handleSubmit}>
        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="actionName">
            種別名 <span className={styles.required}>*</span>
          </label>
          <input
            id="actionName"
            className={styles.input}
            type="text"
            value={actionName}
            onChange={(e) => setActionName(e.target.value)}
            placeholder="例）食事、運動、睡眠"
            maxLength={64}
            disabled={submitting}
          />
        </div>

        <div className={styles.formGroup}>
          <label className={styles.label} htmlFor="description">
            説明
          </label>
          <textarea
            id="description"
            className={styles.textarea}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="例）この行動種別の詳細な説明"
            rows={5}
            disabled={submitting}
          />
        </div>

        <div className={styles.actions}>
          <button className={styles.primaryButton} type="submit" disabled={submitting}>
            {submitting ? "処理中..." : isEditMode ? "更新" : "作成"}
          </button>
          <button
            className={styles.secondaryButton}
            type="button"
            onClick={() => navigate("/action-types")}
            disabled={submitting}
          >
            キャンセル
          </button>
        </div>
      </form>
    </div>
  );
};

export default ActionTypeForm;
