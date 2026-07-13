package internalclient

import (
	"context"
	"encoding/json"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

// namespaceAPI defines methods for interacting with namespace-related functionality.
type namespaceAPI interface {
	// NamespaceLookup retrieves namespace with the specified tenant.
	// It returns the namespace and any encountered errors.
	NamespaceLookup(ctx context.Context, tenant string) (*models.Namespace, error)
	// InviteMember enqueues the invitation-email job carrying the typed notification (signature,
	// expiry, recipient email + name, forwarded proto + host). The worker renders and sends the
	// email from this payload alone. It returns an error if any and panics if the Client has no
	// worker available.
	InviteMember(ctx context.Context, notification *models.MembershipInvitationNotification) error
}

func (c *client) NamespaceLookup(ctx context.Context, tenant string) (*models.Namespace, error) {
	namespace := new(models.Namespace)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("tenant", tenant).
		SetResult(namespace).
		Get(c.config.APIBaseURL + "/internal/namespaces/{tenant}")
	if err := HasError(resp, err); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (c *client) InviteMember(ctx context.Context, notification *models.MembershipInvitationNotification) error {
	c.mustWorker()

	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return c.worker.Submit(ctx, worker.TaskPattern("cloud-api:invites"), payload)
}
