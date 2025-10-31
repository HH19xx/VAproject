import { useAuth } from './contexts/AuthContext';
import useFetchMessages from './hooks/useFetchMessages';
import LoginForm from './components/LoginForm';

function App() {
  const { userID, logout, loading } = useAuth();
  const message = useFetchMessages();

  if (loading) return <p>読み込み中...</p>;

  return (
    <div>
      <h1>仮説フロントエンド</h1>
      {userID ? (
        <>
          <p>ようこそ、ユーザーID {userID} さん</p>
          <button onClick={logout}>ログアウト</button>
          <p>メッセージ: {message ?? "読み込み中"}</p>
        </>
      ) : (
        <LoginForm />
      )}
    </div>
  );
}

export default App;
