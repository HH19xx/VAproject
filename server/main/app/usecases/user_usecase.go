// server/main/app/usecases/user_usecase.go
package usecases

import (
	"context"
	"errors"
	"server/main/domain"
)

// 認証失敗時のエラー
var ErrAuthFailed = errors.New("認証失敗")

// UserRepository はユーザーに関する永続化の抽象インターフェースです。
// main/infra/repository でこのinterfaceを実装します。
type UserRepository interface {
	FindByName(ctx context.Context, name string) (*domain.User, error)

	// OAuth用：メールアドレスからユーザーを探す／作成する
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (int, error)
}

// UserUsecase はユーザーに関するユースケースを提供します。
type UserUsecase struct {
	Repo UserRepository
}

// Login はユーザー名とパスワードを用いて認証を行い、成功すればユーザーIDを返します。
func (u *UserUsecase) Login(ctx context.Context, name, password string) (int, error) {
	user, err := u.Repo.FindByName(ctx, name)
	if err != nil {
		return 0, ErrAuthFailed
	}

	// 現時点では平文比較。今後は bcrypt 等に置き換え予定
	if user.Password != password {
		return 0, ErrAuthFailed
	}

	return user.ID, nil
}

// FindOrCreateUserByEmail はOAuthで取得したメールアドレスからユーザーを照会し、なければ作成します。
func (u *UserUsecase) FindOrCreateUserByEmail(ctx context.Context, email string) (int, error) {
	user, err := u.Repo.FindByEmail(ctx, email)
	if err == nil {
		return user.ID, nil
	}

	newUser := &domain.User{
		Name: email, // NameとEmailの設計を分けていない場合、NameにEmailを流用してもよい
	}
	return u.Repo.Create(ctx, newUser)
}
