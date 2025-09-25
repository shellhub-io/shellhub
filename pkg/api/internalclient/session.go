package internalclient

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
)

// sessionAPI defines methods for interacting with session-related functionality.
type sessionAPI interface {
	// SessionCreate creates a new session based on the provided session creation request.
	// It returns an error if the session creation fails.
	SessionCreate(ctx context.Context, session requests.SessionCreate) error

	// SessionAsAuthenticated marks a session with the specified uid as authenticated.
	// It returns a slice of errors encountered during the operation.
	SessionAsAuthenticated(ctx context.Context, uid string) []error

	// FinishSession finishes the session with the specified uid.
	// It returns a slice of errors encountered during the operation.
	FinishSession(ctx context.Context, uid string) []error

	// KeepAliveSession sends a keep-alive signal for the session with the specified uid.
	// It returns a slice of errors encountered during the operation.
	KeepAliveSession(ctx context.Context, uid string) []error

	// UpdateSession updates some fields of [models.Session] using [models.SessionUpdate].
	UpdateSession(ctx context.Context, uid string, model *models.SessionUpdate) error

	// EventSessionStream creates a WebSocket client connection to endpoint to save session's events.
	EventSessionStream(ctx context.Context, uid string) (*websocket.Conn, error)

	// SaveSession saves a session as a Asciinema file into the Object Storage and delete
	// [models.SessionEventTypePtyOutput] events.
	SaveSession(ctx context.Context, uid string, seat int) error
}

func (c *client) SessionCreate(ctx context.Context, session requests.SessionCreate) error {
	_, err := c.http.
		R().
		SetContext(ctx).
		SetBody(session).
		Post(c.Config.APIBaseURL + "/internal/sessions")

	return err
}

func (c *client) SessionAsAuthenticated(ctx context.Context, uid string) []error {
	var errors []error
	_, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		SetBody(&models.Status{
			Authenticated: true,
		}).
		Patch(c.Config.APIBaseURL + "/internal/sessions/{uid}")
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *client) FinishSession(ctx context.Context, uid string) []error {
	var errors []error
	_, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(c.Config.APIBaseURL + "/internal/sessions/{uid}/finish")
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *client) KeepAliveSession(ctx context.Context, uid string) []error {
	var errors []error
	_, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(c.Config.APIBaseURL + "/internal/sessions/{uid}/keepalive")
	if err != nil {
		errors = append(errors, err)
	}

	return errors
}

func (c *client) UpdateSession(ctx context.Context, uid string, model *models.SessionUpdate) error {
	res, err := c.http.
		R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"tenant": uid,
		}).
		SetBody(model).
		Patch(c.Config.APIBaseURL + "/internal/sessions/{tenant}")
	if err != nil {
		return errors.Join(errors.New("failed to update the session due error"), err)
	}

	if res.StatusCode() != 200 {
		return errors.New("failed to update the session")
	}

	return nil
}

func (c *client) EventSessionStream(ctx context.Context, uid string) (*websocket.Conn, error) {
	// Dial the enterprise events websocket. Convert configured enterprise HTTP scheme to ws(s).
	scheme := "ws"
	if strings.HasPrefix(c.Config.APIBaseURL, "https") {
		scheme = "wss"
	}

	host := strings.TrimPrefix(strings.TrimPrefix(c.Config.APIBaseURL, "http://"), "https://")

	connection, _, err := websocket.DefaultDialer.DialContext(
		ctx,
		scheme+"://"+host+"/internal/sessions/"+uid+"/events",
		nil,
	)
	if err != nil {
		return nil, err
	}

	return connection, nil
}

func (c *client) SaveSession(ctx context.Context, uid string, seat int) error {
	res, err := c.http.
		R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"uid":  uid,
			"seat": strconv.Itoa(seat),
		}).
		Post(c.Config.EnterpriseBaseURL + "/internal/sessions/{uid}/records/{seat}")
	if err != nil {
		return errors.Join(errors.New("failed to save the Asciinema file on Object Storage"), err)
	}

	switch {
	case res.StatusCode() == 404:
		return ErrNotFound
	case res.StatusCode() == http.StatusNotAcceptable:
		// NOTE: [http.StatusNotAcceptable] indicates that session's seat shouldn't be save, but also shouldn't
		// represent an error.
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"seat": strconv.Itoa(seat),
		}).Debug("save session not acceptable")

		return nil
	case res.StatusCode() != 200:
		return errors.New("failed to save the Asciinema due status code")
	}

	return nil
}
