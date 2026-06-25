package holiday

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"skillpass-server-go/internal/testutil"
)

func TestCreateHoliday(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "holco@ex.com", "holco", "pass123", "Hol Co", true)
	svc := NewService(db)

	t.Run("create holiday", func(t *testing.T) {
		h := &Holiday{
			Name:        "Independence Day",
			Date:        "2026-08-17",
			IsRecurring: true,
		}
		err := svc.Create(context.Background(), cID, h)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if h.Name != "Independence Day" {
			t.Fatalf("expected Independence Day, got %s", h.Name)
		}
	})

	t.Run("list holidays", func(t *testing.T) {
		holidays, err := svc.List(context.Background(), cID, 2026)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(holidays) != 1 {
			t.Fatalf("expected 1, got %d", len(holidays))
		}
	})

	t.Run("update holiday", func(t *testing.T) {
		holidays, _ := svc.List(context.Background(), cID, 2026)
		h := holidays[0]
		h.Name = "Hari Kemerdekaan"
		err := svc.Update(context.Background(), cID, h.ID, &h)
		if err != nil {
			t.Fatalf("update: %v", err)
		}
	})

	t.Run("delete holiday", func(t *testing.T) {
		holidays, _ := svc.List(context.Background(), cID, 2026)
		err := svc.Delete(context.Background(), cID, holidays[0].ID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
	})

	t.Run("delete nonexistent holiday", func(t *testing.T) {
		err := svc.Delete(context.Background(), cID, uuid.New())
		if err == nil {
			t.Fatal("expected error for deleting nonexistent holiday")
		}
	})
}
