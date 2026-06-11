package handlers

import (
	"encoding/json"
	"testing"
)

// These tests lock the JSON wire shapes so replacing gin.H with named
// structs cannot change what clients receive.
func TestResponseWireShapes(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"message", MessageResponse{Message: "Logged out"}, `{"message":"Logged out"}`},
		{"refresh", RefreshResponse{AccessToken: "tok"}, `{"accessToken":"tok"}`},
		{"verificationSubmitted", VerificationSubmittedResponse{Message: "Verification submitted", Status: "pending"}, `{"message":"Verification submitted","status":"pending"}`},
		{"verificationStatus", VerificationStatusResponse{VerificationStatus: "none"}, `{"verificationStatus":"none"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestUpdateProfileResponseWireShape(t *testing.T) {
	headline := "Engineer"
	years := 3
	got, err := json.Marshal(UpdateProfileResponse{
		ID:                "id1",
		UserID:            "u1",
		Headline:          &headline,
		About:             nil,
		YearsOfExperience: &years,
		Slug:              "slug1",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":"id1","userId":"u1","headline":"Engineer","about":null,"yearsOfExperience":3,"slug":"slug1"}`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
