package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Services interface {
	UserCreate(ctx context.Context, username, password, email string) (*models.User, error)
	UserDelete(ctx context.Context, username string) error
	UserUpdate(ctx context.Context, username, password string) error
	NamespaceCreate(ctx context.Context, namespace, username, tenant string) (*models.Namespace, error)
	NamespaceAddMember(ctx context.Context, username, namespace, role string) (*models.Namespace, error)
	NamespaceRemoveMember(ctx context.Context, username, namespace string) (*models.Namespace, error)
	NamespaceDelete(ctx context.Context, namespace string) error
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
