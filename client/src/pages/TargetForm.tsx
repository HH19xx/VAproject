import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useTargets } from "../hooks/useTargets";

// 観察対象の追加・編集フォームページ
const TargetForm = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditMode = !!id;

  const { loading, error, fetchTargetByID, createTarget, updateTarget } = useTargets();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // 編集モードの場合、既存データを取得
  useEffect(() => {
    if (isEditMode && id) {
      const loadTarget = async () => {
        const target = await fetchTargetByID(parseInt(id, 10));
        if (target) {
          setName(target.name);
          setDescription(target.description);
        } else {
          setFormError("観察対象が見つかりません");
        }
      };
      loadTarget();
    }
  }, [id, isEditMode]);

  // フォーム送信ハンドラ
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError(null);

    // バリデーション
    if (!name.trim()) {
      setFormError("名前は必須です");
      return;
    }

    if (name.length > 64) {
      setFormError("名前は64文字以内で入力してください");
      return;
    }

    setSubmitting(true);

    try {
      if (isEditMode && id) {
        // 更新
        const result = await updateTarget(parseInt(id, 10), name.trim(), description.trim());
        if (result) {
          alert("観察対象を更新しました");
          navigate("/targets");
        } else {
          setFormError("更新に失敗しました");
        }
      } else {
        // 新規作成
        const result = await createTarget(name.trim(), description.trim());
        if (result) {
          alert("観察対象を作成しました");
          navigate("/targets");
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

  // ローディング中の表示
  if (loading && isEditMode) {
    return (
      <div style={{ padding: "20px" }}>
        <p>読み込み中...</p>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "600px", margin: "0 auto" }}>
      <h1>{isEditMode ? "観察対象を編集" : "観察対象を追加"}</h1>

      {(formError || error) && (
        <div
          style={{
            padding: "10px",
            marginBottom: "20px",
            backgroundColor: "#f8d7da",
            color: "#721c24",
            border: "1px solid #f5c6cb",
            borderRadius: "4px",
          }}
        >
          エラー: {formError || error}
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "20px" }}>
          <label
            htmlFor="name"
            style={{
              display: "block",
              marginBottom: "8px",
              fontWeight: "bold",
              color: "#333",
            }}
          >
            名前 <span style={{ color: "red" }}>*</span>
          </label>
          <input
            type="text"
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            maxLength={64}
            required
            style={{
              width: "100%",
              padding: "10px",
              fontSize: "14px",
              border: "1px solid #ced4da",
              borderRadius: "4px",
              boxSizing: "border-box",
            }}
            placeholder="例: 田中太郎"
            disabled={submitting}
          />
          <small style={{ color: "#666", fontSize: "12px" }}>
            {name.length}/64 文字
          </small>
        </div>

        <div style={{ marginBottom: "20px" }}>
          <label
            htmlFor="description"
            style={{
              display: "block",
              marginBottom: "8px",
              fontWeight: "bold",
              color: "#333",
            }}
          >
            説明
          </label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={5}
            style={{
              width: "100%",
              padding: "10px",
              fontSize: "14px",
              border: "1px solid #ced4da",
              borderRadius: "4px",
              boxSizing: "border-box",
              resize: "vertical",
            }}
            placeholder="例: 営業部所属、2020年入社"
            disabled={submitting}
          />
        </div>

        <div style={{ display: "flex", gap: "10px" }}>
          <button
            type="submit"
            disabled={submitting}
            style={{
              flex: 1,
              padding: "12px",
              backgroundColor: submitting ? "#6c757d" : "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: submitting ? "not-allowed" : "pointer",
              fontSize: "16px",
              fontWeight: "bold",
            }}
          >
            {submitting ? "処理中..." : isEditMode ? "更新" : "作成"}
          </button>

          <button
            type="button"
            onClick={() => navigate("/targets")}
            disabled={submitting}
            style={{
              flex: 1,
              padding: "12px",
              backgroundColor: "#6c757d",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: submitting ? "not-allowed" : "pointer",
              fontSize: "16px",
            }}
          >
            キャンセル
          </button>
        </div>
      </form>
    </div>
  );
};

export default TargetForm;
