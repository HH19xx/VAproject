package usecases

import (
	"context"
	"errors"
	"server/core/domain"
)

// 認証失敗時のエラー
var ErrAuthFailed = errors.New("認証失敗")

// UserRepository はユーザーに関する永続化の抽象インターフェースです。
// infra/repository でこのinterfaceを実装します。
type UserRepository interface {
	FindByName(ctx context.Context, name string) (*domain.User, error)
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
