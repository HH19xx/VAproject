package usecases

import (
	"context"
	"errors"
	"strings"

	"server/main/domain"
)

type AnalysisSnapshotRepository interface {
	FindAllByUserID(ctx context.Context, userID, limit int) ([]*domain.AnalysisSnapshot, error)
	Create(ctx context.Context, snapshot *domain.AnalysisSnapshot) (int, error)
	DeleteAllByUserID(ctx context.Context, userID int) error
}

type AnalysisSnapshotUsecase struct {
	Repo AnalysisSnapshotRepository
}

var (
	ErrSnapshotSeverityInvalid = errors.New("severity must be one of OK/NOTICE/ALERT")
	ErrSnapshotScoreInvalid    = errors.New("score must be >= 0")
	ErrSnapshotCurrentInvalid  = errors.New("current distribution stats are required")
	ErrSnapshotCreateUserEmpty = errors.New("create_user is required")
)

func (u *AnalysisSnapshotUsecase) ListByUserID(ctx context.Context, userID, limit int) ([]*domain.AnalysisSnapshot, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.Repo.FindAllByUserID(ctx, userID, limit)
}

func (u *AnalysisSnapshotUsecase) Create(ctx context.Context, snapshot *domain.AnalysisSnapshot) (*domain.AnalysisSnapshot, error) {
	if err := validateSnapshot(snapshot); err != nil {
		return nil, err
	}
	id, err := u.Repo.Create(ctx, snapshot)
	if err != nil {
		return nil, err
	}
	snapshot.ID = id
	return snapshot, nil
}

func (u *AnalysisSnapshotUsecase) DeleteAllByUserID(ctx context.Context, userID int) error {
	return u.Repo.DeleteAllByUserID(ctx, userID)
}

func validateSnapshot(snapshot *domain.AnalysisSnapshot) error {
	if snapshot == nil {
		return ErrSnapshotCurrentInvalid
	}
	switch strings.ToUpper(strings.TrimSpace(snapshot.Severity)) {
	case "OK", "NOTICE", "ALERT":
	default:
		return ErrSnapshotSeverityInvalid
	}
	if snapshot.Score < 0 {
		return ErrSnapshotScoreInvalid
	}
	if snapshot.Current.Count < 0 {
		return ErrSnapshotCurrentInvalid
	}
	if strings.TrimSpace(snapshot.CreateUser) == "" {
		return ErrSnapshotCreateUserEmpty
	}
	return nil
}
