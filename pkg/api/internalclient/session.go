package internalclient

import (
	"context"
	"errors"
	"fmt"
	"io"

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

	// RecordSession creates a WebSocket client to the URL, and writes each channel request received by it as a
	// session record. When the context is done, the internal reading loop should return.
	RecordSession(ctx context.Context, uid string, camera chan *models.SessionRecorded, recordURL string) error

	// UpdateSession updates some fields of [models.Session] using [models.SessionUpdate].
	UpdateSession(uid string, model *models.SessionUpdate) error
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

func (c *client) RecordSession(ctx context.Context, uid string, camera chan *models.SessionRecorded, recordURL string) error {
	connection, _, err := websocket.
		DefaultDialer.
		DialContext(
			ctx,
			fmt.Sprintf("ws://%s/internal/sessions/%s/record",
				recordURL,
				uid,
			),
			nil)
	if err != nil {
		return err
	}

	defer connection.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		case frame, ok := <-camera:
			if !ok {
				// If the camera channel was closed, we stop the writing process.
				return nil
			}

			err := connection.WriteJSON(frame)
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return err
			}
		}
	}
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
