package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/sftp"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const recordingArchiveCron = "* * * * *"

// TestEnterpriseSessionRecording covers the server-side recording of an enterprise instance: which
// sessions it records, playing a recording back seat by seat, the job that archives a closed
// session's recording to object storage, deleting a recording, and retention pruning the archived
// recordings with the sessions. The cases share one stack, one device and one namespace that
// records, and a case that turns recording off turns it back on.
func TestEnterpriseSessionRecording(t *testing.T) {
	ctx := context.Background()

	compose := newEnterpriseEnvironment(t, ctx,
		environment.New(t, run).WithEnv("SHELLHUB_SESSION_RETENTION_DAYS", "1").WithCronTrigger())
	compose.SetSessionRecording(t, ShellHubNamespace, true)

	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	fixture := &recordingFixture{compose: compose, device: device, signer: signer}

	t.Run("which sessions are recorded", func(t *testing.T) { testWhichSessionsAreRecorded(t, fixture) })
	t.Run("playback", func(t *testing.T) { testRecordingPlayback(t, fixture) })
	t.Run("archive", func(t *testing.T) { testRecordingArchive(t, fixture) })
	t.Run("deletion", func(t *testing.T) { testRecordingDeletion(t, fixture) })
	t.Run("retention", func(t *testing.T) { testRecordingRetention(t, fixture) })
}

