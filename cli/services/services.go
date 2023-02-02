package services

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Services interface {
	UserCreate(username, password, email string) (*models.User, error)
	UserDelete(username string) error
	UserUpdate(username, password string) error
	NamespaceCreate(namespace, username, tenant string) (*models.Namespace, error)
	NamespaceAddMember(username, namespace, role string) (*models.Namespace, error)
	NamespaceRemoveMember(username, namespace string) (*models.Namespace, error)
	NamespaceDelete(namespace string) error
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Services {
	return &service{store}
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))

	return hex.EncodeToString(hash[:])
}

func normalizeField(data string) string {
	return strings.ToLower(data)
}
