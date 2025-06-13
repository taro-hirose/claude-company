package models

import (
	"crypto/rand"
	mathrand "math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func GenerateULID() string {
	entropy := ulid.Monotonic(mathrand.New(mathrand.NewSource(time.Now().UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func GenerateSecureULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

func ParseULID(s string) (ulid.ULID, error) {
	return ulid.Parse(s)
}

func IsValidULID(s string) bool {
	_, err := ulid.Parse(s)
	return err == nil
}