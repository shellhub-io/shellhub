package internalclient

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// sessionAPI defines methods for interacting with session-related functionality.
type sessionAPI interface {
	// SessionCreate creates a new session based on the provided session creation request.
	// It returns an error if the session creation fails.
	SessionCreate(ctx context.Context, session requests.SessionCreate) error

	// FinishSession finishes the session with the specified uid.
	// It returns a slice of errors encountered during the operation.
	FinishSession(ctx context.Context, uid string) error

	// KeepAliveSession sends a keep-alive signal for the session with the specified uid.
	// It returns a slice of errors encountered during the operation.
	KeepAliveSession(ctx context.Context, uid string) error

	// UpdateSession updates some fields of [models.Session] using [models.SessionUpdate].
	UpdateSession(ctx context.Context, uid string, model *models.SessionUpdate) error
}

func (c *client) SessionCreate(ctx context.Context, session requests.SessionCreate) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetBody(session).
		Post(apiBaseURL + "/internal/sessions")

	return HasError(resp, err)
}

func (c *client) FinishSession(ctx context.Context, uid string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(apiBaseURL + "/internal/sessions/{uid}/finish")

	return HasError(resp, err)
}

func (c *client) KeepAliveSession(ctx context.Context, uid string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(apiBaseURL + "/internal/sessions/{uid}/keepalive")

	return HasError(resp, err)
}

func (c *client) UpdateSession(ctx context.Context, uid string, model *models.SessionUpdate) error {
	res, err := c.http.
		R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"tenant": uid,
		}).
		SetBody(model).
		Patch(apiBaseURL + "/internal/sessions/{tenant}")

	return HasError(res, err)
}
