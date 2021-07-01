package namespace

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/store"
)

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrUserNotFound      = errors.New("user not found")
	ErrNamespaceNotFound = errors.New("namespace not found")
)

func IsNamespaceOwner(ctx context.Context, s store.Store, tenantID, ownerID string) error {
	user, _, err := s.UserGetByID(ctx, ownerID, false)
	if err == store.ErrNoDocuments {
		return ErrUnauthorized
	}

	if err != nil {
		return err
	}

	ns, err := s.NamespaceGet(ctx, tenantID)
	if err == store.ErrNoDocuments {
		return ErrNamespaceNotFound
	}

	if err != nil {
		return err
	}

	if ns.Owner != user.ID {
		return ErrUnauthorized
	}

	return nil
}
