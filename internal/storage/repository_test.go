package storage

import (
	"testing"

	"github.com/zxc7563598/key-heat/internal/model"
)

func batchTestDB(t *testing.T) (*DB, Repository) {
	t.Helper()
	cfg := DefaultConfig()
	cfg.Path = ":memory:"
	db, err := NewDB(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.AutoMigrate(&model.KeyEvent{}); err != nil {
		t.Fatal(err)
	}
	return db, NewRepository(db.DB)
}

func TestBatchCountsAccumulatePerDate(t *testing.T) {
	_, repo := batchTestDB(t)
	for _, batch := range []map[string]int{{"A": 2, "B": 3}, {"A": 5, "C": 4}} {
		if err := repo.UpsertKeyCounts("2026-10-30", batch); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.UpsertKeyCounts("2026-10-31", map[string]int{"A": 1}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetStatsByDate("2026-10-30")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got["A"] != 7 || got["B"] != 3 || got["C"] != 4 {
		t.Fatalf("incorrect batch totals: %v", got)
	}
	next, err := repo.GetStatsByDate("2026-10-31")
	if err != nil || next["A"] != 1 {
		t.Fatalf("dates mixed: %v, %v", next, err)
	}
}

func TestBatchCountsRollBackTogether(t *testing.T) {
	db, repo := batchTestDB(t)
	if err := repo.UpsertKeyCounts("2026-10-30", map[string]int{"A": 2}); err != nil {
		t.Fatal(err)
	}
	err := db.Exec(`CREATE TRIGGER reject_test_key BEFORE INSERT ON key_events
 WHEN NEW.key_name = 'Z' BEGIN SELECT RAISE(ABORT, 'test write failure'); END;`).Error
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertKeyCounts("2026-10-30", map[string]int{"A": 5, "Z": 4}); err == nil {
		t.Fatal("expected write failure")
	}
	got, err := repo.GetStatsByDate("2026-10-30")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got["A"] != 2 {
		t.Fatalf("partial batch committed: %v", got)
	}
	if err := repo.UpsertKeyCounts("2026-10-30", map[string]int{"A": 1}); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.GetStatsByDate("2026-10-30")
	if got["A"] != 3 {
		t.Fatalf("subsequent batch failed: %v", got)
	}
}
