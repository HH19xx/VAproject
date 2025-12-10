package usecases

import (
	"context"
	"errors"
	"server/main/domain"
)

// 観察対象に関する永続化の抽象インターフェース
type TargetRepository interface {
	FindAll(ctx context.Context, limit, offset int) ([]*domain.Target, error)
	FindByID(ctx context.Context, id int) (*domain.Target, error)
	Create(ctx context.Context, target *domain.Target) (int, error)
	Update(ctx context.Context, target *domain.Target) error
	Delete(ctx context.Context, id int) error
	Count(ctx context.Context) (int, error)
}

// 観察対象に関するビジネスロジックを提供する
type TargetUsecase struct {
	Repo TargetRepository
}

// バリデーションエラー
var (
	ErrTargetNameRequired = errors.New("観察対象の名前は必須です")
	ErrTargetNameTooLong  = errors.New("観察対象の名前は64文字以内で入力してください")
)

// ページネーション付きで観察対象の一覧を取得する
// page: ページ番号（1から開始）、limit: 1ページあたりの件数
// 戻り値: 観察対象のスライス、総件数、エラー
func (u *TargetUsecase) GetTargets(ctx context.Context, page, limit int) ([]*domain.Target, int, error) {
	// ページ番号とリミットの妥当性チェック
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// オフセット計算（ページ1なら0件スキップ、ページ2なら20件スキップ）
	offset := (page - 1) * limit

	// データベースから観察対象を取得
	targets, err := u.Repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// 総件数を取得（ページネーション情報用）
	total, err := u.Repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return targets, total, nil
}

// 指定されたIDの観察対象を取得する
func (u *TargetUsecase) GetTargetByID(ctx context.Context, id int) (*domain.Target, error) {
	return u.Repo.FindByID(ctx, id)
}

// 新しい観察対象を作成する
// name: 観察対象の名前（必須、1-64文字）
// description: 観察対象の説明（任意）
// createUser: 作成者ユーザー名
// 戻り値: 作成された観察対象、エラー
func (u *TargetUsecase) CreateTarget(ctx context.Context, name, description, createUser string) (*domain.Target, error) {
	// バリデーション: 名前は必須
	if name == "" {
		return nil, ErrTargetNameRequired
	}
	// バリデーション: 名前は64文字以内
	if len(name) > 64 {
		return nil, ErrTargetNameTooLong
	}

	// Targetエンティティを組み立て
	target := &domain.Target{
		Name:        name,
		Description: description,
		CreateUser:  createUser,
		UpdateUser:  createUser,
	}

	// データベースに保存し、生成されたIDを取得
	id, err := u.Repo.Create(ctx, target)
	if err != nil {
		return nil, err
	}

	// 作成された観察対象を取得して返す
	return u.Repo.FindByID(ctx, id)
}

// 更新された観察対象を取得して返す
// id: 更新対象のID
// name: 新しい名前（必須、1-64文字）
// description: 新しい説明（任意）
// updateUser: 更新者ユーザー名
// 戻り値: 更新された観察対象、エラー
func (u *TargetUsecase) UpdateTarget(ctx context.Context, id int, name, description, updateUser string) (*domain.Target, error) {
	// バリデーション: 名前は必須
	if name == "" {
		return nil, ErrTargetNameRequired
	}
	// バリデーション: 名前は64文字以内
	if len(name) > 64 {
		return nil, ErrTargetNameTooLong
	}

	// 更新対象が存在するか確認
	target, err := u.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新内容を設定
	target.Name = name
	target.Description = description
	target.UpdateUser = updateUser

	// データベースを更新
	if err := u.Repo.Update(ctx, target); err != nil {
		return nil, err
	}

	// 更新後の観察対象を取得して返す
	return u.Repo.FindByID(ctx, id)
}

// 指定されたIDの観察対象を削除する
func (u *TargetUsecase) DeleteTarget(ctx context.Context, id int) error {
	return u.Repo.Delete(ctx, id)
}
