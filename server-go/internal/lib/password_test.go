package lib

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("test-password-123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty string")
	}
	if len(hash) < 4 || (hash[:4] != "$2a$" && hash[:4] != "$2b$") {
		t.Fatalf("HashPassword should produce bcrypt hash, got: %s", hash[:min(len(hash), 4)])
	}
}

func TestVerifyPasswordCorrect(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	valid, err := VerifyPassword("correct-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Fatal("VerifyPassword should return true for correct password")
	}
}

func TestVerifyPasswordIncorrect(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	valid, err := VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if valid {
		t.Fatal("VerifyPassword should return false for incorrect password")
	}
}

func TestPasswordUniqueness(t *testing.T) {
	h1, _ := HashPassword("same-password")
	h2, _ := HashPassword("same-password")
	if h1 == h2 {
		t.Fatal("Two hashes of the same password should differ due to salt")
	}
}

func TestBcryptCostConstant(t *testing.T) {
	if BcryptCost != 10 {
		t.Fatalf("Expected BcryptCost=10, got %d", BcryptCost)
	}
}

func TestVerifyPasswordArgon2id(t *testing.T) {
	// A synthetic argon2id hash — just ensure no panic/crash
	argonHash := "$argon2id$v=19$m=65536,t=1,p=4$c29tZXNhbHRzb21lc2FsdA$rHmLqLJzBaLHkJmPjFQmJg"
	valid, err := VerifyPassword("password", argonHash)
	if err != nil {
		t.Logf("Got error for synthetic argon2id hash: %v (acceptable)", err)
	}
	if valid {
		t.Log("Synthetic hash verified (hash format may be coincidental)")
	}
}
