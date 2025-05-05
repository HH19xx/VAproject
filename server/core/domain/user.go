package domain

// User はシステム内で扱うユーザーのエンティティです。
type User struct {
	ID       int
	Name     string
	Password string // 平文ではなく、今後はハッシュ前提で
}

