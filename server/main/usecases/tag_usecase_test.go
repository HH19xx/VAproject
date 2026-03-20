package usecases

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"server/main/domain"
)

type mockTagRepository struct {
	tags   map[int]*domain.Tag
	nextID int
}

func newMockTagRepository() *mockTagRepository {
	return &mockTagRepository{
		tags:   make(map[int]*domain.Tag),
		nextID: 1,
	}
}

func (m *mockTagRepository) FindAll(ctx context.Context, userID int) ([]*domain.Tag, error) {
	res := make([]*domain.Tag, 0)
	for _, t := range m.tags {
		if t.UserID == userID && t.DeletedAt == nil {
			res = append(res, t)
		}
	}
	return res, nil
}

func (m *mockTagRepository) FindByID(ctx context.Context, id, userID int) (*domain.Tag, error) {
	t, ok := m.tags[id]
	if !ok || t.DeletedAt != nil || t.UserID != userID {
		return nil, sql.ErrNoRows
	}
	return t, nil
}

func (m *mockTagRepository) Create(ctx context.Context, tag *domain.Tag) (int, error) {
	id := m.nextID
	m.nextID++
	cp := *tag
	cp.ID = id
	m.tags[id] = &cp
	return id, nil
}

func (m *mockTagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	if _, ok := m.tags[tag.ID]; !ok {
		return sql.ErrNoRows
	}
	cp := *tag
	m.tags[tag.ID] = &cp
	return nil
}

func (m *mockTagRepository) Delete(ctx context.Context, id, userID int) error {
	t, ok := m.tags[id]
	if !ok || t.UserID != userID || t.DeletedAt != nil {
		return sql.ErrNoRows
	}
	now := time.Now()
	t.DeletedAt = &now
	return nil
}

func TestTagUsecase_Create_NormalizesNameAndRejectsDuplicate(t *testing.T) {
	repo := newMockTagRepository()
	u := &TagUsecase{Repo: repo}
	ctx := context.Background()

	created, err := u.Create(ctx, 1, "  Tokyo   Station  ", " place ", " desc ", "user1")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.Name != "tokyo station" {
		t.Fatalf("unexpected normalized name: %q", created.Name)
	}
	if created.GroupName != "place" {
		t.Fatalf("group should be trimmed: %q", created.GroupName)
	}

	if _, err := u.Create(ctx, 1, "TOKYO station", "", "", "user1"); err != ErrTagNameDuplicate {
		t.Fatalf("expected ErrTagNameDuplicate, got: %v", err)
	}
}

func TestTagUsecase_Update_NormalizesAndRejectsDuplicate(t *testing.T) {
	repo := newMockTagRepository()
	u := &TagUsecase{Repo: repo}
	ctx := context.Background()

	tagA, err := u.Create(ctx, 1, "train ride", "group", "d", "user1")
	if err != nil {
		t.Fatalf("create A failed: %v", err)
	}
	tagB, err := u.Create(ctx, 1, "work", "group", "d", "user1")
	if err != nil {
		t.Fatalf("create B failed: %v", err)
	}

	updated, err := u.Update(ctx, 1, tagB.ID, "  Morning   Commute ", " grp ", " memo ", "user1")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "morning commute" {
		t.Fatalf("unexpected normalized updated name: %q", updated.Name)
	}
	if updated.GroupName != "grp" {
		t.Fatalf("group should be trimmed: %q", updated.GroupName)
	}

	if _, err := u.Update(ctx, 1, tagA.ID, "MORNING commute", "x", "y", "user1"); err != ErrTagNameDuplicate {
		t.Fatalf("expected duplicate on update, got: %v", err)
	}
}

func TestTagUsecase_ValidateAndDelete(t *testing.T) {
	repo := newMockTagRepository()
	u := &TagUsecase{Repo: repo}
	ctx := context.Background()

	if _, err := u.Create(ctx, 1, "   ", "group", "desc", "user1"); err != ErrTagNameRequired {
		t.Fatalf("expected ErrTagNameRequired, got: %v", err)
	}

	longName := make([]rune, 129)
	for i := range longName {
		longName[i] = 'a'
	}
	if _, err := u.Create(ctx, 1, string(longName), "group", "desc", "user1"); err != ErrTagNameTooLong {
		t.Fatalf("expected ErrTagNameTooLong, got: %v", err)
	}

	tag, err := u.Create(ctx, 1, "shinjuku", "place", "desc", "user1")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if err := u.Delete(ctx, 1, tag.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := u.GetByID(ctx, tag.ID, 1); err == nil {
		t.Fatalf("expected not found after delete")
	}
}
