package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) NamespaceAddMember(ctx context.Context, tenantID string, member *models.Member) error

func (pg *Pg) NamespaceUpdateMember(ctx context.Context, tenantID string, memberID string, changes *models.MemberChanges) error

func (pg *Pg) NamespaceRemoveMember(ctx context.Context, tenantID string, memberID string) error
