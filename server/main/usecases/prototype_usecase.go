package usecases

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"server/main/domain"
)

type PrototypeRepository interface {
	FindAll(ctx context.Context, userID int) ([]*domain.Prototype, error)
	FindByID(ctx context.Context, id, userID int) (*domain.Prototype, error)
	Create(ctx context.Context, prototype *domain.Prototype) (int, error)
	Update(ctx context.Context, prototype *domain.Prototype) error
	Delete(ctx context.Context, id, userID int) error
}

type PrototypeTagRepository interface {
	FindByID(ctx context.Context, id, userID int) (*domain.Tag, error)
}

type PrototypeUsecase struct {
	Repo    PrototypeRepository
	TagRepo PrototypeTagRepository
}

var (
	ErrPrototypeNameRequired       = errors.New("name は必須です")
	ErrPrototypeNameTooLong        = errors.New("name は128文字以内で入力してください")
	ErrPrototypeNameDuplicate      = errors.New("同名のプロトタイプが既に存在します")
	ErrPrototypeTagIDsRequired     = errors.New("tag_ids は1件以上指定してください")
	ErrPrototypeTagIDInvalid       = errors.New("tag_ids には 1 以上の整数IDのみ指定できます")
	ErrPrototypeTagNotFound        = errors.New("指定した tag_id が見つかりません")
	ErrPrototypeParentNotFound     = errors.New("継承元プロトタイプが見つかりません")
	ErrPrototypeParentNotOwned     = errors.New("継承元プロトタイプの参照権限がありません")
	ErrPrototypeParentSelfNotAllow = errors.New("自分自身を継承元には指定できません")
)

func (u *PrototypeUsecase) GetAll(ctx context.Context, userID int) ([]*domain.Prototype, error) {
	return u.Repo.FindAll(ctx, userID)
}

func (u *PrototypeUsecase) GetByID(ctx context.Context, id, userID int) (*domain.Prototype, error) {
	return u.Repo.FindByID(ctx, id, userID)
}

func (u *PrototypeUsecase) Create(ctx context.Context, userID int, name, description, user string, parentPrototypeID *int, tagIDs []int) (*domain.Prototype, error) {
	normalizedTagIDs, err := u.validatePrototypeInput(ctx, userID, 0, name, parentPrototypeID, tagIDs)
	if err != nil {
		return nil, err
	}

	prototype := &domain.Prototype{
		UserID:            userID,
		Name:              name,
		Description:       description,
		ParentPrototypeID: parentPrototypeID,
		TagIDs:            normalizedTagIDs,
		CreateUser:        user,
		UpdateUser:        user,
	}
	id, err := u.Repo.Create(ctx, prototype)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrPrototypeNameDuplicate
		}
		return nil, err
	}
	return u.Repo.FindByID(ctx, id, userID)
}

func (u *PrototypeUsecase) Update(ctx context.Context, userID, id int, name, description, user string, parentPrototypeID *int, tagIDs []int) (*domain.Prototype, error) {
	normalizedTagIDs, err := u.validatePrototypeInput(ctx, userID, id, name, parentPrototypeID, tagIDs)
	if err != nil {
		return nil, err
	}
	prototype, err := u.Repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	prototype.Name = name
	prototype.Description = description
	prototype.ParentPrototypeID = parentPrototypeID
	prototype.TagIDs = normalizedTagIDs
	prototype.UpdateUser = user

	if err := u.Repo.Update(ctx, prototype); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrPrototypeNameDuplicate
		}
		return nil, err
	}
	return u.Repo.FindByID(ctx, id, userID)
}

func (u *PrototypeUsecase) Delete(ctx context.Context, userID, id int) error {
	return u.Repo.Delete(ctx, id, userID)
}

func (u *PrototypeUsecase) validatePrototypeInput(ctx context.Context, userID, currentID int, name string, parentPrototypeID *int, tagIDs []int) ([]int, error) {
	if name == "" {
		return nil, ErrPrototypeNameRequired
	}
	if len(name) > 128 {
		return nil, ErrPrototypeNameTooLong
	}

	tagIDs = normalizeTargetTagIDs(tagIDs)
	if len(tagIDs) == 0 {
		return nil, ErrPrototypeTagIDsRequired
	}
	for _, tagID := range tagIDs {
		if tagID <= 0 {
			return nil, ErrPrototypeTagIDInvalid
		}
	}
	if err := u.validateTagsOwnership(ctx, userID, tagIDs); err != nil {
		return nil, err
	}

	if parentPrototypeID != nil {
		if *parentPrototypeID <= 0 {
			return nil, ErrPrototypeParentNotFound
		}
		if currentID > 0 && *parentPrototypeID == currentID {
			return nil, ErrPrototypeParentSelfNotAllow
		}
		if _, err := u.Repo.FindByID(ctx, *parentPrototypeID, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrPrototypeParentNotOwned
			}
			return nil, err
		}
	}

	return tagIDs, nil
}

func (u *PrototypeUsecase) validateTagsOwnership(ctx context.Context, userID int, tagIDs []int) error {
	if u.TagRepo == nil {
		return nil
	}
	for _, tagID := range tagIDs {
		if _, err := u.TagRepo.FindByID(ctx, tagID, userID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrPrototypeTagNotFound
			}
			return err
		}
	}
	return nil
}
