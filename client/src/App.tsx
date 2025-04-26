import useFetchMessages from './hooks/useFetchMassages'
import RegistrationForm from './components/RegistrationForm';

function App() {
  const message = useFetchMessages()

  return (
    <div>
      <h1>仮説フロントエンド</h1>
      <p>メッセージ: {message ?? "読み込み中"}</p>
      <hr /> {/* 区切り線を追加 */}
      <RegistrationForm /> {/* ユーザー登録フォームを追加 */}
    </div>
  )
}

export default App
