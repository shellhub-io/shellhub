package internalclient

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

// namespaceAPI defines methods for interacting with namespace-related functionality.
type namespaceAPI interface {
	// NamespaceLookup retrieves namespace with the specified tenant.
	// It returns the namespace and any encountered errors.
	NamespaceLookup(ctx context.Context, tenant string) (*models.Namespace, error)
	// InviteMember sends an invitation to join the namespace with the specified tenant ID to the
	// user with the specified id. The job uses the forwarded host and proto to build the invitation
	// link. It returns an error if any and panics if the Client has no worker available.
	InviteMember(ctx context.Context, tenantID, userID, forwardedHost, forwardedProto string) error
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

func (c *client) InviteMember(ctx context.Context, tenantID, userID, forwardedHost, forwardedProto string) error {
	c.mustWorker()

	// Payload is proto-before-host so the worker can split on ":" without a host:port ambiguity
	// (tenant/user are UUIDs and proto has no colon, while the host is taken as the remainder).
	return c.worker.Submit(ctx, worker.TaskPattern("cloud-api:invites"), []byte(tenantID+":"+userID+":"+forwardedProto+":"+forwardedHost))
}
