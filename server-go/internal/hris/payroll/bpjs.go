package payroll

import "math"

// BPJS — Indonesian social security. Two programs:
//   BPJS Kesehatan (health)         — employee 1%, employer 4%, wage cap 12,000,000
//   BPJS Ketenagakerjaan (labour):
//     JHT (old-age)                 — employee 2%,   employer 3.7%
//     JKK (work-accident)           — employer 0.24% (lowest risk class, default)
//     JKM (death)                   — employer 0.30%
//     JP (pension)                  — employee 1%,   employer 2%, wage cap 10,547,400
//
// Rates/caps are company-configurable (bpjs_config); these are 2024 defaults.

type BPJSConfig struct {
	KesehatanCap      float64 `json:"kesehatanCap"`
	KesehatanEmployee float64 `json:"kesehatanEmployee"`
	KesehatanEmployer float64 `json:"kesehatanEmployer"`
	JHTEmployee       float64 `json:"jhtEmployee"`
	JHTEmployer       float64 `json:"jhtEmployer"`
	JKKEmployer       float64 `json:"jkkEmployer"`
	JKMEmployer       float64 `json:"jkmEmployer"`
	JPCap             float64 `json:"jpCap"`
	JPEmployee        float64 `json:"jpEmployee"`
	JPEmployer        float64 `json:"jpEmployer"`
} //@name BPJSConfig

func DefaultBPJSConfig() BPJSConfig {
	return BPJSConfig{
		KesehatanCap:      12_000_000,
		KesehatanEmployee: 0.01,
		KesehatanEmployer: 0.04,
		JHTEmployee:       0.02,
		JHTEmployer:       0.037,
		JKKEmployer:       0.0024,
		JKMEmployer:       0.003,
		JPCap:             10_547_400,
		JPEmployee:        0.01,
		JPEmployer:        0.02,
	}
}

type BPJSResult struct {
	KesehatanEmployee float64 `json:"kesehatanEmployee"`
	KesehatanEmployer float64 `json:"kesehatanEmployer"`
	JHTEmployee       float64 `json:"jhtEmployee"`
	JHTEmployer       float64 `json:"jhtEmployer"`
	JKKEmployer       float64 `json:"jkkEmployer"`
	JKMEmployer       float64 `json:"jkmEmployer"`
	JPEmployee        float64 `json:"jpEmployee"`
	JPEmployer        float64 `json:"jpEmployer"`
	TotalEmployee     float64 `json:"totalEmployee"`
	TotalEmployer     float64 `json:"totalEmployer"`
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Compute returns the BPJS breakdown for a monthly wage.
func (cfg BPJSConfig) Compute(wage float64) BPJSResult {
	if wage < 0 {
		wage = 0
	}
	kesBase := minf(wage, cfg.KesehatanCap)
	jpBase := minf(wage, cfg.JPCap)

	r := BPJSResult{
		KesehatanEmployee: math.Round(kesBase * cfg.KesehatanEmployee),
		KesehatanEmployer: math.Round(kesBase * cfg.KesehatanEmployer),
		JHTEmployee:       math.Round(wage * cfg.JHTEmployee),
		JHTEmployer:       math.Round(wage * cfg.JHTEmployer),
		JKKEmployer:       math.Round(wage * cfg.JKKEmployer),
		JKMEmployer:       math.Round(wage * cfg.JKMEmployer),
		JPEmployee:        math.Round(jpBase * cfg.JPEmployee),
		JPEmployer:        math.Round(jpBase * cfg.JPEmployer),
	}
	r.TotalEmployee = r.KesehatanEmployee + r.JHTEmployee + r.JPEmployee
	r.TotalEmployer = r.KesehatanEmployer + r.JHTEmployer + r.JKKEmployer + r.JKMEmployer + r.JPEmployer
	return r
}
