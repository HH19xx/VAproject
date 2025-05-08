package domain

import "context"

// User はシステム内で扱うユーザーのエンティティです。
type User struct {
	ID       int
	Name     string
	Password string // 平文ではなく、今後はハッシュ前提で
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) (int, error)
}
