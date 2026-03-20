package usecases

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"server/main/domain"
)

type mockActionLogRepository struct {
	logs   map[int]*domain.ActionLog
	nextID int
}

func newMockActionLogRepository() *mockActionLogRepository {
	return &mockActionLogRepository{
		logs:   make(map[int]*domain.ActionLog),
		nextID: 1,
	}
}

func containsAll(have, need []int) bool {
	m := make(map[int]bool, len(have))
	for _, v := range have {
		m[v] = true
	}
	for _, v := range need {
		if !m[v] {
			return false
		}
	}
	return true
}

func (m *mockActionLogRepository) FindAll(
	ctx context.Context,
	userID int,
	tagIDs []int,
	limit, offset int,
	options ActionLogListOptions,
) ([]*domain.ActionLog, error) {
	res := []*domain.ActionLog{}
	for _, l := range m.logs {
		if l.DeletedAt != nil || l.UserID != userID {
			continue
		}
		if len(tagIDs) > 0 && !containsAll(l.TagIDs, tagIDs) {
			continue
		}
		res = append(res, l)
	}
	return res, nil
}

func (m *mockActionLogRepository) Count(ctx context.Context, userID int, tagIDs []int, options ActionLogListOptions) (int, error) {
	logs, err := m.FindAll(ctx, userID, tagIDs, 0, 0, ActionLogListOptions{})
	if err != nil {
		return 0, err
	}
	return len(logs), nil
}

func (m *mockActionLogRepository) FindByID(ctx context.Context, id, userID int) (*domain.ActionLog, error) {
	l, ok := m.logs[id]
	if !ok || l.DeletedAt != nil || l.UserID != userID {
		return nil, sql.ErrNoRows
	}
	return l, nil
}

func (m *mockActionLogRepository) ListAttributeKeys(ctx context.Context, userID int) ([]string, error) {
	return nil, nil
}

func (m *mockActionLogRepository) Create(ctx context.Context, log *domain.ActionLog) (int, error) {
	id := m.nextID
	m.nextID++
	cp := *log
	cp.ID = id
	cp.TagIDs = append([]int(nil), log.TagIDs...)
	m.logs[id] = &cp
	return id, nil
}

func (m *mockActionLogRepository) Update(ctx context.Context, log *domain.ActionLog) error {
	if _, ok := m.logs[log.ID]; !ok {
		return sql.ErrNoRows
	}
	cp := *log
	cp.TagIDs = append([]int(nil), log.TagIDs...)
	m.logs[log.ID] = &cp
	return nil
}

func (m *mockActionLogRepository) Delete(ctx context.Context, id, userID int) error {
	l, ok := m.logs[id]
	if !ok || l.UserID != userID || l.DeletedAt != nil {
		return sql.ErrNoRows
	}
	now := time.Now()
	l.DeletedAt = &now
	return nil
}

type mockTagRepoForActionLog struct {
	exists map[int]map[int]bool
}

func (m *mockTagRepoForActionLog) FindByID(ctx context.Context, id, userID int) (*domain.Tag, error) {
	if m.exists[userID] != nil && m.exists[userID][id] {
		return &domain.Tag{ID: id, UserID: userID}, nil
	}
	return nil, sql.ErrNoRows
}

type mockPrototypeRepoForActionLog struct {
	prototypes map[int]map[int]*domain.Prototype
	nextID     int
}

