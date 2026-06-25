package spdid

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"skillpass-server-go/internal/testutil"
)

func TestCreateAndGetDID(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "spco@ex.com", "spco", "pass123", "SP Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "SP", "Employee", "sp@spco.com")
	svc := NewService(db)

	t.Run("create DID", func(t *testing.T) {
		did, err := svc.CreateDID(context.Background(), cID, empID)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if did.ID == uuid.Nil {
			t.Fatal("expected DID ID")
		}
		if did.Status != "active" {
			t.Fatalf("expected active, got %s", did.Status)
		}
	})

	t.Run("get DID", func(t *testing.T) {
		did, err := svc.GetDID(context.Background(), cID, empID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if did.ID == uuid.Nil {
			t.Fatal("expected DID ID")
		}
	})

	t.Run("get DID for nonexistent employee", func(t *testing.T) {
		_, err := svc.GetDID(context.Background(), cID, uuid.New())
		if err == nil {
			t.Fatal("expected error for nonexistent employee")
		}
	})
}

func TestRevokeDID(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "spco2@ex.com", "spco2", "pass123", "SP Co 2", true)
	empID, _ := testutil.CreateEmployee(db, cID, "SP2", "Employee2", "sp2@spco2.com")
	svc := NewService(db)

	svc.CreateDID(context.Background(), cID, empID)

	t.Run("revoke DID", func(t *testing.T) {
		err := svc.RevokeDID(context.Background(), cID, empID)
		if err != nil {
			t.Fatalf("revoke: %v", err)
		}
	})

	t.Run("get revoked DID shows revoked status", func(t *testing.T) {
		did, err := svc.GetDID(context.Background(), cID, empID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if did.Status != "revoked" {
			t.Fatalf("expected revoked, got %s", did.Status)
		}
	})
}
