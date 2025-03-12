import useFetchMessages from './hooks/useFetchMassages'

function App() {
  const message = useFetchMessages()

  return (
    <div>
      <h1>仮説フロントエンド</h1>
      <p>メッセージ: {message ?? "読み込み中"}</p>
    </div>
  )
}

export default App
