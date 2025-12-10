package usecases

import (
	"context"
	"database/sql"
	"testing"

	"server/main/domain"
)

// モックリポジトリ: テスト用のTargetRepositoryインターフェース実装
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

// 削除されていないtargetsをlimit件、offsetからスキップして返す
func (m *mockTargetRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Target, error) {
	// 削除されていないtargetsのみ抽出
	var active []*domain.Target
	for _, t := range m.targets {
		if t.DeletedAt == nil {
			active = append(active, t)
		}
	}

	// ページネーション処理
	start := offset
	if start > len(active) {
		return []*domain.Target{}, nil
	}

	end := start + limit
	if end > len(active) {
		end = len(active)
	}

	return active[start:end], nil
}

// 指定IDのtargetを返す（削除済みはエラー）
func (m *mockTargetRepository) FindByID(ctx context.Context, id int) (*domain.Target, error) {
	for _, t := range m.targets {
		if t.ID == id && t.DeletedAt == nil {
			return t, nil
		}
	}
	return nil, sql.ErrNoRows
}

// targetをメモリストレージにIDを割り振って保存
func (m *mockTargetRepository) Create(ctx context.Context, target *domain.Target) (int, error) {
	target.ID = m.nextID
	m.nextID++
	m.targets = append(m.targets, target)
	return target.ID, nil
}

// IDが一致するtargetの内容を上書き（削除済みはエラー）
func (m *mockTargetRepository) Update(ctx context.Context, target *domain.Target) error {
	for i, t := range m.targets {
		if t.ID == target.ID && t.DeletedAt == nil {
			m.targets[i] = target
			return nil
		}
	}
	return sql.ErrNoRows
}

// 論理削除フラグを設定
func (m *mockTargetRepository) Delete(ctx context.Context, id int) error {
	for _, t := range m.targets {
		if t.ID == id && t.DeletedAt == nil {
			// 論理削除: DeletedAtに値を設定
			now := &sql.NullTime{Valid: true}
			t.DeletedAt = &now.Time
			return nil
		}
	}
	return sql.ErrNoRows
}

// 削除されていないtargetsの総数を返す
func (m *mockTargetRepository) Count(ctx context.Context) (int, error) {
	count := 0
	for _, t := range m.targets {
		if t.DeletedAt == nil {
			count++
		}
	}
	return count, nil
}

// CreateTargetメソッドのテスト
func TestTargetUsecase_CreateTarget(t *testing.T) {
	repo := newMockTargetRepository()
	usecase := &TargetUsecase{Repo: repo}
	ctx := context.Background()

	// 正常系: 有効な名前で作成
	target, err := usecase.CreateTarget(ctx, "テスト観察対象", "これはテスト用の説明", "testuser")
	if err != nil {
		t.Fatalf("CreateTarget失敗: %v", err)
	}
	if target.Name != "テスト観察対象" {
		t.Errorf("名前が一致しません: got %s, want テスト観察対象", target.Name)
	}

	// 異常系: 名前が空
	_, err = usecase.CreateTarget(ctx, "", "説明", "testuser")
	if err != ErrTargetNameRequired {
		t.Errorf("名前が空の場合はErrTargetNameRequiredを返すべき: got %v", err)
	}

	// 異常系: 名前が64文字超過
	longName := string(make([]byte, 65))
	_, err = usecase.CreateTarget(ctx, longName, "説明", "testuser")
	if err != ErrTargetNameTooLong {
		t.Errorf("名前が長すぎる場合はErrTargetNameTooLongを返すべき: got %v", err)
	}
}

// GetTargetsメソッドのテスト
func TestTargetUsecase_GetTargets(t *testing.T) {
	repo := newMockTargetRepository()
	usecase := &TargetUsecase{Repo: repo}
	ctx := context.Background()

	// テストデータを5件作成
	for i := 1; i <= 5; i++ {
		_, err := usecase.CreateTarget(ctx, "観察対象"+string(rune('0'+i)), "説明", "testuser")
		if err != nil {
			t.Fatalf("テストデータ作成失敗: %v", err)
		}
	}

	// ページ1、リミット2で取得
	targets, total, err := usecase.GetTargets(ctx, 1, 2)
	if err != nil {
		t.Fatalf("GetTargets失敗: %v", err)
	}
	if len(targets) != 2 {
		t.Errorf("取得件数が一致しません: got %d, want 2", len(targets))
	}
	if total != 5 {
		t.Errorf("総件数が一致しません: got %d, want 5", total)
	}
}

// UpdateTargetメソッドのテスト
func TestTargetUsecase_UpdateTarget(t *testing.T) {
	repo := newMockTargetRepository()
	usecase := &TargetUsecase{Repo: repo}
	ctx := context.Background()

	// テストデータ作成
	target, _ := usecase.CreateTarget(ctx, "旧名前", "旧説明", "testuser")

	// 更新実行
	updated, err := usecase.UpdateTarget(ctx, target.ID, "新名前", "新説明", "updateuser")
	if err != nil {
		t.Fatalf("UpdateTarget失敗: %v", err)
	}
	if updated.Name != "新名前" {
		t.Errorf("名前が更新されていません: got %s, want 新名前", updated.Name)
	}
	if updated.Description != "新説明" {
		t.Errorf("説明が更新されていません: got %s, want 新説明", updated.Description)
	}
}

// DeleteTargetメソッドのテスト（論理削除）
func TestTargetUsecase_DeleteTarget(t *testing.T) {
	repo := newMockTargetRepository()
	usecase := &TargetUsecase{Repo: repo}
	ctx := context.Background()

	// テストデータ作成
	target, _ := usecase.CreateTarget(ctx, "削除対象", "説明", "testuser")

	// 削除前: 1件取得できる
	targets, total, _ := usecase.GetTargets(ctx, 1, 10)
	if total != 1 {
		t.Errorf("削除前の件数が一致しません: got %d, want 1", total)
	}

	// 削除実行（論理削除）
	err := usecase.DeleteTarget(ctx, target.ID)
	if err != nil {
		t.Fatalf("DeleteTarget失敗: %v", err)
	}

	// 削除後: 0件になる（論理削除されたレコードは取得されない）
	targets, total, _ = usecase.GetTargets(ctx, 1, 10)
	if total != 0 {
		t.Errorf("削除後の件数が一致しません: got %d, want 0", total)
	}
	if len(targets) != 0 {
		t.Errorf("削除後に取得されるべきではありません: got %d件", len(targets))
	}

	// 削除済みレコードをFindByIDで取得しようとするとエラー
	_, err = usecase.GetTargetByID(ctx, target.ID)
	if err != sql.ErrNoRows {
		t.Errorf("削除済みレコードの取得はエラーになるべき: got %v", err)
	}
}
