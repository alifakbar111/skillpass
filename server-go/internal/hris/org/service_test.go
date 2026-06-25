package org

import (
	"context"
	"testing"

	"skillpass-server-go/internal/testutil"
)

func TestCreateBranch(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "orgco@ex.com", "orgco", "pass123", "Org Co", true)
	svc := NewService(db)

	t.Run("create branch", func(t *testing.T) {
		b, err := svc.CreateBranch(context.Background(), cID, CreateBranchRequest{
			Name:       "Jakarta Office",
			BranchType: "head_office",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if b.Name != "Jakarta Office" {
			t.Fatalf("expected Jakarta Office, got %s", b.Name)
		}
		if b.BranchType != "head_office" {
			t.Fatalf("expected head_office, got %s", b.BranchType)
		}
	})

	t.Run("list branches", func(t *testing.T) {
		branches, err := svc.ListBranches(context.Background(), cID)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(branches) != 1 {
			t.Fatalf("expected 1, got %d", len(branches))
		}
	})

	t.Run("get branch", func(t *testing.T) {
		branches, _ := svc.ListBranches(context.Background(), cID)
		got, err := svc.GetBranch(context.Background(), cID, branches[0].ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Name != "Jakarta Office" {
			t.Fatalf("expected Jakarta Office, got %s", got.Name)
		}
	})

	t.Run("update branch", func(t *testing.T) {
		branches, _ := svc.ListBranches(context.Background(), cID)
		newName := "Surabaya Office"
		updated, err := svc.UpdateBranch(context.Background(), cID, branches[0].ID, UpdateBranchRequest{
			Name: &newName,
		})
		if err != nil {
			t.Fatalf("update: %v", err)
		}
		if updated.Name != "Surabaya Office" {
			t.Fatalf("expected Surabaya Office, got %s", updated.Name)
		}
	})

	t.Run("delete branch", func(t *testing.T) {
		branches, _ := svc.ListBranches(context.Background(), cID)
		err := svc.DeleteBranch(context.Background(), cID, branches[0].ID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		branches2, _ := svc.ListBranches(context.Background(), cID)
		if len(branches2) != 0 {
			t.Fatalf("expected 0 after delete, got %d", len(branches2))
		}
	})
}

func TestCreateDepartment(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "deptco@ex.com", "deptco", "pass123", "Dept Co", true)
	svc := NewService(db)

	t.Run("create department", func(t *testing.T) {
		d, err := svc.CreateDepartment(context.Background(), cID, CreateDepartmentRequest{
			Name: "Engineering",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if d.Name != "Engineering" {
			t.Fatalf("expected Engineering, got %s", d.Name)
		}
	})

	t.Run("list departments", func(t *testing.T) {
		depts, err := svc.ListDepartments(context.Background(), cID)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(depts) != 1 {
			t.Fatalf("expected 1, got %d", len(depts))
		}
	})
}

func TestCreatePosition(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "posco@ex.com", "posco", "pass123", "Pos Co", true)
	svc := NewService(db)

	dept, _ := svc.CreateDepartment(context.Background(), cID, CreateDepartmentRequest{
		Name: "Product",
	})

	t.Run("create position", func(t *testing.T) {
		p, err := svc.CreatePosition(context.Background(), cID, CreatePositionRequest{
			Name:         "Product Manager",
			DepartmentID: &dept.ID,
			Level:        "manager",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if p.Name != "Product Manager" {
			t.Fatalf("expected Product Manager, got %s", p.Name)
		}
	})

	t.Run("list positions", func(t *testing.T) {
		positions, err := svc.ListPositions(context.Background(), cID)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(positions) != 1 {
			t.Fatalf("expected 1, got %d", len(positions))
		}
	})
}
