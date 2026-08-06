package payroll

import "math"

// PPh21 (Indonesian income tax on salary) using the TER — Tarif Efektif
// Rata-rata (average effective rate) — monthly method from PMK 168/2023.
//
// Monthly PPh21 = monthly gross (bruto) × TER rate for the employee's category.
// The category is derived from the PTKP status (marital status + dependents).
// The December run reconciles against the annual progressive tariff (Art. 17);
// this module implements the monthly TER step.
//
// NOTE: rates follow PMK 168/2023. They are also seeded into `pph21_brackets`
// for display/override, but this table is the calculation authority. Validate
// against the current regulation before using for real disbursement.

type terBracket struct {
	upTo float64 // inclusive upper bound of monthly gross (0 = open-ended)
	rate float64 // effective rate as a fraction
}

// PTKPCategory returns the TER category (A, B, or C) for a PTKP status such as
// "TK/0", "K/1", etc. Defaults to A for unknown input.
func PTKPCategory(status string) string {
	switch status {
	case "TK/0", "TK/1", "K/0":
		return "A"
	case "TK/2", "TK/3", "K/1", "K/2":
		return "B"
	case "K/3":
		return "C"
	default:
		return "A"
	}
}

// terRate returns the effective rate for a monthly gross in a category.
func terRate(category string, monthlyGross float64) float64 {
	var table []terBracket
	switch category {
	case "B":
		table = terB
	case "C":
		table = terC
	default:
		table = terA
	}
	for _, b := range table {
		if b.upTo == 0 || monthlyGross <= b.upTo {
			return b.rate
		}
	}
	return table[len(table)-1].rate
}

// MonthlyPPh21TER computes the monthly PPh21 withholding using the TER method.
func MonthlyPPh21TER(monthlyGross float64, ptkpStatus string) float64 {
	if monthlyGross <= 0 {
		return 0
	}
	rate := terRate(PTKPCategory(ptkpStatus), monthlyGross)
	return math.Round(monthlyGross * rate)
}

// TER Category A — PTKP TK/0, TK/1, K/0 (PMK 168/2023).
var terA = []terBracket{
	{5_400_000, 0.0000}, {5_650_000, 0.0025}, {5_950_000, 0.0050}, {6_300_000, 0.0075},
	{6_750_000, 0.0100}, {7_500_000, 0.0125}, {8_550_000, 0.0150}, {9_650_000, 0.0175},
	{10_050_000, 0.0200}, {10_350_000, 0.0225}, {10_700_000, 0.0250}, {11_050_000, 0.0300},
	{11_600_000, 0.0350}, {12_500_000, 0.0400}, {13_750_000, 0.0500}, {15_100_000, 0.0600},
	{16_950_000, 0.0700}, {19_750_000, 0.0800}, {24_150_000, 0.0900}, {26_450_000, 0.1000},
	{28_000_000, 0.1100}, {30_050_000, 0.1200}, {32_400_000, 0.1300}, {35_400_000, 0.1400},
	{39_100_000, 0.1500}, {43_850_000, 0.1600}, {47_800_000, 0.1700}, {51_400_000, 0.1800},
	{56_300_000, 0.1900}, {62_200_000, 0.2000}, {68_600_000, 0.2100}, {77_500_000, 0.2200},
	{89_000_000, 0.2300}, {103_000_000, 0.2400}, {125_000_000, 0.2500}, {157_000_000, 0.2600},
	{206_000_000, 0.2700}, {337_000_000, 0.2800}, {454_000_000, 0.2900}, {550_000_000, 0.3000},
	{695_000_000, 0.3100}, {910_000_000, 0.3200}, {1_400_000_000, 0.3300}, {0, 0.3400},
}

// TER Category B — PTKP TK/2, TK/3, K/1, K/2 (PMK 168/2023).
var terB = []terBracket{
	{6_200_000, 0.0000}, {6_500_000, 0.0025}, {6_850_000, 0.0050}, {7_300_000, 0.0075},
	{9_200_000, 0.0100}, {10_750_000, 0.0150}, {11_250_000, 0.0200}, {11_600_000, 0.0250},
	{12_600_000, 0.0300}, {13_600_000, 0.0400}, {14_950_000, 0.0500}, {16_400_000, 0.0600},
	{18_450_000, 0.0700}, {21_850_000, 0.0800}, {26_000_000, 0.0900}, {27_700_000, 0.1000},
	{29_350_000, 0.1100}, {31_450_000, 0.1200}, {33_950_000, 0.1300}, {37_100_000, 0.1400},
	{41_100_000, 0.1500}, {45_800_000, 0.1600}, {49_500_000, 0.1700}, {53_800_000, 0.1800},
	{58_500_000, 0.1900}, {64_000_000, 0.2000}, {71_000_000, 0.2100}, {80_000_000, 0.2200},
	{93_000_000, 0.2300}, {109_000_000, 0.2400}, {129_000_000, 0.2500}, {163_000_000, 0.2600},
	{211_000_000, 0.2700}, {374_000_000, 0.2800}, {459_000_000, 0.2900}, {555_000_000, 0.3000},
	{704_000_000, 0.3100}, {957_000_000, 0.3200}, {1_405_000_000, 0.3300}, {0, 0.3400},
}

// TER Category C — PTKP K/3 (PMK 168/2023).
var terC = []terBracket{
	{6_600_000, 0.0000}, {6_950_000, 0.0025}, {7_350_000, 0.0050}, {7_800_000, 0.0075},
	{8_850_000, 0.0100}, {9_800_000, 0.0125}, {10_950_000, 0.0150}, {11_200_000, 0.0175},
	{12_050_000, 0.0200}, {12_950_000, 0.0300}, {14_150_000, 0.0400}, {15_550_000, 0.0500},
	{17_050_000, 0.0600}, {19_500_000, 0.0700}, {22_700_000, 0.0800}, {26_600_000, 0.0900},
	{28_100_000, 0.1000}, {30_100_000, 0.1100}, {32_600_000, 0.1200}, {35_400_000, 0.1300},
	{38_900_000, 0.1400}, {43_000_000, 0.1500}, {47_400_000, 0.1600}, {51_200_000, 0.1700},
	{55_800_000, 0.1800}, {60_400_000, 0.1900}, {66_700_000, 0.2000}, {74_500_000, 0.2100},
	{83_200_000, 0.2200}, {95_000_000, 0.2300}, {110_000_000, 0.2400}, {134_000_000, 0.2500},
	{169_000_000, 0.2600}, {221_000_000, 0.2700}, {390_000_000, 0.2800}, {463_000_000, 0.2900},
	{561_000_000, 0.3000}, {709_000_000, 0.3100}, {965_000_000, 0.3200}, {1_419_000_000, 0.3300},
	{0, 0.3400},
}
