import { useAuth } from './contexts/AuthContext';
import { useFetchMessages } from './hooks/useFetchMessages';
import LoginForm from './components/LoginForm';
import { useEffect } from 'react';

function App() {
  const { userID, logout, loading } = useAuth();
  const { message } = useFetchMessages();

  // URLパラメータからアクセストークンとリフレッシュトークンを取得し保存します。
  // Google OAuthコールバック後は /?token=...&refresh=...&expires=... の形式で遷移します。
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const token = params.get('token');
    const refresh = params.get('refresh');
    if (token && refresh) {
      localStorage.setItem('token', token);
      localStorage.setItem('refresh_token', refresh);
      // URLパラメータをクリアしてリロード
      window.history.replaceState({}, document.title, window.location.pathname);
      window.location.reload();
    }
  }, []);

  if (loading) return <p>読み込み中...</p>;

  return (
    <div>
                <p>メッセージ: {message ?? "読み込み中"}</p>
      <h1>仮説フロントエンド</h1>
      {userID ? (
        <>
          <p>ようこそ、ユーザーID {userID} さん</p>
          <button onClick={logout}>ログアウト</button>
        </>
      ) : (
        <LoginForm />
      )}
    </div>
  );
}

export default App;
