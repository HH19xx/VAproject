package app

import (
	"context"
	"errors"
	"strings"

	deferencedomain "server/internal/modules/deference/domain"
)

type AnalysisSnapshotRepository interface {
	FindAllByUserID(ctx context.Context, userID, limit int) ([]*deferencedomain.AnalysisSnapshot, error)
	Create(ctx context.Context, snapshot *deferencedomain.AnalysisSnapshot) (int, error)
	DeleteAllByUserID(ctx context.Context, userID int) error
}

type SnapshotService struct {
	Repo AnalysisSnapshotRepository
}

var (
	ErrSnapshotSeverityInvalid = errors.New("severity must be one of OK/NOTICE/ALERT")
	ErrSnapshotScoreInvalid    = errors.New("score must be >= 0")
	ErrSnapshotCurrentInvalid  = errors.New("current distribution stats are required")
	ErrSnapshotCreateUserEmpty = errors.New("create_user is required")
)

func (s *SnapshotService) ListByUserID(ctx context.Context, userID, limit int) ([]*deferencedomain.AnalysisSnapshot, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.Repo.FindAllByUserID(ctx, userID, limit)
}

func (s *SnapshotService) Create(ctx context.Context, snapshot *deferencedomain.AnalysisSnapshot) (*deferencedomain.AnalysisSnapshot, error) {
	if err := validateSnapshot(snapshot); err != nil {
		return nil, err
	}

	id, err := s.Repo.Create(ctx, snapshot)
	if err != nil {
		return nil, err
	}
	snapshot.ID = id
	return snapshot, nil
}

func (s *SnapshotService) DeleteAllByUserID(ctx context.Context, userID int) error {
	return s.Repo.DeleteAllByUserID(ctx, userID)
}

func validateSnapshot(snapshot *deferencedomain.AnalysisSnapshot) error {
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
