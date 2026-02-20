package internalclient

import (
	"context"
	"fmt"
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
	SessionAsAuthenticated(ctx context.Context, uid string) error

	// FinishSession finishes the session with the specified uid.
	// It returns a slice of errors encountered during the operation.
	FinishSession(ctx context.Context, uid string) error

	// KeepAliveSession sends a keep-alive signal for the session with the specified uid.
	// It returns a slice of errors encountered during the operation.
	KeepAliveSession(ctx context.Context, uid string) error

	// UpdateSession updates some fields of [models.Session] using [models.SessionUpdate].
	UpdateSession(ctx context.Context, uid string, model *models.SessionUpdate) error

	// EventSessionStream creates a WebSocket client connection to endpoint to save session's events.
	EventSessionStream(ctx context.Context, uid string) (*websocket.Conn, error)

	// SaveSession saves a session as a Asciinema file into the Object Storage and delete
	// [models.SessionEventTypePtyOutput] events.
	SaveSession(ctx context.Context, uid string, seat int) error
}

func (c *client) SessionCreate(ctx context.Context, session requests.SessionCreate) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetBody(session).
		Post(c.config.APIBaseURL + "/internal/sessions")

	return HasError(resp, err)
}

func (c *client) SessionAsAuthenticated(ctx context.Context, uid string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		SetBody(&models.Status{
			Authenticated: true,
		}).
		Patch(c.config.APIBaseURL + "/internal/sessions/{uid}")

	return HasError(resp, err)
}

func (c *client) FinishSession(ctx context.Context, uid string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(c.config.APIBaseURL + "/internal/sessions/{uid}/finish")

	return HasError(resp, err)
}

func (c *client) KeepAliveSession(ctx context.Context, uid string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(c.config.APIBaseURL + "/internal/sessions/{uid}/keepalive")

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
		Patch(c.config.APIBaseURL + "/internal/sessions/{tenant}")

	return HasError(res, err)
}

func (c *client) EventSessionStream(ctx context.Context, uid string) (*websocket.Conn, error) {
	// Dial the enterprise events websocket. Convert configured enterprise HTTP scheme to ws(s).
	scheme := "ws"
	if strings.HasPrefix(c.config.APIBaseURL, "https") {
		scheme = "wss"
	}

	host := strings.TrimPrefix(strings.TrimPrefix(c.config.APIBaseURL, "http://"), "https://")

	connection, _, err := websocket.DefaultDialer.DialContext(
		ctx,
		fmt.Sprintf("%s://%s/internal/sessions/%s/events", scheme, host, uid),
		nil,
	)
	if err != nil {
		return nil, HasError(nil, err)
	}

	return connection, nil
}

func (c *client) SaveSession(ctx context.Context, uid string, seat int) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"uid":  uid,
			"seat": strconv.Itoa(seat),
		}).
		Post(c.config.APIBaseURL + "/internal/sessions/{uid}/records/{seat}")
	if err := HasError(resp, err); err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusNotAcceptable {
		// NOTE: [http.StatusNotAcceptable] indicates that session's seat shouldn't be save, but also shouldn't
		// represent an error.
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"seat": strconv.Itoa(seat),
		}).Debug("save session not acceptable")

		return nil
	}

	return HasError(resp, err)
}
