package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type backend struct{}

func (p *backend) Do(plain string) (string, error) {
	minCost := 10

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), minCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (p *backend) CompareWith(plain string, hash string) bool {
	if !strings.HasPrefix(hash, "$") {
		sha := sha256.Sum256([]byte(plain))

		return hash == hex.EncodeToString(sha[:])
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
