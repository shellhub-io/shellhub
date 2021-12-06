package validator

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/models"
)

func HashPassword(password string) string {
	s := sha256.Sum256([]byte(password))

	return hex.EncodeToString(s[:])
}

// FormatUser apply some formation rules to a models.User and encrypt the password.
func FormatUser(user *models.User) {
	user.ID = ""
	user.Username = strings.ToLower(user.Username)
	user.Email = strings.ToLower(user.Email)
	if user.Password != "" {
		user.Password = HashPassword(user.Password)
	}
}
