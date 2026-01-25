package usecases

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"server/main/domain"
)

type ActionLogRepository interface {
	FindAll(ctx context.Context, targetID int, limit, offset int) ([]*domain.ActionLog, error)
	Count(ctx context.Context, targetID int) (int, error)
	FindByID(ctx context.Context, id int) (*domain.ActionLog, error)
	Create(ctx context.Context, log *domain.ActionLog) (int, error)
	Update(ctx context.Context, log *domain.ActionLog) error
	Delete(ctx context.Context, id int) error
}

type ActionLogUsecase struct {
	Repo ActionLogRepository
}

var (
	ErrActionLogTargetInvalid = errors.New("target_id は 1 以上を指定するか、省略してください")
	ErrActionLogTypeRequired  = errors.New("action_type は必須です")
	ErrActionLogTimeRequired  = errors.New("timestamp は必須です")
)

// 一覧取得。targetID が 0 以下なら全件。
func (u *ActionLogUsecase) GetActionLogs(ctx context.Context, targetID, page, limit int) ([]*domain.ActionLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	logs, err := u.Repo.FindAll(ctx, targetID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := u.Repo.Count(ctx, targetID)
	if err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (u *ActionLogUsecase) GetActionLogByID(ctx context.Context, id int) (*domain.ActionLog, error) {
	return u.Repo.FindByID(ctx, id)
}

func (u *ActionLogUsecase) CreateActionLog(ctx context.Context, targetID *int, actionType int, ts time.Time, notes, user string) (*domain.ActionLog, error) {
	if targetID != nil && *targetID <= 0 {
		return nil, ErrActionLogTargetInvalid
	}
	if actionType <= 0 {
		return nil, ErrActionLogTypeRequired
	}
	if ts.IsZero() {
		return nil, ErrActionLogTimeRequired
	}

	log := &domain.ActionLog{
		TargetID:   targetID,
		ActionType: actionType,
		Timestamp:  ts,
		Notes:      notes,
		CreateUser: user,
		UpdateUser: user,
	}
	id, err := u.Repo.Create(ctx, log)
	if err != nil {
		return nil, err
	}
	return u.Repo.FindByID(ctx, id)
}

func (u *ActionLogUsecase) UpdateActionLog(ctx context.Context, id int, targetID *int, actionType int, ts time.Time, notes, user string) (*domain.ActionLog, error) {
	if targetID != nil && *targetID <= 0 {
		return nil, ErrActionLogTargetInvalid
	}
	if actionType <= 0 {
		return nil, ErrActionLogTypeRequired
	}
	if ts.IsZero() {
		return nil, ErrActionLogTimeRequired
	}
	log, err := u.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	log.TargetID = targetID
	log.ActionType = actionType
	log.Timestamp = ts
	log.Notes = notes
	log.UpdateUser = user

	if err := u.Repo.Update(ctx, log); err != nil {
		return nil, err
	}
	return u.Repo.FindByID(ctx, id)
}

func (u *ActionLogUsecase) DeleteActionLog(ctx context.Context, id int) error {
	return u.Repo.Delete(ctx, id)
}

// sentinel to check not found
var ErrActionLogNotFound = sql.ErrNoRows
