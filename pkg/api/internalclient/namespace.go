package internalclient

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// namespaceAPI defines methods for interacting with namespace-related functionality.
type namespaceAPI interface {
	// NamespaceLookup retrieves namespace with the specified tenant.
	// It returns the namespace and any encountered errors.
	NamespaceLookup(tenant string) (*models.Namespace, []error)
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
