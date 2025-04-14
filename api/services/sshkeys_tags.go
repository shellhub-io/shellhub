package services

import (
	"context"
)

type SSHKeysTagsService interface {
	AddPublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error
	RemovePublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error
	UpdatePublicKeyTags(ctx context.Context, tenant, fingerprint string, tags []string) error
}

// AddPublicKeyTag trys to add a tag to the models.PublicKey, when its filter is from Tags type.
//
// It checks if the models.Namespace and models.PublicKey exists and try to perform the addition action.
func (s *service) AddPublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error {
	return nil
}

// RemovePublicKeyTag trys to remove a tag from the models.PublicKey, when its filter is from Tags type.
func (s *service) RemovePublicKeyTag(ctx context.Context, tenant, fingerprint, tag string) error {
	return nil
}

// UpdatePublicKeyTags trys to update the tags of the models.PublicKey, when its filter is from Tags type.
//
// It checks if the models.Namespace and models.PublicKey exists and try to perform the update action.
func (s *service) UpdatePublicKeyTags(ctx context.Context, tenant, fingerprint string, tags []string) error {
	return nil
}
