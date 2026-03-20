package repository

import (
	"context"
	"testing"

	"server/main/domain"
	"server/main/infra/db"
)

func TestTargetRepository_CRUD(t *testing.T) {
	conn, err := db.NewPostgresConnection()
	if err != nil {
		t.Skipf("DB接続がないためスキップします: %v", err)
	}
	defer conn.Close()

	repo := NewTargetRepository(conn)
	ctx := context.Background()
	userID := 1

	target := &domain.Target{
		UserID:      userID,
		Name:        "repo_test_target",
		Description: "repo test",
		MatchMode:   "AND",
		TagIDs:      []int{1},
		CreateUser:  "test_user",
		UpdateUser:  "test_user",
	}

	id, err := repo.Create(ctx, target)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(ctx, id, userID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.MatchMode != "AND" {
		t.Fatalf("unexpected match_mode: %s", found.MatchMode)
	}

	found.Description = "updated"
	found.TagIDs = []int{1, 2}
	if err := repo.Update(ctx, found); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, err := repo.FindByID(ctx, id, userID)
	if err != nil {
		t.Fatalf("FindByID after update failed: %v", err)
	}
	if len(updated.TagIDs) == 0 {
		t.Fatalf("tag_ids should not be empty")
	}

	if err := repo.Delete(ctx, id, userID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
