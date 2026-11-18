package services

import (
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// NamespacePolicy is what an instance gives every namespace it creates.
type NamespacePolicy struct {
	Edition                  envs.Edition
	RequireVerifiedDeviceKey bool
}

// New returns the namespace named name, lowercased, owned by ownerID and of type typ, created at
// now. Cloud caps it at [models.NamespaceFreeDevices] accepted devices; the other editions leave it
// unlimited. It records sessions, carries the edition's connection announcement and requires a
// verified device key when p does. The store issues the tenant ID and adds the owner as a member
// when it creates the namespace.
func (p NamespacePolicy) New(name, ownerID string, typ models.Type, now time.Time) *models.Namespace {
	maxDevices := -1
	if p.Edition.IsCloud() {
		maxDevices = 3
	}

	return &models.Namespace{
		Name:       strings.ToLower(name),
		Owner:      ownerID,
		Type:       typ,
		MaxDevices: maxDevices,
		CreatedAt:  now,
		Settings: &models.NamespaceSettings{
			SessionRecord:            true,
			ConnectionAnnouncement:   envs.AnnouncementFor(p.Edition),
			RequireVerifiedDeviceKey: p.RequireVerifiedDeviceKey,
		},
	}
}

func (s *service) namespacePolicy() NamespacePolicy {
	return NamespacePolicy{Edition: s.edition, RequireVerifiedDeviceKey: s.namespaceRequireVerifiedDeviceKey}
}
