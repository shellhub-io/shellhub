package environment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/moby/moby/api/types/container"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

const (
	objectStorageBucketEnv = "SHELLHUB_OBJECT_STORAGE_BUCKET"
	unreadablePrefix       = "unreadable:"
)

// SetSessionRecording turns the server-side recording of the namespace tenant's sessions on or
// off through PUT /api/users/security/:tenant, failing t unless the server answers 200. The
// client must be authenticated as a member allowed to change it. A connection reads the setting
// once, when it opens, so it applies to connections made afterwards, not to new channels on one
// already open. It outlives t's context, so a t.Cleanup can restore the setting with it.
func (dc *DockerCompose) SetSessionRecording(t *testing.T, tenant string, on bool) {
	t.Helper()

	resp, err := dc.R(context.WithoutCancel(t.Context())).
		SetBody(map[string]bool{"session_record": on}).
		Put("/api/users/security/" + tenant)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

// GetSessionRecord fetches the asciinema recording of seat in session uid through
// GET /api/sessions/:uid/records/:seat and returns the answer whatever its status code. ctx bounds
// the request. It returns the error only for a request that never got an answer.
func (dc *DockerCompose) GetSessionRecord(ctx context.Context, uid string, seat int) (*resty.Response, error) {
	return dc.R(ctx).Get(sessionRecordPath(uid, seat))
}

// DeleteSessionRecord deletes the recording of seat in session uid through
// DELETE /api/sessions/:uid/records/:seat and returns the answer whatever its status code. ctx
// bounds the request. It returns the error only for a request that never got an answer.
func (dc *DockerCompose) DeleteSessionRecord(ctx context.Context, uid string, seat int) (*resty.Response, error) {
	return dc.R(ctx).Delete(sessionRecordPath(uid, seat))
}

func sessionRecordPath(uid string, seat int) string {
	return "/api/sessions/" + uid + "/records/" + strconv.Itoa(seat)
}

// SessionEventCountOfType returns how many events of eventType the database holds for the session
// uid, failing t when the count cannot be read. See [Stack.SessionEventCountOfType].
func (dc *DockerCompose) SessionEventCountOfType(t *testing.T, uid string, eventType models.SessionEventType) int {
	t.Helper()

	count, err := dc.stack.SessionEventCountOfType(t.Context(), uid, eventType)
	require.NoError(t, err)

	return count
}

// AwaitSessionEventOfType waits up to 30 seconds for the database to hold an event of eventType for
// the session uid, failing t if it never does. The server writes a session's events in batches, so
// an event the session has already produced reaches the table a moment later.
func (dc *DockerCompose) AwaitSessionEventOfType(t *testing.T, uid string, eventType models.SessionEventType) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		count, err := dc.stack.SessionEventCountOfType(t.Context(), uid, eventType)
		assert.NoError(tt, err)
		assert.Positive(tt, count, "session %s has no %s event yet", uid, eventType)
	}, 30*time.Second, time.Second)
}

// SessionEventCountOfType returns how many events of eventType the database holds for the session
// uid, reading the table directly, so it counts the terminal output the API never lists. ctx bounds
// the exec. It returns the errors [Stack.SQLValue] does, and strconv.Atoi's error when the count
// psql prints is not an integer.
func (s *Stack) SessionEventCountOfType(ctx context.Context, uid string, eventType models.SessionEventType) (int, error) {
	return s.sqlInt(ctx,
		"SELECT count(*) FROM session_events WHERE session_id = :'uid' AND type = :'type'",
		map[string]string{"uid": uid, "type": string(eventType)})
}

// SessionConverted reports whether the archive job has marked the session uid converted, its
// recording moved to object storage, failing t when the mark cannot be read. See
// [Stack.SessionConverted].
func (dc *DockerCompose) SessionConverted(t *testing.T, uid string) bool {
	t.Helper()

	converted, err := dc.stack.SessionConverted(t.Context(), uid)
	require.NoError(t, err)

	return converted
}

