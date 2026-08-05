package payroll

import "testing"

func TestPTKPCategory(t *testing.T) {
	cases := map[string]string{
		"TK/0": "A", "TK/1": "A", "K/0": "A",
		"TK/2": "B", "TK/3": "B", "K/1": "B", "K/2": "B",
		"K/3": "C",
		"":    "A", "bogus": "A",
	}
	for status, want := range cases {
		if got := PTKPCategory(status); got != want {
			t.Errorf("PTKPCategory(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestMonthlyPPh21TER(t *testing.T) {
	cases := []struct {
		name  string
		gross float64
		ptkp  string
		want  float64
	}{
		// Below the first threshold → 0% for every category.
		{"A zero band", 5_000_000, "TK/0", 0},
		{"B zero band", 6_000_000, "K/1", 0},
		{"C zero band", 6_500_000, "K/3", 0},
		// Category A: 10,000,000 falls in 9,650,001–10,050,000 → 2.00%.
		{"A 10M -> 2%", 10_000_000, "TK/0", 200_000},
		// Category A: 12,000,000 falls in 11,600,001–12,500,000 → 4.00%.
		{"A 12M -> 4%", 12_000_000, "K/0", 480_000},
		// Category A: 50,000,000 falls in 47,800,001–51,400,000 → 18.00%.
		{"A 50M -> 18%", 50_000_000, "TK/1", 9_000_000},
		// Category B: 10,000,000 falls in 9,200,001–10,750,000 → 1.50%.
		{"B 10M -> 1.5%", 10_000_000, "K/2", 150_000},
		// Category C: 15,000,000 falls in 14,150,001–15,550,000 → 5.00%.
		{"C 15M -> 5%", 15_000_000, "K/3", 750_000},
		// Top bracket A: over 1.4B → 34%.
		{"A top 34%", 2_000_000_000, "TK/0", 680_000_000},
		// Zero / negative income.
		{"zero income", 0, "TK/0", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := MonthlyPPh21TER(c.gross, c.ptkp); got != c.want {
				t.Errorf("MonthlyPPh21TER(%.0f, %q) = %.0f, want %.0f", c.gross, c.ptkp, got, c.want)
			}
		})
	}
}

func TestTERMonotonic(t *testing.T) {
	// Effective rate must be non-decreasing across the table.
	for _, cat := range []string{"A", "B", "C"} {
		prev := -1.0
		for g := 5_000_000.0; g <= 1_500_000_000; g += 250_000 {
			r := terRate(cat, g)
			if r < prev {
				t.Fatalf("category %s: rate decreased at %.0f (%.4f < %.4f)", cat, g, r, prev)
			}
			prev = r
		}
	}
}
