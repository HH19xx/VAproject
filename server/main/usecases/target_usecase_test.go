package usecases

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"testing"
	"time"

	"server/main/domain"
)

// メモリ実装のTargetRepository
type mockTargetRepository struct {
	targets []*domain.Target
	nextID  int
}

func newMockTargetRepository() *mockTargetRepository {
	return &mockTargetRepository{
		targets: []*domain.Target{},
		nextID:  1,
	}
}

func (m *mockTargetRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Target, error) {
	var active []*domain.Target
	for _, t := range m.targets {
		if t.DeletedAt == nil {
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

func (m *mockTargetRepository) FindByID(ctx context.Context, id int) (*domain.Target, error) {
	for _, t := range m.targets {
		if t.ID == id && t.DeletedAt == nil {
			return t, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockTargetRepository) Create(ctx context.Context, target *domain.Target) (int, error) {
	for _, t := range m.targets {
		if t.DeletedAt == nil && t.Name == target.Name {
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
		if t.ID == target.ID && t.DeletedAt == nil {
			// 重複チェック
			for _, other := range m.targets {
				if other.ID != target.ID && other.DeletedAt == nil && other.Name == target.Name {
					return ErrTargetNameDuplicate
				}
			}
			m.targets[i] = target
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockTargetRepository) Delete(ctx context.Context, id int) error {
	for _, t := range m.targets {
		if t.ID == id && t.DeletedAt == nil {
			now := time.Now()
			t.DeletedAt = &now
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockTargetRepository) Count(ctx context.Context) (int, error) {
	count := 0
	for _, t := range m.targets {
		if t.DeletedAt == nil {
			count++
		}
	}
	return count, nil
}

// CreateTargetのテスト
func TestTargetUsecase_CreateTarget(t *testing.T) {
	repo := newMockTargetRepository()
	usecase := &TargetUsecase{Repo: repo}
	ctx := context.Background()

	// 正常系
	target, err := usecase.CreateTarget(ctx, "テスト観察対象", "これはテスト用の説明です", "testuser")
	if err != nil {
		t.Fatalf("CreateTarget 失敗: %v", err)
	}
	if target.Name != "テスト観察対象" {
		t.Errorf("名前が一致しません: got %s, want テスト観察対象", target.Name)
	}

	// 異常系: 名前が空
	if _, err := usecase.CreateTarget(ctx, "", "説明", "testuser"); err != ErrTargetNameRequired {
		t.Errorf("空名前はErrTargetNameRequiredを返すべき: %v", err)
	}

	// 異常系: 名前が長すぎ
	longName := strings.Repeat("a", 65)
	if _, err := usecase.CreateTarget(ctx, longName, "説明", "testuser"); err != ErrTargetNameTooLong {
		t.Errorf("長すぎる名前はErrTargetNameTooLongを返すべき: %v", err)
	}

	// 異常系: 重複
	if _, err := usecase.CreateTarget(ctx, "テスト観察対象", "重複", "testuser"); err != ErrTargetNameDuplicate {
		t.Errorf("重複名はErrTargetNameDuplicateを返すべき: %v", err)
	}
}

// GetTargetsのテスト
func TestTargetUsecase_GetTargets(t *testing.T) {
	repo := newMockTargetRepository()
	usecase := &TargetUsecase{Repo: repo}
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		name := "観察対象" + strconv.Itoa(i)
		if _, err := usecase.CreateTarget(ctx, name, "説明", "testuser"); err != nil {
			t.Fatalf("テストデータ作成失敗: %v", err)
		}
	}

	targets, total, err := usecase.GetTargets(ctx, 1, 2)
	if err != nil {
		t.Fatalf("GetTargets 失敗: %v", err)
	}
	if len(targets) != 2 {
		t.Errorf("取得件数が一致しません: got %d, want 2", len(targets))
	}
	if total != 5 {
		t.Errorf("総件数が一致しません: got %d, want 5", total)
	}
}

// UpdateTargetのテスト
func TestTargetUsecase_UpdateTarget(t *testing.T) {
	repo := newMockTargetRepository()
	usecase := &TargetUsecase{Repo: repo}
	ctx := context.Background()

	target, _ := usecase.CreateTarget(ctx, "旧名前", "旧説明", "testuser")

	updated, err := usecase.UpdateTarget(ctx, target.ID, "新名前", "新説明", "updateuser")
	if err != nil {
		t.Fatalf("UpdateTarget 失敗: %v", err)
	}
	if updated.Name != "新名前" {
		t.Errorf("名前が更新されていません: got %s, want 新名前", updated.Name)
	}
	if updated.Description != "新説明" {
		t.Errorf("説明が更新されていません: got %s, want 新説明", updated.Description)
	}

	// 重複名更新の検証
	other, _ := usecase.CreateTarget(ctx, "別名", "説明", "testuser")
	if _, err := usecase.UpdateTarget(ctx, other.ID, "新名前", "説明", "updateuser"); err != ErrTargetNameDuplicate {
		t.Errorf("重複名更新はErrTargetNameDuplicateを返すべき: %v", err)
	}
}

// DeleteTargetのテスト（論理削除）
func TestTargetUsecase_DeleteTarget(t *testing.T) {
	repo := newMockTargetRepository()
	usecase := &TargetUsecase{Repo: repo}
	ctx := context.Background()

	target, _ := usecase.CreateTarget(ctx, "削除対象", "説明", "testuser")

	if _, total, _ := usecase.GetTargets(ctx, 1, 10); total != 1 {
		t.Fatalf("削除前件数が期待と異なります")
	}

	if err := usecase.DeleteTarget(ctx, target.ID); err != nil {
		t.Fatalf("DeleteTarget 失敗: %v", err)
	}

	if _, total, _ := usecase.GetTargets(ctx, 1, 10); total != 0 {
		t.Errorf("削除後件数が期待と異なります")
	}

	if _, err := usecase.GetTargetByID(ctx, target.ID); err != sql.ErrNoRows {
		t.Errorf("削除済みID取得はsql.ErrNoRowsを返すべき: %v", err)
	}
}
