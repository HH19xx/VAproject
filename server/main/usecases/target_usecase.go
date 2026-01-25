package usecases

import (
	"context"
	"errors"

	"github.com/lib/pq"

	"server/main/domain"
)

// 観察対象の永続化を抽象化したインターフェース
type TargetRepository interface {
	FindAll(ctx context.Context, limit, offset int) ([]*domain.Target, error)
	FindByID(ctx context.Context, id int) (*domain.Target, error)
	Create(ctx context.Context, target *domain.Target) (int, error)
	Update(ctx context.Context, target *domain.Target) error
	Delete(ctx context.Context, id int) error
	Count(ctx context.Context) (int, error)
}

// 観察対象ユースケース
type TargetUsecase struct {
	Repo TargetRepository
}

// バリデーション・ドメインエラー
var (
	ErrTargetNameRequired  = errors.New("観察対象の名前は必須です")
	ErrTargetNameTooLong   = errors.New("観察対象の名前は64文字以内で入力してください")
	ErrTargetNameDuplicate = errors.New("観察対象の名前が重複しています")
)

// ページネーション付きで観察対象一覧を取得
// page: 1始まりのページ番号, limit: 1ページあたりの件数
func (u *TargetUsecase) GetTargets(ctx context.Context, page, limit int) ([]*domain.Target, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	targets, err := u.Repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.Repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return targets, total, nil
}

// ID指定で観察対象を取得
func (u *TargetUsecase) GetTargetByID(ctx context.Context, id int) (*domain.Target, error) {
	return u.Repo.FindByID(ctx, id)
}

// 観察対象を新規登録
func (u *TargetUsecase) CreateTarget(ctx context.Context, name, description, createUser string) (*domain.Target, error) {
	if name == "" {
		return nil, ErrTargetNameRequired
	}
	if len(name) > 64 {
		return nil, ErrTargetNameTooLong
	}

	target := &domain.Target{
		Name:        name,
		Description: description,
		CreateUser:  createUser,
		UpdateUser:  createUser,
	}

	id, err := u.Repo.Create(ctx, target)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrTargetNameDuplicate
		}
		return nil, err
	}

	return u.Repo.FindByID(ctx, id)
}

// 観察対象を更新
func (u *TargetUsecase) UpdateTarget(ctx context.Context, id int, name, description, updateUser string) (*domain.Target, error) {
	if name == "" {
		return nil, ErrTargetNameRequired
	}
	if len(name) > 64 {
		return nil, ErrTargetNameTooLong
	}

	target, err := u.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	target.Name = name
	target.Description = description
	target.UpdateUser = updateUser

	if err := u.Repo.Update(ctx, target); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrTargetNameDuplicate
		}
		return nil, err
	}

	return u.Repo.FindByID(ctx, id)
}

// 観察対象を論理削除
func (u *TargetUsecase) DeleteTarget(ctx context.Context, id int) error {
	return u.Repo.Delete(ctx, id)
}
