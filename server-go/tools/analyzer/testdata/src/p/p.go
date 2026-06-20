package p

type GoodResponse struct {
	ID    string            `json:"id"`
	Name  string            `json:"name"`
	Tags  []string          `json:"tags,omitempty"`
	Bio   *string           `json:"bio,omitempty"`
	Meta  map[string]int    `json:"meta,omitempty"`
	Raw   interface{}       `json:"raw,omitempty"`
}

type BadResponse struct {
	ID    string   `json:"id"`
	Tags  []string `json:"tags"`    // want "struct BadResponse: field Tags .* missing omitempty"
	Bio   *string  `json:"bio"`     // want "struct BadResponse: field Bio .* missing omitempty"
}

type PartialResponse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name,omitempty"`
	Tags  []string        `json:"tags"`     // want "struct PartialResponse: field Tags .* missing omitempty"
	Extra *string         `json:"extra,omitempty"`
	Meta  map[string]bool `json:"meta"`     // want "struct PartialResponse: field Meta .* missing omitempty"
}
