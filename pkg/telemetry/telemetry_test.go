package telemetry_test

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/telemetry"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wireEvent struct {
	Event      string         `json:"event"`
	DistinctID string         `json:"distinct_id"`
	UUID       string         `json:"uuid"`
	Options    map[string]any `json:"options"`
	Properties map[string]any `json:"properties"`
}

type request struct {
	path          string
	authorization string
	encoding      string
	createdAt     string
	batch         []wireEvent
}

type collector struct {
	mu       sync.Mutex
	requests []request
	drop     string
	listener *httptest.Server
}

func newCollector(t *testing.T, drop string) *collector {
	t.Helper()

	c := &collector{drop: drop}
	c.listener = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body io.Reader = r.Body
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			body = reader
		}

		var envelope struct {
			CreatedAt string      `json:"created_at"`
			Batch     []wireEvent `json:"batch"`
		}
		if err := json.NewDecoder(body).Decode(&envelope); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		results := map[string]map[string]string{}
		for _, event := range envelope.Batch {
			result := "ok"
			if event.Event == c.drop {
				result = "drop"
			}

			results[event.UUID] = map[string]string{"result": result}
		}

		c.mu.Lock()
		c.requests = append(c.requests, request{
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
			encoding:      r.Header.Get("Content-Encoding"),
			createdAt:     envelope.CreatedAt,
			batch:         envelope.Batch,
		})
		c.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"results": results})
	}))
	t.Cleanup(c.listener.Close)

	return c
}

func (c *collector) events() []wireEvent {
	c.mu.Lock()
	defer c.mu.Unlock()

	var all []wireEvent
	for _, request := range c.requests {
		all = append(all, request.batch...)
	}

	return all
}

func heartbeat() telemetry.Event {
	return telemetry.Event{
		Name:       "instance heartbeat",
		DistinctID: "instance-1",
		Properties: map[string]any{"edition": "community", "devices_accepted": 10},
	}
}

func TestClientSendsCaptureV1(t *testing.T) {
	c := newCollector(t, "")

	client, err := telemetry.New(c.listener.URL, "community-write-key")
	require.NoError(t, err)

	require.NoError(t, client.Send(
		heartbeat(),
		telemetry.Event{
			Name:       "namespace heartbeat",
			DistinctID: "instance-1",
			Groups:     map[string]string{"namespace": "hashed-tenant"},
		},
	))
	require.NoError(t, client.Close())

	c.mu.Lock()
	require.Len(t, c.requests, 1)
	sent := c.requests[0]
	c.mu.Unlock()

	assert.Equal(t, "/i/v1/analytics/events", sent.path)
	assert.Equal(t, "Bearer community-write-key", sent.authorization)
	assert.Equal(t, "gzip", sent.encoding)
	assert.NotEmpty(t, sent.createdAt)

	require.Len(t, sent.batch, 2)

	event := sent.batch[0]
	assert.Equal(t, "instance heartbeat", event.Event)
	assert.Equal(t, "instance-1", event.DistinctID)
	assert.NotEmpty(t, event.UUID)
	assert.Equal(t, "community", event.Properties["edition"])
	assert.InDelta(t, telemetry.SchemaVersion, event.Properties["schema_version"], 0)
	assert.Equal(t, map[string]any{"process_person_profile": false}, event.Options)

	assert.Equal(t, map[string]any{"namespace": "hashed-tenant"}, sent.batch[1].Properties["$groups"])
}

func TestPreviewIsTheEventOnTheWire(t *testing.T) {
	c := newCollector(t, "")

	client, err := telemetry.New(c.listener.URL, "community-write-key")
	require.NoError(t, err)

	require.NoError(t, client.Send(heartbeat()))
	require.NoError(t, client.Close())

	events := c.events()
	require.Len(t, events, 1)
	wire := events[0]

	preview := heartbeat().Preview()

	assert.Equal(t, wire.Event, preview.Event)
	assert.Equal(t, wire.DistinctID, preview.DistinctID)
	assert.Equal(t, wire.Options, preview.Options)

	encoded, err := json.Marshal(preview.Properties)
	require.NoError(t, err)

	var properties map[string]any
	require.NoError(t, json.Unmarshal(encoded, &properties))

	for name, value := range properties {
		assert.Equal(t, value, wire.Properties[name], "property %s", name)
	}

	for name := range wire.Properties {
		if _, ours := properties[name]; !ours {
			assert.Contains(t, telemetry.ClientProperties, name, "posthog-go added a property the telemetry page does not list")
		}
	}
}

func TestClientLogsAnEventTheCollectorDrops(t *testing.T) {
	hook := logtest.NewGlobal()
	level := log.GetLevel()
	log.SetLevel(log.DebugLevel)
	t.Cleanup(func() {
		log.SetLevel(level)
		hook.Reset()
	})

	c := newCollector(t, "instance heartbeat")

	client, err := telemetry.New(c.listener.URL, "community-write-key")
	require.NoError(t, err)

	require.NoError(t, client.Send(heartbeat()))
	require.NoError(t, client.Close())

	var dropped *log.Entry
	for _, entry := range hook.AllEntries() {
		if entry.Message == "the telemetry collector did not accept an event" {
			dropped = entry
		}
	}

	require.NotNil(t, dropped)
	assert.Equal(t, log.DebugLevel, dropped.Level)
	assert.Equal(t, "instance heartbeat", dropped.Data["event"])
}
