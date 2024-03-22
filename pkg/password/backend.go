package password

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type backend struct{}

// Ensures that backend implements Password
var _ Password = (*backend)(nil)

func (p *backend) Hash(pwd string) (string, error) {
	minCost := 10

	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), minCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (p *backend) Compare(plain string, hash string) bool {
	if !strings.HasPrefix(hash, "$") {
		sha := sha256.Sum256([]byte(plain))

		return hash == hex.EncodeToString(sha[:])
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
