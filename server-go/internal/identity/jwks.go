package identity

import "encoding/base64"

// JWK is a single JSON Web Key (Ed25519 / OKP).
type JWK struct {
	Kty string `json:"kty"` // OKP
	Crv string `json:"crv"` // Ed25519
	X   string `json:"x"`   // base64url public key (no padding)
	Use string `json:"use"` // sig
	Kid string `json:"kid"`
	Alg string `json:"alg"` // EdDSA
} //@name JWK

// JWKS is the JSON Web Key Set published at /.well-known/jwks.json.
type JWKS struct {
	Keys []JWK `json:"keys,omitempty"`
} //@name JWKS

// JWKS returns the issuer public key in JWKS form so third parties can verify
// SkillPass-signed credentials without contacting us.
func (s *Service) JWKS() JWKS {
	// PublicKeyBase64 is standard base64; JWK x is base64url without padding.
	raw, err := base64.StdEncoding.DecodeString(s.signer.PublicKeyBase64())
	if err != nil {
		return JWKS{Keys: []JWK{}}
	}
	return JWKS{Keys: []JWK{{
		Kty: "OKP",
		Crv: "Ed25519",
		X:   base64.RawURLEncoding.EncodeToString(raw),
		Use: "sig",
		Kid: "skillpass-issuer-1",
		Alg: "EdDSA",
	}}}
}
