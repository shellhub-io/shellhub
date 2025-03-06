package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type MemberService interface {
	// EditNamespace updates a namespace for the specified requests.NamespaceEdit#Tenant.
	// It returns the namespace with the updated fields and an error, if any.
	EditNamespace(ctx context.Context, req *requests.NamespaceEdit) (*models.Namespace, error)

	// AddNamespaceMember adds a member to a namespace.
	//
	// In cloud environments, the member is assigned a [MemberStatusPending] status until they accept the invite via
	// an invitation email. If the target user does not exist, the email will redirect them to the registration page,
	// and the invite can be accepted after finishing. In community and enterprise environments, the status is set to
	// [MemberStatusAccepted] without sending an email.
	//
	// The role assigned to the new member must not grant more authority than the user adding them (e.g.,
	// an administrator cannot add a member with a higher role such as an owner). Owners cannot be created.
	//
	// It returns the namespace and an error, if any.
	AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error)

	// UpdateNamespaceMember updates a member with the specified ID in the specified namespace. The member's role cannot
	// have more authority than the user who is updating the member; owners cannot be created. It returns an error, if any.
	UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error

	// RemoveNamespaceMember removes a specified member from a namespace. The action must be performed by a user with higher
	// authority than the target member. Owners cannot be removed. Returns the updated namespace and an error, if any.
	RemoveNamespaceMember(ctx context.Context, req *requests.NamespaceRemoveMember) (*models.Namespace, error)

	// LeaveNamespace allows an authenticated user to remove themselves from a namespace. Owners cannot leave a namespace.
	// If the user attempts to leave the namespace they are authenticated to, their authentication token will be invalidated.
	// Returns an error, if any.
	LeaveNamespace(ctx context.Context, req *requests.LeaveNamespace) (*models.UserAuthResponse, error)
}

func (s *service) AddNamespaceMember(ctx context.Context, req *requests.NamespaceAddMember) (*models.Namespace, error) {
	return nil, nil
}

// addMember returns a transaction callback that adds a member and sends an invite if the instance is cloud.
func (s *service) addMember(memberID string, req *requests.NamespaceAddMember) store.TransactionCb {
	return nil
}

// resendMemberInvite returns a transaction callback that resends an invitation to the member with the
// specified ID.
func (s *service) resendMemberInvite(memberID string, req *requests.NamespaceAddMember) store.TransactionCb {
	return nil
}

func (s *service) UpdateNamespaceMember(ctx context.Context, req *requests.NamespaceUpdateMember) error {
	return nil
}

func (s *service) RemoveNamespaceMember(ctx context.Context, req *requests.NamespaceRemoveMember) (*models.Namespace, error) {
	return nil, nil
}

func (s *service) LeaveNamespace(ctx context.Context, req *requests.LeaveNamespace) (*models.UserAuthResponse, error) {
	return nil, nil
}

func (s *service) removeMember(ctx context.Context, ns *models.Namespace, userID string) error {
	return nil
}
