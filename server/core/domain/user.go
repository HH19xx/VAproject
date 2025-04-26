package domain

// User はシステムユーザーを表すエンティティです。
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// パスワードフィールドはセキュリティ上の理由からエンティティに含めないか、
	// ハッシュ化された形式で別途管理することを検討します。
	// ここでは簡略化のため含めませんが、実際のアプリケーションでは考慮が必要です。
}
