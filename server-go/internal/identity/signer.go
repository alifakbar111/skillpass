// Package identity provides signature-based verifiable identity — DIDs, signed
// credentials, and an append-only hash-chained integrity log. NO blockchain:
// authenticity comes from Ed25519 signatures, integrity from SHA-256, and
// tamper-evidence from the hash chain.
package identity

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
)

// IssuerDID identifies SkillPass as the credential issuer.
const IssuerDID = "did:skillpass:issuer"

// Signer holds the issuer Ed25519 key, derived deterministically from a server
// secret so the same key is reproduced across restarts.
type Signer struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

func NewSigner(secret string) *Signer {
	seed := sha256.Sum256([]byte("skillpass-identity-issuer:" + secret))
	priv := ed25519.NewKeyFromSeed(seed[:])
	return &Signer{priv: priv, pub: priv.Public().(ed25519.PublicKey)}
}

func (s *Signer) Sign(msg []byte) string {
	return base64.StdEncoding.EncodeToString(ed25519.Sign(s.priv, msg))
}

func (s *Signer) PublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(s.pub)
}

// Verify checks a base64 signature against a base64 public key.
func Verify(pubB64, sigB64 string, msg []byte) bool {
	pub, err := base64.StdEncoding.DecodeString(pubB64)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return false
	}
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return false
	}
	return ed25519.Verify(pub, msg, sig)
}

// HashHex returns the hex SHA-256 of data.
func HashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(sum[:])
}
