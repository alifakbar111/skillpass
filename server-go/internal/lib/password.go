package lib

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const BcryptCost = 4

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	if strings.HasPrefix(encodedHash, "$2a$") || strings.HasPrefix(encodedHash, "$2b$") {
		err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return err == nil, err
	}

	return verifyArgon2id(password, encodedHash)
}

func verifyArgon2id(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format: expected 6 parts, got %d", len(parts))
	}
	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parse version: %w", err)
	}
	_ = version

	params := strings.Split(parts[3], ",")
	var memory, time uint32
	var threads uint8
	for _, p := range params {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		v, err := strconv.Atoi(kv[1])
		if err != nil {
			continue
		}
		switch kv[0] {
		case "m":
			memory = uint32(v)
		case "t":
			time = uint32(v)
		case "p":
			threads = uint8(v)
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(decodedHash)))

	return subtle.ConstantTimeCompare(computedHash, decodedHash) == 1, nil
}
