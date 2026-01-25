package usecases

import (
	"context"
	"database/sql"
	"errors"

	"server/main/domain"
)

type ActionTypeRepository interface {
	FindAll(ctx context.Context) ([]*domain.ActionType, error)
	FindByID(ctx context.Context, id int) (*domain.ActionType, error)
	Create(ctx context.Context, at *domain.ActionType) (int, error)
	Update(ctx context.Context, at *domain.ActionType) error
	Delete(ctx context.Context, id int) error
}

type ActionTypeUsecase struct {
	Repo ActionTypeRepository
}

var (
	ErrActionTypeNameRequired = errors.New("action_name は必須です")
	ErrActionTypeNameTooLong  = errors.New("action_name は64文字以内で入力してください")
)

func (u *ActionTypeUsecase) GetAll(ctx context.Context) ([]*domain.ActionType, error) {
	return u.Repo.FindAll(ctx)
}

func (u *ActionTypeUsecase) GetByID(ctx context.Context, id int) (*domain.ActionType, error) {
	return u.Repo.FindByID(ctx, id)
}

func (u *ActionTypeUsecase) Create(ctx context.Context, name, description, user string) (*domain.ActionType, error) {
	if name == "" {
		return nil, ErrActionTypeNameRequired
	}
	if len(name) > 64 {
		return nil, ErrActionTypeNameTooLong
	}
	at := &domain.ActionType{
		ActionName:  name,
		Description: description,
		CreateUser:  user,
		UpdateUser:  user,
	}
	id, err := u.Repo.Create(ctx, at)
	if err != nil {
		return nil, err
	}
	return u.Repo.FindByID(ctx, id)
}

func (u *ActionTypeUsecase) Update(ctx context.Context, id int, name, description, user string) (*domain.ActionType, error) {
	if name == "" {
		return nil, ErrActionTypeNameRequired
	}
	if len(name) > 64 {
		return nil, ErrActionTypeNameTooLong
	}
	at, err := u.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	at.ActionName = name
	at.Description = description
	at.UpdateUser = user
	if err := u.Repo.Update(ctx, at); err != nil {
		return nil, err
	}
	return u.Repo.FindByID(ctx, id)
}

func (u *ActionTypeUsecase) Delete(ctx context.Context, id int) error {
	return u.Repo.Delete(ctx, id)
}

var ErrActionTypeNotFound = sql.ErrNoRows
