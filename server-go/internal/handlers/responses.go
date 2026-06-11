package handlers

// Shared named response structs. Wire shapes must stay identical to the
// gin.H maps they replace — see responses_test.go.

// MessageResponse is a generic confirmation payload.
type MessageResponse struct {
	Message string `json:"message"`
} //@name MessageResponse

// RefreshResponse is returned by POST /auth/refresh.
type RefreshResponse struct {
	AccessToken string `json:"accessToken"`
} //@name RefreshResponse

// VerificationSubmittedResponse is returned by POST /companies/verification.
type VerificationSubmittedResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
} //@name VerificationSubmittedResponse

// VerificationStatusResponse is returned by GET /companies/verification.
type VerificationStatusResponse struct {
	VerificationStatus string `json:"verificationStatus"`
} //@name VerificationStatusResponse

// UpdateProfileResponse is returned by PUT /profiles/me.
type UpdateProfileResponse struct {
	ID                string  `json:"id"`
	UserID            string  `json:"userId"`
	Headline          *string `json:"headline"`
	About             *string `json:"about"`
	YearsOfExperience *int    `json:"yearsOfExperience"`
	Slug              string  `json:"slug"`
} //@name UpdateProfileResponse
