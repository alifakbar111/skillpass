package attendance

import (
	"context"
	"testing"
	"time"

	"skillpass-server-go/internal/testutil"
)

func TestClockIn(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "attco@ex.com", "attco", "pass123", "Att Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "Att", "Employee", "att@attco.com")
	svc := NewService(db)

	t.Run("clock in success", func(t *testing.T) {
		record, err := svc.ClockIn(context.Background(), cID, empID, ClockInRequest{
			Lat: -6.2088,
			Lng: 106.8456,
		})
		if err != nil {
			t.Fatalf("clock in: %v", err)
		}
		if record.AttendanceCode != "P" {
			t.Fatalf("expected P, got %s", record.AttendanceCode)
		}
	})

	t.Run("double clock in rejected", func(t *testing.T) {
		_, err := svc.ClockIn(context.Background(), cID, empID, ClockInRequest{Lat: 0, Lng: 0})
		if err == nil {
			t.Fatal("expected error for double clock in")
		}
	})
}

func TestClockOut(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "coout@ex.com", "coout", "pass123", "CoOut Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "Co", "Out", "checkout@coout.com")
	svc := NewService(db)

	// Clock in first
	svc.ClockIn(context.Background(), cID, empID, ClockInRequest{Lat: 0, Lng: 0})

	t.Run("clock out success", func(t *testing.T) {
		record, err := svc.ClockOut(context.Background(), cID, empID, ClockOutRequest{Lat: 0, Lng: 0})
		if err != nil {
			t.Fatalf("clock out: %v", err)
		}
		if record.ClockOut == nil {
			t.Fatal("expected clock out time")
		}
	})

	t.Run("clock out without clock in", func(t *testing.T) {
		empID2, _ := testutil.CreateEmployee(db, cID, "No", "Clockin", "nocheckin@coout.com")
		_, err := svc.ClockOut(context.Background(), cID, empID2, ClockOutRequest{Lat: 0, Lng: 0})
		if err == nil {
			t.Fatal("expected error for clock out without clock in")
		}
	})
}

func TestGetDashboard(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "dashco@ex.com", "dashco", "pass123", "Dash Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "Dash", "Board", "dash@dashco.com")
	svc := NewService(db)

	// Create attendance for today
	today := time.Now().Format("2006-01-02")
	testutil.CreateAttendanceLog(db, cID, empID, today)

	stats, logs, err := svc.GetDashboard(context.Background(), cID, today)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if stats.TotalEmployees != 1 {
		t.Fatalf("expected 1 employee, got %d", stats.TotalEmployees)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
}