// AwaitSessionConverted waits up to 30 seconds for the archive job to mark the session uid
// converted, failing t if it does not.
func (dc *DockerCompose) AwaitSessionConverted(t *testing.T, uid string) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		converted, err := dc.stack.SessionConverted(t.Context(), uid)
		assert.NoError(tt, err)
		assert.True(tt, converted, "the archive job has not marked session %s converted", uid)
	}, 30*time.Second, time.Second)
}

// SessionConverted reports whether the session uid is marked converted. Only the database holds
// the mark, so it reads the row directly. ctx bounds the exec. It returns the errors
// [Stack.SQLValue] does, the one for no row when no session has the uid, and strconv.ParseBool's
// error when the mark is NULL.
func (s *Stack) SessionConverted(ctx context.Context, uid string) (bool, error) {
	value, err := s.SQLValue(ctx,
		"SELECT converted FROM sessions WHERE id = :'uid'",
		map[string]string{"uid": uid})
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(value)
}

// BreakSessionPtyRequest makes the terminal request the session uid recorded unreadable, so
// building its recording fails until [DockerCompose.RepairSessionPtyRequest] restores it. It stands
// in for a session whose archive fails for a reason of its own, which nothing in the product can
// be asked to cause. It fails t unless the database holds exactly one unbroken terminal request
// for the session, so wait for the session's events to be written first.
func (dc *DockerCompose) BreakSessionPtyRequest(t *testing.T, uid string) {
	t.Helper()

	dc.requireOnePtyRequestUpdated(t,
		"UPDATE session_events SET data = :'prefix' || data "+
			"WHERE session_id = :'uid' AND type = :'type' AND NOT starts_with(data, :'prefix')",
		uid)
}

// RepairSessionPtyRequest restores the terminal request [DockerCompose.BreakSessionPtyRequest]
// broke, failing t unless it finds exactly one broken request on the session uid.
func (dc *DockerCompose) RepairSessionPtyRequest(t *testing.T, uid string) {
	t.Helper()

	dc.requireOnePtyRequestUpdated(t,
		"UPDATE session_events SET data = substr(data, length(:'prefix') + 1) "+
			"WHERE session_id = :'uid' AND type = :'type' AND starts_with(data, :'prefix')",
		uid)
}

func (dc *DockerCompose) requireOnePtyRequestUpdated(t *testing.T, statement, uid string) {
	t.Helper()

	output, err := dc.stack.SQL(t.Context(), statement, map[string]string{
		"uid":    uid,
		"type":   string(models.SessionEventTypePtyRequest),
		"prefix": unreadablePrefix,
	})
	require.NoError(t, err)
	require.Equal(t, "UPDATE 1", strings.TrimSpace(output))
}

// RecordingObjects returns the names of the objects object storage keeps for the session uid, one
// "<seat>.asciinema" per archived seat, sorted. It fails t when the stack runs no object storage
// or the bucket cannot be listed. See [Stack.RecordingObjects].
func (dc *DockerCompose) RecordingObjects(t *testing.T, uid string) []string {
	t.Helper()

	objects, err := dc.stack.RecordingObjects(t.Context(), uid)
	require.NoError(t, err)

	return objects
}

// RemoveRecordingObject deletes the archived recording of seat in session uid from object storage
// behind the server's back, failing t unless the object was there. See
// [Stack.RemoveRecordingObject].
func (dc *DockerCompose) RemoveRecordingObject(t *testing.T, uid string, seat int) {
	t.Helper()

	require.NoError(t, dc.stack.RemoveRecordingObject(t.Context(), uid, seat))
}

// RecordingObjects lists the objects the stack's bucket keeps under the session uid, by name
// relative to the session, sorted. ctx bounds the exec. It returns an error when the stack runs no
// object storage, the error of an exec that could not start or whose output could not be read, an
// error carrying the client's output when it exits non-zero, and an error naming the first line of
// the listing that is not an entry the client listed successfully.
func (s *Stack) RecordingObjects(ctx context.Context, uid string) ([]string, error) {
	output, err := s.objectStorage(ctx, "ls", "--recursive", "--json", s.recordingPrefix(uid))
	if err != nil {
		return nil, err
	}

	objects := []string{}

	for line := range strings.Lines(output) {
		entry := struct {
			Status string `json:"status"`
			Key    string `json:"key"`
			Type   string `json:"type"`
		}{}

		if err := json.Unmarshal([]byte(line), &entry); err != nil || entry.Status != "success" {
			return nil, fmt.Errorf("mc listed %q", strings.TrimSpace(line))
		}

		if entry.Type == "file" {
			objects = append(objects, entry.Key)
		}
	}

	slices.Sort(objects)

	return objects, nil
}

