import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useTargets } from "../hooks/useTargets";

// 観察対象一覧ページ
const Targets = () => {
  const navigate = useNavigate();
  const { targets, pagination, loading, error, fetchTargets, deleteTarget } = useTargets();

  // コンポーネントマウント時に観察対象一覧を取得
  useEffect(() => {
    fetchTargets();
  }, []);

  // 削除ボタンのクリックハンドラ
  const handleDelete = async (id: number, name: string) => {
    if (!window.confirm(`「${name}」を削除してもよろしいですか？`)) {
      return;
    }

    const success = await deleteTarget(id);
    if (success) {
      alert("観察対象を削除しました");
    }
  };

  // ローディング中の表示
  if (loading && targets.length === 0) {
    return (
      <div style={{ padding: "20px" }}>
        <p>読み込み中...</p>
      </div>
    );
  }

  return (
    <div style={{ padding: "20px", maxWidth: "1200px", margin: "0 auto" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
        <h1>観察対象管理</h1>
        <button
          onClick={() => navigate("/targets/new")}
          style={{
            padding: "10px 20px",
            backgroundColor: "#007bff",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer",
            fontSize: "14px",
          }}
        >
          新規追加
        </button>
      </div>

      {error && (
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
          エラー: {error}
        </div>
      )}

      <div style={{ marginBottom: "20px" }}>
        <p style={{ color: "#666" }}>
          全 {pagination.total} 件中 {targets.length} 件を表示
        </p>
      </div>

      {targets.length === 0 ? (
        <div
          style={{
            padding: "40px",
            textAlign: "center",
            backgroundColor: "#f8f9fa",
            border: "1px solid #dee2e6",
            borderRadius: "4px",
          }}
        >
          <p style={{ color: "#666" }}>観察対象が登録されていません</p>
          <button
            onClick={() => navigate("/targets/new")}
            style={{
              marginTop: "10px",
              padding: "10px 20px",
              backgroundColor: "#007bff",
              color: "white",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            最初の観察対象を追加
          </button>
        </div>
      ) : (
        <div style={{ overflowX: "auto" }}>
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              backgroundColor: "white",
              boxShadow: "0 1px 3px rgba(0,0,0,0.1)",
            }}
          >
            <thead>
              <tr style={{ backgroundColor: "#f8f9fa", borderBottom: "2px solid #dee2e6" }}>
                <th style={{ padding: "12px", textAlign: "left", width: "100px" }}>ID</th>
                <th style={{ padding: "12px", textAlign: "left", width: "200px" }}>名前</th>
                <th style={{ padding: "12px", textAlign: "left" }}>説明</th>
                <th style={{ padding: "12px", textAlign: "left", width: "150px" }}>作成日時</th>
                <th style={{ padding: "12px", textAlign: "center", width: "200px" }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {targets.map((target) => (
                <tr key={target.id} style={{ borderBottom: "1px solid #dee2e6" }}>
                  <td style={{ padding: "12px" }}>{target.id}</td>
                  <td style={{ padding: "12px", fontWeight: "bold" }}>{target.name}</td>
                  <td style={{ padding: "12px", color: "#666" }}>
                    {target.description || <span style={{ fontStyle: "italic" }}>説明なし</span>}
                  </td>
                  <td style={{ padding: "12px", fontSize: "13px", color: "#666" }}>
                    {new Date(target.created_at).toLocaleString("ja-JP")}
                  </td>
                  <td style={{ padding: "12px", textAlign: "center" }}>
                    <button
                      onClick={() => navigate(`/targets/${target.id}/edit`)}
                      style={{
                        marginRight: "8px",
                        padding: "6px 12px",
                        backgroundColor: "#28a745",
                        color: "white",
                        border: "none",
                        borderRadius: "4px",
                        cursor: "pointer",
                        fontSize: "13px",
                      }}
                    >
                      編集
                    </button>
                    <button
                      onClick={() => handleDelete(target.id, target.name)}
                      style={{
                        padding: "6px 12px",
                        backgroundColor: "#dc3545",
                        color: "white",
                        border: "none",
                        borderRadius: "4px",
                        cursor: "pointer",
                        fontSize: "13px",
                      }}
                      disabled={loading}
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

      <div style={{ marginTop: "20px", textAlign: "center" }}>
        <button
          onClick={() => navigate("/dashboard")}
          style={{
            padding: "10px 20px",
            backgroundColor: "#6c757d",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: "pointer",
          }}
        >
          ダッシュボードに戻る
        </button>
      </div>
    </div>
  );
};

export default Targets;
