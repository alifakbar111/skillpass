package lib

import (
    "fmt"
    "github.com/google/uuid"
)

// ParseUUID validates and parses a UUID string, returning a meaningful error if invalid
func ParseUUID(id string) (uuid.UUID, error) {
    parsed, err := uuid.Parse(id)
    if err != nil {
        return uuid.Nil, fmt.Errorf("invalid UUID: %w", err)
    }
    return parsed, nil
}

// MustParseUUID panics if the UUID is invalid (for internal use where panic is acceptable)
func MustParseUUID(id string) uuid.UUID {
    parsed, err := uuid.Parse(id)
    if err != nil {
        panic(err)
    }
    return parsed
}
