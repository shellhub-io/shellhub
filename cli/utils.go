package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))

	return hex.EncodeToString(hash[:])
}

func normalizeString(data string) string {
	return strings.ToLower(data)
}
