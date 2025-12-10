package domain

// User はシステム内で扱うユーザーのエンティティ
type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}
