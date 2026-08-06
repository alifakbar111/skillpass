package face

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

// cryptor encrypts biometric embeddings at rest with AES-256-GCM. The key is
// derived from a server secret (SHA-256), so no raw embedding is ever stored.
type cryptor struct {
	gcm cipher.AEAD
}

func newCryptor(secret string) (*cryptor, error) {
	if secret == "" {
		return nil, errors.New("encryption secret is required")
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &cryptor{gcm: gcm}, nil
}

func (c *cryptor) encrypt(plain []byte) ([]byte, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return c.gcm.Seal(nonce, nonce, plain, nil), nil
}

func (c *cryptor) decrypt(enc []byte) ([]byte, error) {
	ns := c.gcm.NonceSize()
	if len(enc) < ns {
		return nil, errors.New("ciphertext too short")
	}
	return c.gcm.Open(nil, enc[:ns], enc[ns:], nil)
}
