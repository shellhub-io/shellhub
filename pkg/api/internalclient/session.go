package internalclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// sessionAPI defines methods for interacting with session-related functionality.
type sessionAPI interface {
	// SessionCreate creates a new session based on the provided session creation request.
	// It returns an error if the session creation fails.
	SessionCreate(session requests.SessionCreate) error

	// SessionAsAuthenticated marks a session with the specified uid as authenticated.
	// It returns a slice of errors encountered during the operation.
	SessionAsAuthenticated(uid string) []error

	// FinishSession finishes the session with the specified uid.
	// It returns a slice of errors encountered during the operation.
	FinishSession(uid string) []error

	// KeepAliveSession sends a keep-alive signal for the session with the specified uid.
	// It returns a slice of errors encountered during the operation.
	KeepAliveSession(uid string) []error

	// RecordSession creates a WebSocket client connection to the URL.
	// RecordSession(ctx context.Context, uid string, seat int, recordURL string) (*websocket.Conn, error)

	// UpdateSession updates some fields of [models.Session] using [models.SessionUpdate].
	UpdateSession(uid string, model *models.SessionUpdate) error

	// ConnectSessionEvents connects a Web Socket to receive session's events.
	ConnectSessionEvents(ctx context.Context, uid string) (*websocket.Conn, error)
}

func (c *client) SessionCreate(session requests.SessionCreate) error {
	_, err := c.http.
		R().
		SetBody(session).
		Post("/internal/sessions")

	return err
}

func (c *client) SessionAsAuthenticated(uid string) []error {
	var errors []error

	_, err := c.http.
		R().
		SetBody(&models.Status{
			Authenticated: true,
		}).
		Patch(fmt.Sprintf("/internal/sessions/%s", uid))
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *client) FinishSession(uid string) []error {
	var errors []error

	_, err := c.http.
		R().
		Post(fmt.Sprintf("/internal/sessions/%s/finish", uid))
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *client) KeepAliveSession(uid string) []error {
	var errors []error

	_, err := c.http.
		R().
		Post(fmt.Sprintf("/internal/sessions/%s/keepalive", uid))
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *client) UpdateSession(uid string, model *models.SessionUpdate) error {
	res, err := c.http.
		R().
		SetPathParams(map[string]string{
			"tenant": uid,
		}).
		SetBody(model).
		Patch("/internal/sessions/{tenant}")
	if err != nil {
		return errors.Join(errors.New("failed to update the session due error"), err)
	}

	if res.StatusCode() != 200 {
		return errors.New("failed to update the session")
	}

	return nil
}

// ConnectSessionEvents connects a Web Socket to receive session's events.
func (c *client) ConnectSessionEvents(ctx context.Context, uid string) (*websocket.Conn, error) {
	connection, _, err := websocket.
		DefaultDialer.
		DialContext(
			ctx,
			fmt.Sprintf("ws://%s/internal/sessions/%s/events",
				"api:8080", // TODO: set the right path here.
				uid,
			),
			nil)
	if err != nil {
		return nil, err
	}

	return connection, nil
}