func testWhichSessionsAreRecorded(t *testing.T, fixture *recordingFixture) {
	t.Helper()

	compose := fixture.compose

	t.Run("an interactive shell is recorded", func(t *testing.T) {
		uid := fixture.record(t, "shell")

		assert.True(t, getSession(t, t.Context(), compose, uid).Recorded)
		assert.Contains(t, fixture.playback(t, uid, 0), "recorded-shell")
	})

	cases := []struct {
		name       string
		recordsOff bool
		open       func(t *testing.T, conn *ssh.Client, uid string)
	}{
		{
			name: "a command run without a terminal is not recorded",
			open: func(t *testing.T, conn *ssh.Client, _ string) {
				t.Helper()

				sess, err := conn.NewSession()
				require.NoError(t, err)

				output, err := sess.CombinedOutput("printf 'recorded-%s\\n' exec")
				require.NoError(t, err)
				require.Contains(t, string(output), "recorded-exec")
			},
		},
		{
			name: "an SFTP transfer is not recorded",
			open: func(t *testing.T, conn *ssh.Client, _ string) {
				t.Helper()

				client, err := sftp.NewClient(conn)
				require.NoError(t, err)

				file, err := client.OpenFile("/tmp/recorded-by-sftp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
				require.NoError(t, err)

				_, err = file.Write([]byte("sent"))
				require.NoError(t, err)

				require.NoError(t, file.Close())
				require.NoError(t, client.Close())
			},
		},
		{
			name: "an scp copy is not recorded",
			open: func(t *testing.T, conn *ssh.Client, _ string) {
				t.Helper()

				client, err := scp.NewClientBySSH(conn)
				require.NoError(t, err)

				file := bytes.NewBuffer(make([]byte, 64))
				require.NoError(t, client.CopyFilePassThru(t.Context(), file, "/tmp/recorded-by-scp", "0644", io.LimitReader))

				client.Close()
			},
		},
		{
			name:       "a shell is not recorded while the namespace records nothing",
			recordsOff: true,
			open: func(t *testing.T, conn *ssh.Client, _ string) {
				t.Helper()

				openShell(t, conn).prints(t, "unrecorded")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.recordsOff {
				stopRecording(t, compose)
			}

			uid := fixture.finish(t, tc.open)

			assert.False(t, getSession(t, t.Context(), compose, uid).Recorded)
			assert.Zero(t, compose.SessionEventCountOfType(t, uid, models.SessionEventTypePtyOutput),
				"no terminal output was kept")
		})
	}
}

func testRecordingPlayback(t *testing.T, fixture *recordingFixture) {
	t.Helper()

	compose := fixture.compose

	t.Run("each seat of a connection plays back only its own terminal", func(t *testing.T) {
		uid := fixture.finish(t, func(t *testing.T, conn *ssh.Client, _ string) {
			t.Helper()

			first := openShell(t, conn)
			second := openShell(t, conn)

			first.prints(t, "first-seat")
			second.prints(t, "second-seat")
		})

		assert.ElementsMatch(t, []int{0, 1}, sessionDetail(t, t.Context(), compose, uid).Events.Seats)

		first := fixture.playback(t, uid, 0)
		assert.Contains(t, first, "recorded-first-seat")
		assert.NotContains(t, first, "recorded-second-seat")

		second := fixture.playback(t, uid, 1)
		assert.Contains(t, second, "recorded-second-seat")
		assert.NotContains(t, second, "recorded-first-seat")
	})

	t.Run("a session that was not recorded has no recording to play", func(t *testing.T) {
		stopRecording(t, compose)

		uid := fixture.record(t, "unrecorded")

		fixture.requireNoPlayback(t, uid, 0)
	})

	t.Run("a session without a terminal has no recording to play", func(t *testing.T) {
		uid := finishSession(t, t.Context(), compose, fixture.device, fixture.signer, "uptime")

		fixture.requireNoPlayback(t, uid, 0)
	})
}

func testRecordingArchive(t *testing.T, fixture *recordingFixture) {
	t.Helper()

	compose := fixture.compose

	t.Run("the archive job moves a closed session's terminal output to object storage", func(t *testing.T) {
		uid := fixture.finish(t, func(t *testing.T, conn *ssh.Client, uid string) {
			t.Helper()

			openShell(t, conn).prints(t, "archived")

			compose.AwaitSessionEventOfType(t, uid, models.SessionEventTypePtyOutput)
		})

		compose.RunCron(t, recordingArchiveCron)
		fixture.awaitArchived(t, uid, "0.asciinema")

		assert.Zero(t, compose.SessionEventCountOfType(t, uid, models.SessionEventTypePtyOutput),
			"the archived output leaves the database")
		assert.Positive(t, compose.SessionEventCount(t, uid), "the session keeps its timeline")
		assert.Contains(t, fixture.playback(t, uid, 0), "recorded-archived",
			"the recording plays back from object storage")
	})

	t.Run("a session that fails to archive is skipped and retried on the next run", func(t *testing.T) {
		mark := compose.ServerLogMark(t)

		var broken string

		healthy := fixture.finish(t, func(t *testing.T, conn *ssh.Client, _ string) {
			t.Helper()

			openShell(t, conn).prints(t, "not-held-up")

			broken = fixture.finish(t, func(t *testing.T, conn *ssh.Client, uid string) {
				t.Helper()

				openShell(t, conn).prints(t, "retried")

				compose.AwaitSessionEventOfType(t, uid, models.SessionEventTypePtyOutput)
				compose.BreakSessionPtyRequest(t, uid)
			})
		})

		compose.RunCron(t, recordingArchiveCron)
		fixture.awaitArchived(t, healthy, "0.asciinema")
		compose.AwaitServerLogLine(t, mark, "failed to archive session", broken)

		assert.False(t, compose.SessionConverted(t, broken))
		assert.Empty(t, compose.RecordingObjects(t, broken))
		assert.Positive(t, compose.SessionEventCountOfType(t, broken, models.SessionEventTypePtyOutput),
			"the output of a session that failed to archive stays in the database")

		compose.RepairSessionPtyRequest(t, broken)
		compose.RunCron(t, recordingArchiveCron)
		fixture.awaitArchived(t, broken, "0.asciinema")

		assert.Contains(t, fixture.playback(t, broken, 0), "recorded-retried")
	})

	t.Run("each seat is archived to an object of its own", func(t *testing.T) {
		uid := fixture.finish(t, func(t *testing.T, conn *ssh.Client, _ string) {
			t.Helper()

			first := openShell(t, conn)
			second := openShell(t, conn)

			first.prints(t, "first-archived")
			second.prints(t, "second-archived")
		})

		compose.RunCron(t, recordingArchiveCron)
		fixture.awaitArchived(t, uid, "0.asciinema", "1.asciinema")

		second := fixture.playback(t, uid, 1)
		assert.Contains(t, second, "recorded-second-archived")
		assert.NotContains(t, second, "recorded-first-archived")
	})
}

func testRecordingDeletion(t *testing.T, fixture *recordingFixture) {
	t.Helper()

	compose := fixture.compose

	t.Run("deleting a recording removes its object and clears the recorded flag", func(t *testing.T) {
		uid := fixture.archived(t, "deleted")

		fixture.requireDeleted(t, uid)

		assert.Empty(t, compose.RecordingObjects(t, uid))
		fixture.requireNoPlayback(t, uid, 0)
	})

	t.Run("deleting a recording whose object is already gone clears the recorded flag", func(t *testing.T) {
		uid := fixture.archived(t, "gone")

		compose.RemoveRecordingObject(t, uid, 0)

		fixture.requireDeleted(t, uid)
	})
}

func testRecordingRetention(t *testing.T, fixture *recordingFixture) {
	t.Helper()

	compose := fixture.compose

	t.Run("a session past the retention window takes its archived recording with it", func(t *testing.T) {
		uid := fixture.archived(t, "expired")

		compose.AgeSession(t, uid, 48*time.Hour)
		compose.RunCron(t, sessionRetentionCron)

		awaitSessionDeleted(t, t.Context(), compose, uid)

		assert.Empty(t, compose.RecordingObjects(t, uid), "the recording goes with the session")
	})

	t.Run("a session whose recording cannot be purged outlives the retention window until it can", func(t *testing.T) {
		recorded := fixture.archived(t, "unpurged")
		unrecorded := finishSession(t, t.Context(), compose, fixture.device, fixture.signer, "true")

		compose.StopObjectStorage(t)

		compose.AgeSession(t, recorded, 48*time.Hour)
		compose.AgeSession(t, unrecorded, 48*time.Hour)

		compose.RunCron(t, sessionRetentionCron)

		awaitSessionDeleted(t, t.Context(), compose, unrecorded)

		assert.True(t, getSession(t, t.Context(), compose, recorded).Recorded,
			"a session whose recording is still stored is kept")

		compose.StartObjectStorage(t)
		compose.RunCron(t, sessionRetentionCron)

		awaitSessionDeleted(t, t.Context(), compose, recorded)

		assert.Empty(t, compose.RecordingObjects(t, recorded), "the recording goes with the session")
	})
}

func stopRecording(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	compose.SetSessionRecording(t, ShellHubNamespace, false)
	t.Cleanup(func() { compose.SetSessionRecording(t, ShellHubNamespace, true) })
}

type recordingFixture struct {
	compose *environment.DockerCompose
	device  *models.Device
	signer  ssh.Signer
}

func (f *recordingFixture) finish(t *testing.T, open func(t *testing.T, conn *ssh.Client, uid string)) string {
	t.Helper()

	before := currentSessions(t, t.Context(), f.compose)

	var conn *ssh.Client

	opened := sessionAfter(t, t.Context(), f.compose, before, func() {
		conn = dialDevice(t, t.Context(), f.compose, f.device, f.signer)
		t.Cleanup(func() { _ = conn.Close() })
	})

	open(t, conn, opened.UID)

	require.NoError(t, conn.Close())
	requireSessionActive(t, t.Context(), f.compose, opened.UID, false)

	return opened.UID
}

func (f *recordingFixture) record(t *testing.T, word string) string {
	t.Helper()

	return f.finish(t, func(t *testing.T, conn *ssh.Client, _ string) {
		t.Helper()

		openShell(t, conn).prints(t, word)
	})
}

func (f *recordingFixture) archived(t *testing.T, word string) string {
	t.Helper()

	uid := f.record(t, word)

	f.compose.RunCron(t, recordingArchiveCron)
	f.awaitArchived(t, uid, "0.asciinema")

	return uid
}

func (f *recordingFixture) awaitArchived(t *testing.T, uid string, objects ...string) {
	t.Helper()

	f.compose.AwaitSessionConverted(t, uid)
	assert.Equal(t, objects, f.compose.RecordingObjects(t, uid))
}

func (f *recordingFixture) playback(t *testing.T, uid string, seat int) string {
	t.Helper()

	resp, err := f.compose.GetSessionRecord(t.Context(), uid, seat)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return castOutput(t, resp.String())
}

func (f *recordingFixture) requireNoPlayback(t *testing.T, uid string, seat int) {
	t.Helper()

	resp, err := f.compose.GetSessionRecord(t.Context(), uid, seat)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode(), resp.String())
}

