package repository

import (
	"context"
	"testing"

	"server/main/domain"
	"server/main/infra/db"
)

// TargetRepositoryの基本CRUD操作をテストする。
func TestTargetRepository_CRUD(t *testing.T) {
	// テスト用データベース接続を取得
	conn, err := db.NewPostgresConnection()
	if err != nil {
		t.Fatalf("データベース接続失敗: %v", err)
	}
	defer conn.Close()

	repo := NewTargetRepository(conn)
	ctx := context.Background()

	// Create: 新規観察対象を作成
	target := &domain.Target{
		Name:        "テスト観察対象",
		Description: "テスト用の観察対象です",
		CreateUser:  "test_user",
		UpdateUser:  "test_user",
	}

	id, err := repo.Create(ctx, target)
	if err != nil {
		t.Fatalf("Create失敗: %v", err)
	}
	if id == 0 {
		t.Fatal("Create: IDが生成されていません")
	}

	// FindByID: 作成した観察対象を取得
	found, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("FindByID失敗: %v", err)
	}
	if found.Name != target.Name {
		t.Errorf("FindByID: Name不一致。期待=%s, 実際=%s", target.Name, found.Name)
	}

	// Update: 観察対象を更新
	found.Name = "更新後の名前"
	found.Description = "更新後の説明"
	found.UpdateUser = "updated_user"
	err = repo.Update(ctx, found)
	if err != nil {
		t.Fatalf("Update失敗: %v", err)
	}

	// FindByID: 更新内容を確認
	updated, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("Update後のFindByID失敗: %v", err)
	}
	if updated.Name != "更新後の名前" {
		t.Errorf("Update: Name未更新。期待=%s, 実際=%s", "更新後の名前", updated.Name)
	}

	// FindAll: 一覧取得をテスト
	targets, err := repo.FindAll(ctx, 10, 0)
	if err != nil {
		t.Fatalf("FindAll失敗: %v", err)
	}
	if len(targets) == 0 {
		t.Error("FindAll: 結果が0件です")
	}

	// Count: 総数取得をテスト
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count失敗: %v", err)
	}
	if count == 0 {
		t.Error("Count: カウントが0です")
	}

	// Delete: 観察対象を削除
	err = repo.Delete(ctx, id)
	if err != nil {
		t.Fatalf("Delete失敗: %v", err)
	}

	// FindByID: 削除後は取得できないことを確認
	_, err = repo.FindByID(ctx, id)
	if err == nil {
		t.Error("Delete後にFindByIDが成功してしまいました")
	}
}
