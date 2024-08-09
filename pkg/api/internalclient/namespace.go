package internalclient

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

// namespaceAPI defines methods for interacting with namespace-related functionality.
type namespaceAPI interface {
	// NamespaceLookup retrieves namespace with the specified tenant.
	// It returns the namespace and any encountered errors.
	NamespaceLookup(tenant string) (*models.Namespace, []error)
	// InviteMember sends an invitation to join the namespace with the specified tenant ID to the
	// user with the specified email.
	// It returns an error if any and panics if the Client has no worker available.
	InviteMember(ctx context.Context, tenantID, email string) error
}

func (c *client) NamespaceLookup(tenant string) (*models.Namespace, []error) {
	namespace := new(models.Namespace)

	res, err := c.http.
		R().
		SetResult(namespace).
		Get("/api/namespaces/" + tenant)
	if err != nil {
		return nil, []error{err}
	}

	if res.StatusCode() != http.StatusOK {
		return nil, []error{err}
	}

	return namespace, nil
}

func (c *client) InviteMember(ctx context.Context, tenantID, email string) error {
	c.mustWorker()

	return c.worker.Submit(ctx, worker.TaskPattern("cloud-api:invite-member"), []byte(tenantID+":"+email))
}
