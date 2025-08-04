package domain

// User はシステム内で扱うユーザーのエンティティです。
type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}
