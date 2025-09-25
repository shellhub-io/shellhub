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

var (
	// ErrSessionRequestFailed indicates that the session request failed.
	ErrSessionRequestFailed = errors.New("session request failed")
	// ErrSessionCreationFailed indicates that the operation to create a session failed.
	ErrSessionCreationFailed = errors.New("session creation failed")
)

func (c *client) SessionCreate(ctx context.Context, session requests.SessionCreate) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetBody(session).
		Post(c.Config.APIBaseURL + "/internal/sessions")
	if err != nil {
		return errors.Join(ErrSessionRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return ErrSessionCreationFailed
	}

	return nil
}

// ErrSessionAsAuthenticatedFailed indicates that the operation to mark a session as authenticated failed.
var ErrSessionAsAuthenticatedFailed = errors.New("mark session as authenticated failed")

func (c *client) SessionAsAuthenticated(ctx context.Context, uid string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		SetBody(&models.Status{
			Authenticated: true,
		}).
		Patch(c.Config.APIBaseURL + "/internal/sessions/{uid}")
	if err != nil {
		return errors.Join(ErrSessionRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return ErrSessionAsAuthenticatedFailed
	}

	return nil
}

// ErrFinishSessionFailed indicates that the operation to finish a session failed.
var ErrFinishSessionFailed = errors.New("finish session failed")

func (c *client) FinishSession(ctx context.Context, uid string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(c.Config.APIBaseURL + "/internal/sessions/{uid}/finish")
	if err != nil {
		return errors.Join(ErrSessionRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return ErrFinishSessionFailed
	}

	return nil
}

// ErrKeepAliveSessionFailed indicates that the operation to keep alive a session failed.
var ErrKeepAliveSessionFailed = errors.New("keep alive session failed")

func (c *client) KeepAliveSession(ctx context.Context, uid string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("uid", uid).
		Post(c.Config.APIBaseURL + "/internal/sessions/{uid}/keepalive")
	if err != nil {
		return errors.Join(ErrSessionRequestFailed, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return ErrKeepAliveSessionFailed
	}

	return nil
}

// ErrUpdateSessionFailed indicates that the operation to update a session failed.
var ErrUpdateSessionFailed = errors.New("update session failed")

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
		return errors.Join(ErrSessionRequestFailed, err)
	}

	if res.StatusCode() != 200 {
		return ErrUpdateSessionFailed
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
		return nil, errors.Join(ErrSessionRequestFailed, err)
	}

	return connection, nil
}

// ErrSaveSessionFailed indicates that the operation to save a session failed.
var ErrSaveSessionFailed = errors.New("save session failed")

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
		return errors.Join(ErrSessionRequestFailed, err)
	}

	if res.StatusCode() == http.StatusNotAcceptable {
		// NOTE: [http.StatusNotAcceptable] indicates that session's seat shouldn't be save, but also shouldn't
		// represent an error.
		logrus.WithFields(logrus.Fields{
			"uid":  uid,
			"seat": strconv.Itoa(seat),
		}).Debug("save session not acceptable")

		return nil
	}

	if res.StatusCode() != http.StatusOK {
		return ErrSaveSessionFailed
	}

	return nil
}
