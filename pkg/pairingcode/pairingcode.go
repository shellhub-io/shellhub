// Package pairingcode generates and validates short, human-readable codes — the
// kind a person reads off one screen and types on another (device pairing,
// invitation links). Codes use a Crockford-style alphabet with the visually
// ambiguous characters removed.
package pairingcode

import (
	"crypto/rand"
	"strings"
)

// Alphabet is a Crockford-style alphabet with the visually ambiguous characters
// removed (no 0/O, 1/I/L, or U), so a human can read a code off one screen and
// type it on another without confusion. 30 characters.
const Alphabet = "23456789ABCDEFGHJKMNPQRSTVWXYZ"

// Canonical code lengths. Device pairing keeps 8 (~2^39, paired with a short TTL
// and per-IP rate limit); invitation codes use 12 (~2^59) because their link
// lives for days.
const (
	DeviceCodeLength = 8
	InviteCodeLength = 12
)

// New returns a fresh canonical code of the given length: uppercase, no
// separator. It draws from crypto/rand with rejection sampling so every
// character is uniform (a plain byte%30 would bias the first 256%30 characters).
func New(length int) (string, error) {
	// The largest multiple of the alphabet size that fits in a byte; bytes at or
	// above it are rejected to keep the distribution uniform.
	limit := byte(256 - (256 % len(Alphabet)))

	out := make([]byte, 0, length)
	buf := make([]byte, length)

	for len(out) < length {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}

		for _, b := range buf {
			if b >= limit {
				continue
			}

			out = append(out, Alphabet[int(b)%len(Alphabet)])
			if len(out) == length {
				break
			}
		}
	}

	return string(out), nil
}

// Normalize canonicalizes a code as typed by a user: uppercased, with the
// display grouping (hyphens) and any stray spaces removed. It does not validate
// — pair it with IsValid.
func Normalize(code string) string {
	code = strings.ToUpper(code)
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ReplaceAll(code, " ", "")

	return code
}

// IsValid reports whether code is a well-formed canonical code (right length,
// only alphabet characters). Callers should Normalize first.
func IsValid(code string, length int) bool {
	if len(code) != length {
		return false
	}

	for _, r := range code {
		if !strings.ContainsRune(Alphabet, r) {
			return false
		}
	}

	return true
}
