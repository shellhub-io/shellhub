package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
)

type SSHKeysTagsService interface {
	AddPublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error
	RemovePublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error
}

// AddPublicKeyTag trys a tag to the tag's list in models.PublicKey.
//
// It checks if the models.Namespace and models.PublicKey exists and try to perform the addition action.
func (s *service) AddPublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error {
	exist := func(item string, list []string) bool {
		for _, elem := range list {
			if elem == item {
				return true
			}
		}

		return false
	}

	// Checks if the namespace exists.
	namespace, err := s.GetNamespace(ctx, tenant)
	if err != nil || namespace == nil {
		return ErrNamespaceNotFound
	}

	// Checks if the public key exists.
	key, err := s.GetPublicKey(ctx, fingerprint, tenant)
	if err != nil || key == nil {
		return ErrPublicKeyNotFound
	}

	if len(key.Filter.Tags) == DeviceMaxTags {
		return ErrMaxTagReached
	}

	tags, _, err := s.GetTags(ctx, tenant)
	if err != nil {
		return err
	}

	if !exist(tag, tags) {
		return ErrTagNameNotFound
	}

	// Trys to add a public key.
	err = s.store.PublicKeyAddTag(ctx, tenant, fingerprint, tag)
	if err != nil {
		switch err {
		case store.ErrNoDocuments:
			return ErrDuplicateTagName
		default:
			return err
		}
	}

	return nil
}

func (s *service) RemovePublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error {
	exist := func(item string, list []string) bool {
		for _, elem := range list {
			if elem == item {
				return true
			}
		}

		return false
	}

	// Checks if the namespace exists.
	namespace, err := s.GetNamespace(ctx, tenant)
	if err != nil || namespace == nil {
		return ErrNamespaceNotFound
	}

	// Checks if the public key exists.
	key, err := s.GetPublicKey(ctx, fingerprint, tenant)
	if err != nil || key == nil {
		return ErrPublicKeyNotFound
	}

	// Checks if the tag already exists in the device.
	if !exist(tag, key.Filter.Tags) {
		return ErrTagNameNotFound
	}

	// Trys to remove a public key.
	err = s.store.PublicKeyRemoveTag(ctx, tenant, fingerprint, tag)
	if err != nil {
		return err
	}

	return nil
}
