package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	//"github.com/shellhub-io/shellhub/pkg/models"
)

type MiddlewareService interface {
	IsNamespaceOwner(ctx context.Context, tenantID, ownerID string) error
	IsNamespaceMember(ctx context.Context, tenantID, memberID string) error
}

func contains(members []interface{}, user string) bool {
	for _, member := range members {
		if member.(string) == user {
			return true
		}
	}
	return false
}

func (s *service) IsNamespaceOwner(ctx context.Context, tenantID, ownerID string) error {
	user, _, err := s.store.UserGetByID(ctx, ownerID, false)
	if err == store.ErrNoDocuments {
		return ErrUnauthorized
	}

	if err != nil {
		return err
	}

	ns, err := s.store.NamespaceGet(ctx, tenantID)
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

func (s *service) IsNamespaceMember(ctx context.Context, tenantID, memberID string) error {
	ns, err := s.store.NamespaceGet(ctx, tenantID)
	if err == store.ErrNoDocuments {
		return ErrNamespaceNotFound
	}
	if !contains(ns.Members, memberID) {
		return ErrUnauthorized
	}

	return nil

}
