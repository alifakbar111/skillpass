package activity

import (
	"context"
	"testing"

	"skillpass-server-go/internal/testutil"
)

func TestActivityLog(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db,
		testutil.UniqueEmail("activity"),
		testutil.UniqueUsername("activity"),
		"pass123", "Activity Test Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "Activity", "Tester", testutil.UniqueEmail("activitytester"))
	svc := NewService(db)

	t.Run("log an action", func(t *testing.T) {
		entityID := empID
		err := svc.Log(context.Background(), cID, empID, "role.created", "role", &entityID, map[string]any{
			"roleName": "Test Role",
		})
		if err != nil {
			t.Fatalf("Log: %v", err)
		}
	})

	t.Run("list logs", func(t *testing.T) {
		logs, total, err := svc.List(context.Background(), cID, 10, 0)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 1 {
			t.Fatalf("expected 1 log, got %d", total)
		}
		if len(logs) != 1 {
			t.Fatalf("expected 1 log entry, got %d", len(logs))
		}
		if logs[0].Action != "role.created" {
			t.Fatalf("expected role.created, got %s", logs[0].Action)
		}
		if logs[0].ActorName != "Activity Tester" {
			t.Fatalf("expected 'Activity Tester', got '%s'", logs[0].ActorName)
		}
	})

	t.Run("log without entity ID", func(t *testing.T) {
		err := svc.Log(context.Background(), cID, empID, "settings.updated", "settings", nil, nil)
		if err != nil {
			t.Fatalf("Log without entity: %v", err)
		}

		logs, total, _ := svc.List(context.Background(), cID, 10, 0)
		if total != 2 {
			t.Fatalf("expected 2 logs, got %d", total)
		}
		if len(logs) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(logs))
		}
	})

	t.Run("list with pagination", func(t *testing.T) {
		// Log 5 more actions
		for i := 0; i < 5; i++ {
			svc.Log(context.Background(), cID, empID, "test.action", "test", nil, nil)
		}

		// First page: 3 items
		logs, total, err := svc.List(context.Background(), cID, 3, 0)
		if err != nil {
			t.Fatalf("List page 1: %v", err)
		}
		if total != 7 {
			t.Fatalf("expected total 7, got %d", total)
		}
		if len(logs) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(logs))
		}

		// Second page: next 3
		logs2, _, _ := svc.List(context.Background(), cID, 3, 3)
		if len(logs2) != 3 {
			t.Fatalf("expected 3 entries on page 2, got %d", len(logs2))
		}
	})
}
