package domain

// ユーザーの認証プロバイダー情報を表すエンティティ
type UserAuthProvider struct {
	ID             int
	UserID         int
	Provider       string // 'local', 'google', 'github' 等
	ProviderUserID string // OAuth側のユーザーID（localの場合は空文字）
}
