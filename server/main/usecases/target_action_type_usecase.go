package usecases

import (
	"context"
	"errors"

	"server/main/domain"
)

type TargetActionTypeRepository interface {
	FindByTargetID(ctx context.Context, targetID int) ([]*domain.TargetActionType, error)
	FindByActionTypeID(ctx context.Context, actionTypeID int) ([]*domain.TargetActionType, error)
	Create(ctx context.Context, tat *domain.TargetActionType) (int, error)
	Delete(ctx context.Context, id int) error
	DeleteByTargetAndActionType(ctx context.Context, targetID, actionTypeID int) error
	Exists(ctx context.Context, targetID, actionTypeID int) (bool, error)
}

type TargetActionTypeUsecase struct {
	Repo TargetActionTypeRepository
}

var (
	ErrTargetActionTypeAlreadyExists = errors.New("この紐づけは既に存在します")
	ErrTargetActionTypeNotFound      = errors.New("紐づけが見つかりません")
	ErrInvalidTargetID               = errors.New("対象IDが不正です")
	ErrInvalidActionTypeID           = errors.New("行動種別IDが不正です")
)

// GetActionTypesByTarget は対象に紐づく行動種別一覧を取得
func (u *TargetActionTypeUsecase) GetActionTypesByTarget(ctx context.Context, targetID int) ([]*domain.TargetActionType, error) {
	if targetID <= 0 {
		return nil, ErrInvalidTargetID
	}
	return u.Repo.FindByTargetID(ctx, targetID)
}

// GetTargetsByActionType は行動種別に紐づく対象一覧を取得
func (u *TargetActionTypeUsecase) GetTargetsByActionType(ctx context.Context, actionTypeID int) ([]*domain.TargetActionType, error) {
	if actionTypeID <= 0 {
		return nil, ErrInvalidActionTypeID
	}
	return u.Repo.FindByActionTypeID(ctx, actionTypeID)
}

// Link は対象と行動種別を紐づける
func (u *TargetActionTypeUsecase) Link(ctx context.Context, targetID, actionTypeID int, user string) (*domain.TargetActionType, error) {
	if targetID <= 0 {
		return nil, ErrInvalidTargetID
	}
	if actionTypeID <= 0 {
		return nil, ErrInvalidActionTypeID
	}

	exists, err := u.Repo.Exists(ctx, targetID, actionTypeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrTargetActionTypeAlreadyExists
	}

	tat := &domain.TargetActionType{
		TargetID:     targetID,
		ActionTypeID: actionTypeID,
		CreateUser:   user,
	}

	id, err := u.Repo.Create(ctx, tat)
	if err != nil {
		return nil, err
	}

	tat.ID = id
	return tat, nil
}

// Unlink は対象と行動種別の紐づけを解除
func (u *TargetActionTypeUsecase) Unlink(ctx context.Context, targetID, actionTypeID int) error {
	if targetID <= 0 {
		return ErrInvalidTargetID
	}
	if actionTypeID <= 0 {
		return ErrInvalidActionTypeID
	}
	return u.Repo.DeleteByTargetAndActionType(ctx, targetID, actionTypeID)
}