// RemoveRecordingObject deletes the object holding the archived recording of seat in session uid.
// ctx bounds the exec. It returns the errors [Stack.RecordingObjects] does, and an error when the
// bucket holds no such object.
func (s *Stack) RemoveRecordingObject(ctx context.Context, uid string, seat int) error {
	object := strconv.Itoa(seat) + ".asciinema"

	objects, err := s.RecordingObjects(ctx, uid)
	if err != nil {
		return err
	}

	if !slices.Contains(objects, object) {
		return fmt.Errorf("the bucket holds no %s for session %s, only %v", object, uid, objects)
	}

	_, err = s.objectStorage(ctx, "rm", s.recordingPrefix(uid)+object)

	return err
}

// StopObjectStorage stops the stack's object storage, so every call the server makes to it fails,
// until [DockerCompose.StartObjectStorage] starts it again. It starts it again when t ends if the
// test has not, even when stopping it failed, so a failed test leaves the next one a stack that
// stores recordings. t's context bounds the stop. It fails t when the stack runs no object storage
// or the container does not stop, and, when t ends with the storage stopped, in each case
// [DockerCompose.StartObjectStorage] fails it.
func (dc *DockerCompose) StopObjectStorage(t *testing.T) {
	t.Helper()

	storage := dc.Service(ServiceObjectStorage)
	require.NotNil(t, storage, "the stack runs no object storage")

	t.Cleanup(func() {
		state, err := storage.State(context.WithoutCancel(t.Context()))
		if err == nil && state.Running {
			return
		}

		dc.StartObjectStorage(t)
	})

	timeout := 30 * time.Second
	require.NoError(t, storage.Stop(t.Context(), &timeout))
}

// StartObjectStorage starts the stack's object storage again after
// [DockerCompose.StopObjectStorage], keeping the objects it held, and waits until its health check
// passes and it lists the stack's bucket. It fails t when the stack runs no object storage, when
// the container does not start, and when it is not healthy or does not list the bucket within a
// minute. It outlives t's context, so a t.Cleanup can start the storage with it.
func (dc *DockerCompose) StartObjectStorage(t *testing.T) {
	t.Helper()

	ctx := context.WithoutCancel(t.Context())
	storage := dc.Service(ServiceObjectStorage)
	require.NotNil(t, storage, "the stack runs no object storage")

	require.NoError(t, storage.Start(ctx))

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		state, err := storage.State(ctx)
		if !assert.NoError(tt, err) || !assert.NotNil(tt, state.Health, "the object storage has no health check") ||
			!assert.Equal(tt, container.Healthy, state.Health.Status) {

			return
		}

		assert.NoError(tt, dc.stack.listBucket(ctx), "the object storage does not list the bucket yet")
	}, time.Minute, time.Second)
}

func (s *Stack) recordingPrefix(uid string) string {
	return path.Join(s.bucketPath(), uid) + "/"
}

func (s *Stack) bucketPath() string {
	return "store/" + s.envs[objectStorageBucketEnv]
}

func (s *Stack) listBucket(ctx context.Context) error {
	_, err := s.objectStorage(ctx, "ls", s.bucketPath())

	return err
}

func (s *Stack) objectStorage(ctx context.Context, args ...string) (string, error) {
	storage := s.Service(ServiceObjectStorage)
	if storage == nil {
		return "", errors.New("the stack runs no object storage")
	}

	cmd := append([]string{
		"sh", "-c", `mc alias set store http://localhost:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null && mc "$@"`, "mc",
	}, args...)

	code, output, err := storage.Exec(ctx, cmd, tcexec.Multiplexed())
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(output)
	if err != nil {
		return "", err
	}

	if code != 0 {
		return "", fmt.Errorf("mc %v exited with %d: %s", args, code, body)
	}

	return string(body), nil
}
