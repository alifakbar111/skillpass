package employee

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"skillpass-server-go/internal/testutil"
)

func TestCreateEmployee(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "empco@ex.com", "empco", "pass123", "Emp Co", true)
	svc := NewService(db)

	t.Run("create employee success", func(t *testing.T) {
		emp, err := svc.Create(context.Background(), cID, CreateRequest{
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john@empco.com",
			EmploymentType: "permanent",
			JoinDate:       time.Now().Format("2006-01-02"),
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if emp.FirstName != "John" {
			t.Fatalf("expected John, got %s", emp.FirstName)
		}
		if emp.CompanyID != cID {
			t.Fatalf("expected company %v, got %v", cID, emp.CompanyID)
		}
	})

	t.Run("create employee duplicate email", func(t *testing.T) {
		_, err := svc.Create(context.Background(), cID, CreateRequest{
			FirstName:      "Jane",
			LastName:       "Doe",
			Email:          "john@empco.com", // duplicate
			EmploymentType: "permanent",
			JoinDate:       time.Now().Format("2006-01-02"),
		})
		if err == nil {
			t.Fatal("expected error for duplicate email")
		}
	})

	t.Run("create employee invalid company", func(t *testing.T) {
		_, err := svc.Create(context.Background(), uuid.New(), CreateRequest{
			FirstName:      "Ghost",
			LastName:       "Employee",
			Email:          "ghost@test.com",
			EmploymentType: "permanent",
			JoinDate:       time.Now().Format("2006-01-02"),
		})
		if err == nil {
			t.Fatal("expected error for invalid company")
		}
	})
}

func TestListEmployees(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "listco@ex.com", "listco", "pass123", "List Co", true)
	svc := NewService(db)

	// Create 3 employees
	for i := 0; i < 3; i++ {
		svc.Create(context.Background(), cID, CreateRequest{
			FirstName:      "Emp",
			LastName:       "One",
			Email:          uuid.New().String()[:8] + "@test.com",
			EmploymentType: "permanent",
			JoinDate:       time.Now().Format("2006-01-02"),
		})
	}

	result, err := svc.List(context.Background(), ListParams{
		CompanyID: cID,
		Page:      1,
		PageSize:  10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Total != 3 {
		t.Fatalf("expected 3, got %d", result.Total)
	}
	if len(result.Employees) != 3 {
		t.Fatalf("expected 3 employees, got %d", len(result.Employees))
	}
}

func TestGetEmployee(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "getco@ex.com", "getco", "pass123", "Get Co", true)
	svc := NewService(db)

	created, _ := svc.Create(context.Background(), cID, CreateRequest{
		FirstName:      "Get",
		LastName:       "Me",
		Email:          "getme@getco.com",
		EmploymentType: "permanent",
		JoinDate:       time.Now().Format("2006-01-02"),
	})

	t.Run("get existing employee", func(t *testing.T) {
		got, err := svc.Get(context.Background(), cID, created.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.FirstName != "Get" {
			t.Fatalf("expected Get, got %s", got.FirstName)
		}
	})

	t.Run("get nonexistent employee", func(t *testing.T) {
		_, err := svc.Get(context.Background(), cID, uuid.New())
		if err == nil {
			t.Fatal("expected error for nonexistent employee")
		}
	})

	t.Run("get employee from wrong company", func(t *testing.T) {
		_, cID2, _ := testutil.CreateCompanyUser(db, "wrongco@ex.com", "wrongco", "pass123", "Wrong Co", true)
		_, err := svc.Get(context.Background(), cID2, created.ID)
		if err == nil {
			t.Fatal("expected error for wrong company")
		}
	})
}

func TestUpdateEmployee(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "upco@ex.com", "upco", "pass123", "Up Co", true)
	svc := NewService(db)

	created, _ := svc.Create(context.Background(), cID, CreateRequest{
		FirstName:      "Old",
		LastName:       "Name",
		Email:          "old@upco.com",
		EmploymentType: "permanent",
		JoinDate:       time.Now().Format("2006-01-02"),
	})

	updated, err := svc.Update(context.Background(), cID, created.ID, UpdateRequest{
		FirstName: strPtr("New"),
		LastName:  strPtr("Name"),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.FirstName != "New" {
		t.Fatalf("expected New, got %s", updated.FirstName)
	}
}

func strPtr(s string) *string { return &s }