func (m *mockPrototypeRepoForActionLog) FindByID(ctx context.Context, id, userID int) (*domain.Prototype, error) {
	if m.prototypes[userID] != nil {
		if p, ok := m.prototypes[userID][id]; ok {
			cp := *p
			cp.TagIDs = append([]int(nil), p.TagIDs...)
			return &cp, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockPrototypeRepoForActionLog) Create(ctx context.Context, prototype *domain.Prototype) (int, error) {
	if m.nextID == 0 {
		m.nextID = 1
	}
	id := m.nextID
	m.nextID++
	if m.prototypes[prototype.UserID] == nil {
		m.prototypes[prototype.UserID] = make(map[int]*domain.Prototype)
	}
	cp := *prototype
	cp.ID = id
	cp.TagIDs = append([]int(nil), prototype.TagIDs...)
	m.prototypes[prototype.UserID][id] = &cp
	return id, nil
}

func (m *mockPrototypeRepoForActionLog) Update(ctx context.Context, prototype *domain.Prototype) error {
	if m.prototypes[prototype.UserID] == nil {
		return sql.ErrNoRows
	}
	if _, ok := m.prototypes[prototype.UserID][prototype.ID]; !ok {
		return sql.ErrNoRows
	}
	cp := *prototype
	cp.TagIDs = append([]int(nil), prototype.TagIDs...)
	m.prototypes[prototype.UserID][prototype.ID] = &cp
	return nil
}

func TestActionLogUsecase_CreateActionLog(t *testing.T) {
	repo := newMockActionLogRepository()
	tagRepo := &mockTagRepoForActionLog{
		exists: map[int]map[int]bool{
			1: {1: true, 2: true, 3: true},
		},
	}
	prototypeRepo := &mockPrototypeRepoForActionLog{
		prototypes: map[int]map[int]*domain.Prototype{
			1: {
				10: {ID: 10, UserID: 1, Name: "base", TagIDs: []int{1}},
			},
		},
		nextID: 100,
	}
	u := &ActionLogUsecase{Repo: repo, TagRepo: tagRepo, PrototypeRepo: prototypeRepo}
	ctx := context.Background()
	occurredAt := time.Now().UTC().Truncate(time.Second)

	if _, err := u.CreateActionLog(ctx, 1, "", nil, occurredAt, "memo", "user1", []int{1}, nil); err != ErrActionLogTitleRequired {
		t.Fatalf("title required error expected: %v", err)
	}

	if _, err := u.CreateActionLog(ctx, 1, "ok", nil, time.Time{}, "memo", "user1", []int{1}, nil); err != ErrActionLogOccurredAtRequired {
		t.Fatalf("occurred_at required error expected: %v", err)
	}

	if _, err := u.CreateActionLog(ctx, 1, "ok", nil, occurredAt, "memo", "user1", nil, nil); err != ErrActionLogTagIDsRequired {
		t.Fatalf("tag_ids required error expected: %v", err)
	}

	if _, err := u.CreateActionLog(ctx, 1, "ok", nil, occurredAt, "memo", "user1", []int{1, 0}, nil); err != ErrActionLogTagIDInvalid {
		t.Fatalf("tag_id invalid error expected: %v", err)
	}

	if _, err := u.CreateActionLog(ctx, 1, "ok", nil, occurredAt, "memo", "user1", []int{99}, nil); err != ErrActionLogTagNotFound {
		t.Fatalf("tag not found error expected: %v", err)
	}

	parentID := 99
	if _, err := u.CreateActionLog(ctx, 1, "ok", &parentID, occurredAt, "memo", "user1", []int{1}, nil); err != ErrActionLogParentPrototypeNotFound {
		t.Fatalf("parent_prototype_id not found error expected: %v", err)
	}

	parentID = 10
	created, err := u.CreateActionLog(ctx, 1, "ride", &parentID, occurredAt, "memo", "user1", []int{1, 2, 2}, nil)
	if err != nil {
		t.Fatalf("CreateActionLog failed: %v", err)
	}
	if created.Title != "ride" {
		t.Fatalf("title mismatch: %s", created.Title)
	}
	if created.PrototypeID == nil || *created.PrototypeID == parentID {
		t.Fatalf("new prototype_id expected: %+v", created.PrototypeID)
	}
	createdPrototype, err := prototypeRepo.FindByID(ctx, *created.PrototypeID, 1)
	if err != nil {
		t.Fatalf("created prototype not found: %v", err)
	}
	if createdPrototype.ParentPrototypeID == nil || *createdPrototype.ParentPrototypeID != parentID {
		t.Fatalf("parent_prototype_id mismatch: %+v", createdPrototype.ParentPrototypeID)
	}
	if !containsAll(created.TagIDs, []int{1, 2}) {
		t.Fatalf("tag_ids mismatch: %+v", created.TagIDs)
	}
}

func TestActionLogUsecase_UpdateAndDelete(t *testing.T) {
	repo := newMockActionLogRepository()
	tagRepo := &mockTagRepoForActionLog{
		exists: map[int]map[int]bool{
			1: {1: true, 2: true, 3: true},
		},
	}
	prototypeRepo := &mockPrototypeRepoForActionLog{
		prototypes: map[int]map[int]*domain.Prototype{
			1: {
				10: {ID: 10, UserID: 1, Name: "base-1", TagIDs: []int{1}},
				11: {ID: 11, UserID: 1, Name: "base-2", TagIDs: []int{2}},
			},
		},
		nextID: 1000,
	}
	u := &ActionLogUsecase{Repo: repo, TagRepo: tagRepo, PrototypeRepo: prototypeRepo}
	ctx := context.Background()
	occurredAt := time.Now().UTC().Truncate(time.Second)

	initialParent := 10
	created, err := u.CreateActionLog(ctx, 1, "before", &initialParent, occurredAt, "before", "user1", []int{1, 2}, nil)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.PrototypeID == nil {
		t.Fatalf("prototype_id must not be nil")
	}
	firstPrototypeID := *created.PrototypeID

	nextParent := 11
	updated, err := u.UpdateActionLog(ctx, 1, created.ID, "after", &nextParent, occurredAt.Add(time.Minute), "after", "user1", []int{3}, nil)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Title != "after" {
		t.Fatalf("title update mismatch: %s", updated.Title)
	}
	if updated.PrototypeID == nil || *updated.PrototypeID != firstPrototypeID {
		t.Fatalf("prototype_id should be stable on update: %+v", updated.PrototypeID)
	}
	updatedPrototype, err := prototypeRepo.FindByID(ctx, *updated.PrototypeID, 1)
	if err != nil {
		t.Fatalf("updated prototype not found: %v", err)
	}
	if updatedPrototype.ParentPrototypeID == nil || *updatedPrototype.ParentPrototypeID != nextParent {
		t.Fatalf("updated parent_prototype_id mismatch: %+v", updatedPrototype.ParentPrototypeID)
	}
	if !containsAll(updated.TagIDs, []int{3}) || len(updated.TagIDs) != 1 {
		t.Fatalf("updated tag_ids mismatch: %+v", updated.TagIDs)
	}

	if err := u.DeleteActionLog(ctx, 1, created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := u.GetActionLogByID(ctx, 1, created.ID); err == nil {
		t.Fatalf("record should not be found after delete")
	}
}
