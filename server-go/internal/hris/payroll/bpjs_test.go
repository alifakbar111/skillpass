package payroll

import "testing"

func TestBPJSComputeBelowCaps(t *testing.T) {
	cfg := DefaultBPJSConfig()
	r := cfg.Compute(10_000_000)

	// Wage below both caps → straight percentages.
	if r.KesehatanEmployee != 100_000 { // 1%
		t.Errorf("KesehatanEmployee = %.0f, want 100000", r.KesehatanEmployee)
	}
	if r.KesehatanEmployer != 400_000 { // 4%
		t.Errorf("KesehatanEmployer = %.0f, want 400000", r.KesehatanEmployer)
	}
	if r.JHTEmployee != 200_000 { // 2%
		t.Errorf("JHTEmployee = %.0f, want 200000", r.JHTEmployee)
	}
	if r.JPEmployee != 100_000 { // 1%
		t.Errorf("JPEmployee = %.0f, want 100000", r.JPEmployee)
	}
	// Employee total = 100k + 200k + 100k.
	if r.TotalEmployee != 400_000 {
		t.Errorf("TotalEmployee = %.0f, want 400000", r.TotalEmployee)
	}
	// Employer total = 400k(kes) + 370k(jht) + 24k(jkk) + 30k(jkm) + 200k(jp).
	if r.TotalEmployer != 1_024_000 {
		t.Errorf("TotalEmployer = %.0f, want 1024000", r.TotalEmployer)
	}
}

func TestBPJSKesehatanCap(t *testing.T) {
	cfg := DefaultBPJSConfig()
	// Wage above the 12,000,000 Kesehatan cap → capped base.
	r := cfg.Compute(20_000_000)
	if r.KesehatanEmployee != 120_000 { // 1% of 12,000,000
		t.Errorf("capped KesehatanEmployee = %.0f, want 120000", r.KesehatanEmployee)
	}
	if r.KesehatanEmployer != 480_000 { // 4% of 12,000,000
		t.Errorf("capped KesehatanEmployer = %.0f, want 480000", r.KesehatanEmployer)
	}
}

func TestBPJSJPCap(t *testing.T) {
	cfg := DefaultBPJSConfig()
	// Wage above the JP cap (10,547,400) → JP uses the cap.
	r := cfg.Compute(20_000_000)
	wantEmp := roundHalf(10_547_400 * 0.01)
	if r.JPEmployee != wantEmp {
		t.Errorf("capped JPEmployee = %.0f, want %.0f", r.JPEmployee, wantEmp)
	}
}

func roundHalf(v float64) float64 {
	// mirror math.Round used in Compute
	if v >= 0 {
		return float64(int64(v + 0.5))
	}
	return float64(int64(v - 0.5))
}