func (f *recordingFixture) requireDeleted(t *testing.T, uid string) {
	t.Helper()

	resp, err := f.compose.DeleteSessionRecord(t.Context(), uid, 0)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	assert.False(t, getSession(t, t.Context(), f.compose, uid).Recorded)
}

func castOutput(t *testing.T, cast string) string {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(cast), "\n")

	var output strings.Builder

	for _, line := range lines[1:] {
		event := []any{}
		require.NoError(t, json.Unmarshal([]byte(line), &event), "a cast event is a JSON array: %s", line)
		require.Len(t, event, 3, "a cast event is [time, kind, data]: %s", line)

		if kind, _ := event[1].(string); kind == "o" {
			data, _ := event[2].(string)
			output.WriteString(data)
		}
	}

	return output.String()
}

type shell struct {
	stdin  io.Writer
	output *syncBuffer
}

func openShell(t *testing.T, conn *ssh.Client) *shell {
	t.Helper()

	sess, err := conn.NewSession()
	require.NoError(t, err)

	stdin, err := sess.StdinPipe()
	require.NoError(t, err)

	output := new(syncBuffer)
	sess.Stdout = output

	require.NoError(t, sess.RequestPty("xterm", 24, 80, ssh.TerminalModes{ssh.ECHO: 1}))
	require.NoError(t, sess.Shell())

	return &shell{stdin: stdin, output: output}
}

func (s *shell) prints(t *testing.T, word string) {
	t.Helper()

	_, err := s.stdin.Write([]byte("printf 'recorded-%s\\n' " + word + "\n"))
	require.NoError(t, err)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		assert.Contains(tt, s.output.String(), "recorded-"+word)
	}, 30*time.Second, 200*time.Millisecond, "the shell never ran the command")
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}
