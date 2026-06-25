package payroll

import (
	"context"
	"testing"

	"skillpass-server-go/internal/testutil"
)

func TestCreatePayrollRun(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "payco@ex.com", "payco", "pass123", "Pay Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "Run", "Creator", "run@payco.com")
	svc := NewService(db)

	t.Run("create payroll run", func(t *testing.T) {
		run, err := svc.CreateRun(context.Background(), cID, empID, "2026-06-01", "2026-06-30", nil)
		if err != nil {
			t.Fatalf("create run: %v", err)
		}
		if run.Status != "draft" {
			t.Fatalf("expected draft, got %s", run.Status)
		}
	})

	t.Run("list runs", func(t *testing.T) {
		runs, err := svc.ListRuns(context.Background(), cID)
		if err != nil {
			t.Fatalf("list runs: %v", err)
		}
		if len(runs) != 1 {
			t.Fatalf("expected 1 run, got %d", len(runs))
		}
	})
}

func TestCreateSalaryComponent(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "paycomp@ex.com", "paycomp", "pass123", "Pay Comp Co", true)
	svc := NewService(db)

	t.Run("create component", func(t *testing.T) {
		comp := &SalaryComponent{
			Name:          "Basic Salary",
			Code:          "BASIC",
			Type:          "earning",
			IsTaxable:     true,
			IsFixed:       true,
			DefaultAmount: 5000000,
		}
		err := svc.CreateComponent(context.Background(), cID, comp)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if comp.Name != "Basic Salary" {
			t.Fatalf("expected Basic Salary, got %s", comp.Name)
		}
	})

	t.Run("list components", func(t *testing.T) {
		comps, err := svc.ListComponents(context.Background(), cID)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(comps) != 1 {
			t.Fatalf("expected 1, got %d", len(comps))
		}
	})

	t.Run("update component", func(t *testing.T) {
		comps, _ := svc.ListComponents(context.Background(), cID)
		comps[0].Name = "Updated Basic"
		err := svc.UpdateComponent(context.Background(), cID, comps[0].ID, &comps[0])
		if err != nil {
			t.Fatalf("update: %v", err)
		}
	})

	t.Run("delete component", func(t *testing.T) {
		comp := &SalaryComponent{
			Name:          "To Delete",
			Code:          "DEL",
			Type:          "deduction",
			IsTaxable:     false,
			IsFixed:       false,
			DefaultAmount: 100000,
		}
		svc.CreateComponent(context.Background(), cID, comp)
		err := svc.DeleteComponent(context.Background(), cID, comp.ID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		comps, _ := svc.ListComponents(context.Background(), cID)
		if len(comps) != 1 {
			t.Fatalf("expected 1 after delete, got %d", len(comps))
		}
	})
}

func TestSetEmployeeSalary(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "payemp@ex.com", "payemp", "pass123", "Pay Emp Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "Pay", "Emp", "payemp@co.com")
	svc := NewService(db)

	comp := &SalaryComponent{Name: "Basic", Code: "BASIC", Type: "earning", IsTaxable: true, IsFixed: true, DefaultAmount: 5000000}
	svc.CreateComponent(context.Background(), cID, comp)

	t.Run("set employee salary", func(t *testing.T) {
		err := svc.SetEmployeeSalary(context.Background(), cID, empID, []EmployeeSalary{
			{ComponentID: comp.ID, Amount: 5000000},
		})
		if err != nil {
			t.Fatalf("set salary: %v", err)
		}
	})

	t.Run("get employee salary", func(t *testing.T) {
		salaries, err := svc.GetEmployeeSalary(context.Background(), cID, empID)
		if err != nil {
			t.Fatalf("get salary: %v", err)
		}
		if len(salaries) != 1 {
			t.Fatalf("expected 1, got %d", len(salaries))
		}
		if salaries[0].Amount != 5000000 {
			t.Fatalf("expected 5000000, got %f", salaries[0].Amount)
		}
	})
}

func TestCalculateAndApproveRun(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "paycalc@ex.com", "paycalc", "pass123", "Pay Calc Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "Pay", "Calc", "paycalc@paycalc.com")
	svc := NewService(db)

	// Create salary component and assign to employee
	comp := &SalaryComponent{Name: "Basic", Code: "BASIC", Type: "earning", IsTaxable: true, IsFixed: true, DefaultAmount: 5000000}
	svc.CreateComponent(context.Background(), cID, comp)
	svc.SetEmployeeSalary(context.Background(), cID, empID, []EmployeeSalary{
		{ComponentID: comp.ID, Amount: 5000000},
	})

	// Create run (run_by must be an employee, not a user)
	run, _ := svc.CreateRun(context.Background(), cID, empID, "2026-07-01", "2026-07-31", nil)

	t.Run("calculate run", func(t *testing.T) {
		err := svc.CalculateRun(context.Background(), cID, run.ID)
		if err != nil {
			t.Fatalf("calculate: %v", err)
		}
	})

	t.Run("approve run", func(t *testing.T) {
		err := svc.ApproveRun(context.Background(), cID, run.ID, empID)
		if err != nil {
			t.Fatalf("approve: %v", err)
		}
	})

	t.Run("cannot approve non-calculated run", func(t *testing.T) {
		run2, _ := svc.CreateRun(context.Background(), cID, empID, "2026-08-01", "2026-08-31", nil)
		err := svc.ApproveRun(context.Background(), cID, run2.ID, empID)
		if err == nil {
			t.Fatal("expected error for approving draft run")
		}
	})
}
