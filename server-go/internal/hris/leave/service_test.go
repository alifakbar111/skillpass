package leave

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"skillpass-server-go/internal/testutil"
)

func TestCreateLeaveType(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "leaveco@ex.com", "leaveco", "pass123", "Leave Co", true)
	svc := NewService(db)

	t.Run("create leave type", func(t *testing.T) {
		lt := &LeaveType{
			Name:               "Annual Leave",
			Code:               "AL",
			DefaultDaysPerYear: 12,
			IsPaid:             true,
			RequiresAttachment: false,
		}
		err := svc.CreateType(context.Background(), cID, lt)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if lt.Name != "Annual Leave" {
			t.Fatalf("expected Annual Leave, got %s", lt.Name)
		}
	})

	t.Run("list leave types", func(t *testing.T) {
		types, err := svc.ListTypes(context.Background(), cID)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(types) != 1 {
			t.Fatalf("expected 1, got %d", len(types))
		}
	})
}

func TestCreateLeaveRequest(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "leavereq@ex.com", "leavereq", "pass123", "Leave Req Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "Leave", "Employee", "leave@leavereq.com")
	svc := NewService(db)

	// Create leave type and init balance
	lt := &LeaveType{Name: "Annual", Code: "AL", DefaultDaysPerYear: 12, IsPaid: true}
	svc.CreateType(context.Background(), cID, lt)
	svc.InitBalances(context.Background(), cID, empID, 2026)

	t.Run("create leave request", func(t *testing.T) {
		start := time.Now().AddDate(0, 0, 1)
		end := start.AddDate(0, 0, 2)
		req := &LeaveRequest{
			LeaveTypeID: lt.ID,
			TotalDays:   3,
			StartDate:   start.Format("2006-01-02"),
			EndDate:     end.Format("2006-01-02"),
			Reason:      "Vacation",
		}
		err := svc.CreateRequest(context.Background(), cID, empID, req)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		if req.Status != "pending" {
			t.Fatalf("expected pending, got %s", req.Status)
		}
	})

	t.Run("insufficient balance rejected", func(t *testing.T) {
		start := time.Now().AddDate(0, 0, 20)
		req := &LeaveRequest{
			LeaveTypeID: lt.ID,
			TotalDays:   100, // way more than available
			StartDate:   start.Format("2006-01-02"),
			EndDate:     start.AddDate(0, 0, 99).Format("2006-01-02"),
			Reason:      "Long vacation",
		}
		err := svc.CreateRequest(context.Background(), cID, empID, req)
		if err == nil {
			t.Fatal("expected error for insufficient balance")
		}
	})
}

func TestReviewLeaveRequest(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "appleave@ex.com", "appleave", "pass123", "App Leave Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "App", "Leave", "app@leaveco.com")
	svc := NewService(db)

	// Setup leave type and balance
	lt := &LeaveType{Name: "Sick", Code: "SL", DefaultDaysPerYear: 10, IsPaid: true}
	svc.CreateType(context.Background(), cID, lt)
	svc.InitBalances(context.Background(), cID, empID, 2026)

	// Create request
	start := time.Now().AddDate(0, 0, 5)
	req := &LeaveRequest{
		LeaveTypeID: lt.ID,
		TotalDays:   1,
		StartDate:   start.Format("2006-01-02"),
		EndDate:     start.Format("2006-01-02"),
		Reason:      "Doctor appointment",
	}
	svc.CreateRequest(context.Background(), cID, empID, req)

	// reviewer_id FK references employees(id), not users(id)
	reviewerEmpID, _ := testutil.CreateEmployee(db, cID, "Reviewer", "Manager", "reviewer@leaveco.com")

	t.Run("approve leave", func(t *testing.T) {
		err := svc.ReviewRequest(context.Background(), cID, req.ID, reviewerEmpID, "approved", "OK")
		if err != nil {
			t.Fatalf("approve: %v", err)
		}
	})

	t.Run("cannot approve already approved", func(t *testing.T) {
		err := svc.ReviewRequest(context.Background(), cID, req.ID, reviewerEmpID, "approved", "Again")
		if err == nil {
			t.Fatal("expected error for double approve")
		}
	})
}

