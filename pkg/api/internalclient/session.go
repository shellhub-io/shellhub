package internalclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
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

	// UpdateSession updates some fields of [models.Session] using [models.SessionUpdate].
	UpdateSession(uid string, model *models.SessionUpdate) error

	// EventSessionStream creates a WebSocket client connection to endpoint to save session's events.
	EventSessionStream(ctx context.Context, uid string) (*websocket.Conn, error)

	// SaveSession saves a session as a Asciinema file into the Object Storage and delete
	// [models.SessionEventTypePtyOutput] events.
	SaveSession(uid string, seat int) error
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

	if res.StatusCode() != http.StatusOK {
		return errors.New("failed to update the session")
	}

	return nil
}

func (c *client) EventSessionStream(ctx context.Context, uid string) (*websocket.Conn, error) {
	connection, _, err := websocket.
		DefaultDialer.
		DialContext(
			ctx,
			fmt.Sprintf("ws://api:8080/internal/sessions/%s/events",
				uid,
			),
			nil)
	if err != nil {
		return nil, err
	}

	return connection, nil
}

func (c *client) SaveSession(uid string, seat int) error {
	res, err := c.http.
		R().
		SetPathParams(map[string]string{
			"uid":  uid,
			"seat": fmt.Sprintf("%d", seat),
		}).
		Post("http://cloud:8080/internal/sessions/{uid}/records/{seat}")
	if err != nil {
		return errors.Join(errors.New("failed to save the Asciinema file on Object Storage"), err)
	}

	switch {
	case res.StatusCode() == http.StatusNotFound:
		return ErrNotFound
	case res.StatusCode() == http.StatusNotAcceptable:
		// NOTE: [http.StatusNotAcceptable] indicates that session's seat shouldn't be save, but also shouldn't
		// represent an error.
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"seat": fmt.Sprintf("%d", seat),
		}).Debug("save session not acceptable")

		return nil
	case res.StatusCode() != http.StatusOK:
		return errors.New("failed to save the Asciinema due status code")
	}

	return nil
}
