package identity

import "testing"

func TestSignVerify(t *testing.T) {
	s := NewSigner("test-secret-abcdefghijklmnop")
	msg := []byte("credential-claim-hash")
	sig := s.Sign(msg)
	pub := s.PublicKeyBase64()

	if !Verify(pub, sig, msg) {
		t.Fatal("valid signature failed to verify")
	}
	// Tampered message must fail.
	if Verify(pub, sig, []byte("credential-claim-hashX")) {
		t.Fatal("tampered message verified — should not")
	}
	// Wrong key must fail.
	other := NewSigner("different-secret-abcdefghij")
	if Verify(other.PublicKeyBase64(), sig, msg) {
		t.Fatal("signature verified under wrong key — should not")
	}
}

func TestSignerDeterministic(t *testing.T) {
	// Same secret must reproduce the same issuer key across instances.
	a := NewSigner("stable-secret-1234567890")
	b := NewSigner("stable-secret-1234567890")
	if a.PublicKeyBase64() != b.PublicKeyBase64() {
		t.Fatal("issuer key is not deterministic for the same secret")
	}
}