func TestRejectLeaveRequest(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "rejleave@ex.com", "rejleave", "pass123", "Rej Leave Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "Rej", "Leave", "rej@leaveco.com")
	svc := NewService(db)

	// Setup
	lt := &LeaveType{Name: "Annual", Code: "AL", DefaultDaysPerYear: 12, IsPaid: true}
	svc.CreateType(context.Background(), cID, lt)
	svc.InitBalances(context.Background(), cID, empID, 2026)

	start := time.Now().AddDate(0, 0, 10)
	req := &LeaveRequest{
		LeaveTypeID: lt.ID,
		TotalDays:   2,
		StartDate:   start.Format("2006-01-02"),
		EndDate:     start.AddDate(0, 0, 1).Format("2006-01-02"),
		Reason:      "Holiday",
	}
	svc.CreateRequest(context.Background(), cID, empID, req)

	// reviewer_id FK references employees(id), not users(id)
	reviewerEmpID, _ := testutil.CreateEmployee(db, cID, "Reviewer2", "Manager", "rejreviewer@leaveco.com")

	err := svc.ReviewRequest(context.Background(), cID, req.ID, reviewerEmpID, "rejected", "Not enough coverage")
	if err != nil {
		t.Fatalf("reject: %v", err)
	}

	// Verify balance not reduced for rejected
	balances, _ := svc.GetBalances(context.Background(), cID, empID, 2026)
	for _, b := range balances {
		if b.UsedDays != 0 {
			t.Fatalf("expected 0 used days after rejection, got %d", b.UsedDays)
		}
	}
}

func TestListLeaveRequests(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "listleave@ex.com", "listleave", "pass123", "List Leave Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "List", "Leave", "listleave@co.com")
	svc := NewService(db)

	lt := &LeaveType{Name: "Annual", Code: "AL", DefaultDaysPerYear: 12, IsPaid: true}
	svc.CreateType(context.Background(), cID, lt)
	svc.InitBalances(context.Background(), cID, empID, 2026)

	start := time.Now().AddDate(0, 0, 3)
	req := &LeaveRequest{
		LeaveTypeID: lt.ID,
		TotalDays:   1,
		StartDate:   start.Format("2006-01-02"),
		EndDate:     start.Format("2006-01-02"),
		Reason:      "Day off",
	}
	svc.CreateRequest(context.Background(), cID, empID, req)

	requests, err := svc.ListRequests(context.Background(), cID, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1, got %d", len(requests))
	}
	if requests[0].Status != "pending" {
		t.Fatalf("expected pending, got %s", requests[0].Status)
	}
}

func TestMyLeaveRequests(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "myleave@ex.com", "myleave", "pass123", "My Leave Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "My", "Leave", "myleave@co.com")
	svc := NewService(db)

	lt := &LeaveType{Name: "Annual", Code: "AL", DefaultDaysPerYear: 12, IsPaid: true}
	svc.CreateType(context.Background(), cID, lt)
	svc.InitBalances(context.Background(), cID, empID, 2026)

	start := time.Now().AddDate(0, 0, 7)
	req := &LeaveRequest{
		LeaveTypeID: lt.ID,
		TotalDays:   1,
		StartDate:   start.Format("2006-01-02"),
		EndDate:     start.Format("2006-01-02"),
		Reason:      "Personal day",
	}
	svc.CreateRequest(context.Background(), cID, empID, req)

	myReqs, err := svc.MyRequests(context.Background(), cID, empID)
	if err != nil {
		t.Fatalf("my requests: %v", err)
	}
	if len(myReqs) != 1 {
		t.Fatalf("expected 1, got %d", len(myReqs))
	}
}

// silence unused import
var _ = uuid.Nil
