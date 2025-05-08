package domain

// User はシステム内で扱うユーザーのエンティティです。
type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

// type UserRepository interface {
// 	FindByEmail(ctx context.Context, email string) (*User, error)
// 	Create(ctx context.Context, user *User) (int, error)
// }
