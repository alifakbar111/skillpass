package handlers

import (
	"testing"
)

func TestConstantTimeEqualize(t *testing.T) {
	// First call initializes the dummy hash; should not panic.
	constantTimeEqualize("hunter2")
	constantTimeEqualize("correct horse battery staple")
	// Call many times to confirm the dummy hash is cached and stable.
	first := getDummyBcryptHash()
	for i := 0; i < 100; i++ {
		got := getDummyBcryptHash()
		if got != first {
			t.Fatalf("dummy bcrypt hash is not stable: got %q vs %q", got, first)
		}
	}
}
