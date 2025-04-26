package usecases

import (
	"errors"
	"server/core/domain"
	"server/infra/repository"
)

// UserUseCase はユーザー関連のユースケースを定義します。
type UserUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase は新しいUserUseCaseのインスタンスを作成します。
func NewUserUseCase(userRepo repository.UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

// RegisterUser は新しいユーザーを登録します。
func (uc *UserUseCase) RegisterUser(user *domain.User) error {
	// ここでユーザー情報の検証などビジネスロジックを実装します。
	// 例: メールアドレスの形式チェック、必須フィールドの確認など
	if user.Name == "" || user.Email == "" {
		return errors.New("ユーザー名とメールアドレスは必須です")
	}

	// リポジトリを介してユーザー情報をデータベースに保存します。
	err := uc.userRepo.Save(user)
	if err != nil {
		return errors.New("ユーザーの保存に失敗しました: " + err.Error())
	}

	return nil
}
