package usecases

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"server/main/domain"
)

type mockTargetRepository struct {
	targets      []*domain.Target
	nextID       int
	logsByTarget map[int][]*domain.ActionLog
}

func newMockTargetRepository() *mockTargetRepository {
	return &mockTargetRepository{
		targets:      []*domain.Target{},
		nextID:       1,
		logsByTarget: map[int][]*domain.ActionLog{},
	}
}

func (m *mockTargetRepository) FindAll(ctx context.Context, userID, limit, offset int) ([]*domain.Target, error) {
	var active []*domain.Target
	for _, t := range m.targets {
		if t.DeletedAt == nil && t.UserID == userID {
			active = append(active, t)
		}
	}
	if offset > len(active) {
		return []*domain.Target{}, nil
	}
	end := offset + limit
	if end > len(active) {
		end = len(active)
	}
	return active[offset:end], nil
}

func (m *mockTargetRepository) FindByID(ctx context.Context, id, userID int) (*domain.Target, error) {
	for _, t := range m.targets {
		if t.ID == id && t.UserID == userID && t.DeletedAt == nil {
			return t, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockTargetRepository) Create(ctx context.Context, target *domain.Target) (int, error) {
	for _, t := range m.targets {
		if t.DeletedAt == nil && t.UserID == target.UserID && t.Name == target.Name {
			return 0, ErrTargetNameDuplicate
		}
	}
	target.ID = m.nextID
	m.nextID++
	m.targets = append(m.targets, target)
	return target.ID, nil
}

func (m *mockTargetRepository) Update(ctx context.Context, target *domain.Target) error {
	for i, t := range m.targets {
		if t.ID == target.ID && t.UserID == target.UserID && t.DeletedAt == nil {
			m.targets[i] = target
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockTargetRepository) Delete(ctx context.Context, id, userID int) error {
	for _, t := range m.targets {
		if t.ID == id && t.UserID == userID && t.DeletedAt == nil {
			now := time.Now()
			t.DeletedAt = &now
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockTargetRepository) Count(ctx context.Context, userID int) (int, error) {
	count := 0
	for _, t := range m.targets {
		if t.DeletedAt == nil && t.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockTargetRepository) FindActionLogsByTargetID(ctx context.Context, targetID, userID, limit, offset int) ([]*domain.ActionLog, error) {
	if _, err := m.FindByID(ctx, targetID, userID); err != nil {
		return nil, err
	}
	logs := m.logsByTarget[targetID]
	if offset > len(logs) {
		return []*domain.ActionLog{}, nil
	}
	end := offset + limit
	if end > len(logs) {
		end = len(logs)
	}
	return logs[offset:end], nil
}

func (m *mockTargetRepository) CountActionLogsByTargetID(ctx context.Context, targetID, userID int) (int, error) {
	if _, err := m.FindByID(ctx, targetID, userID); err != nil {
		return 0, err
	}
	return len(m.logsByTarget[targetID]), nil
}

type mockTargetTagRepository struct {
	owned map[int]bool
}

func (m *mockTargetTagRepository) FindByID(ctx context.Context, id, userID int) (*domain.Tag, error) {
	if m.owned[id] {
		return &domain.Tag{ID: id, UserID: userID}, nil
	}
	return nil, sql.ErrNoRows
}

func TestTargetUsecase_CreateTarget(t *testing.T) {
	repo := newMockTargetRepository()
	tagRepo := &mockTargetTagRepository{owned: map[int]bool{1: true, 2: true}}
	usecase := &TargetUsecase{Repo: repo, TagRepo: tagRepo}

	target, err := usecase.CreateTarget(context.Background(), 1, "通勤", "説明", "tester", "AND", []int{1, 2, 2})
	if err != nil {
		t.Fatalf("CreateTarget failed: %v", err)
	}
	if target.MatchMode != "AND" {
		t.Fatalf("unexpected match_mode: %s", target.MatchMode)
	}
	if len(target.TagIDs) != 2 {
		t.Fatalf("unexpected tag_ids length: %d", len(target.TagIDs))
	}
}

func TestTargetUsecase_CreateTarget_Validation(t *testing.T) {
	repo := newMockTargetRepository()
	tagRepo := &mockTargetTagRepository{owned: map[int]bool{1: true}}
	usecase := &TargetUsecase{Repo: repo, TagRepo: tagRepo}

	if _, err := usecase.CreateTarget(context.Background(), 1, "", "説明", "tester", "AND", []int{1}); err != ErrTargetNameRequired {
		t.Fatalf("expected ErrTargetNameRequired, got: %v", err)
	}
	if _, err := usecase.CreateTarget(context.Background(), 1, "name", "説明", "tester", "OR", []int{1}); err != ErrTargetMatchModeInvalid {
		t.Fatalf("expected ErrTargetMatchModeInvalid, got: %v", err)
	}
	if _, err := usecase.CreateTarget(context.Background(), 1, "name", "説明", "tester", "AND", []int{}); err != ErrTargetTagIDsRequired {
		t.Fatalf("expected ErrTargetTagIDsRequired, got: %v", err)
	}
	if _, err := usecase.CreateTarget(context.Background(), 1, "name", "説明", "tester", "AND", []int{9}); err != ErrTargetTagNotFound {
		t.Fatalf("expected ErrTargetTagNotFound, got: %v", err)
	}
}

func TestTargetUsecase_GetActionLogsByTargetID(t *testing.T) {
	repo := newMockTargetRepository()
	repo.targets = append(repo.targets, &domain.Target{
		ID:        1,
		UserID:    1,
		Name:      "通勤",
		MatchMode: "AND",
		TagIDs:    []int{1, 2},
	})
	repo.logsByTarget[1] = []*domain.ActionLog{
		{ID: 100, UserID: 1, TagIDs: []int{1, 2, 3}},
		{ID: 101, UserID: 1, TagIDs: []int{1, 2}},
	}

	usecase := &TargetUsecase{Repo: repo}
	logs, total, err := usecase.GetActionLogsByTargetID(context.Background(), 1, 1, 1, 20)
	if err != nil {
		t.Fatalf("GetActionLogsByTargetID failed: %v", err)
	}
	if total != 2 || len(logs) != 2 {
		t.Fatalf("unexpected result total=%d len=%d", total, len(logs))
	}
}
