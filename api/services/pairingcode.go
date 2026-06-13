package services

import (
	"crypto/rand"
	"strings"
)

// pairingCodeAlphabet is a Crockford-style alphabet with the visually ambiguous
// characters removed (no 0/O, 1/I/L, or U), so a human can read a code off one
// screen and type it on another without confusion. 30 characters.
const pairingCodeAlphabet = "23456789ABCDEFGHJKMNPQRSTVWXYZ"

// pairingCodeLength is the number of characters in a canonical code. 30^8 is
// ~2^39 of entropy, which combined with the 10-minute TTL and per-IP rate limit
// is far more than a brute-force can cover.
const pairingCodeLength = 8

// newPairingCode returns a fresh canonical code: uppercase, no separator. It
// draws from crypto/rand with rejection sampling so every character is uniform
// (a plain byte%30 would bias the first 256%30 characters).
func newPairingCode() (string, error) {
	// The largest multiple of the alphabet size that fits in a byte; bytes at or
	// above it are rejected to keep the distribution uniform.
	limit := byte(256 - (256 % len(pairingCodeAlphabet)))

	out := make([]byte, 0, pairingCodeLength)
	buf := make([]byte, pairingCodeLength)

	for len(out) < pairingCodeLength {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}

		for _, b := range buf {
			if b >= limit {
				continue
			}

			out = append(out, pairingCodeAlphabet[int(b)%len(pairingCodeAlphabet)])
			if len(out) == pairingCodeLength {
				break
			}
		}
	}

	return string(out), nil
}

// normalizePairingCode canonicalizes a code as typed by a user: uppercased, with
// the display grouping (hyphens) and any stray spaces removed. It does not
// validate — pair it with isValidPairingCode.
func normalizePairingCode(code string) string {
	code = strings.ToUpper(code)
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ReplaceAll(code, " ", "")

	return code
}

// isValidPairingCode reports whether code is a well-formed canonical code (right
// length, only alphabet characters). Callers should normalizePairingCode first.
func isValidPairingCode(code string) bool {
	if len(code) != pairingCodeLength {
		return false
	}

	for _, r := range code {
		if !strings.ContainsRune(pairingCodeAlphabet, r) {
			return false
		}
	}

	return true
}
