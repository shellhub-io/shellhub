// Package telemetry sends usage events to the ShellHub telemetry collector, which speaks
// PostHog's capture v1 protocol. It is the only package that imports posthog-go.
package telemetry

import (
	"maps"
	"time"

	"github.com/posthog/posthog-go"
	log "github.com/sirupsen/logrus"
)

// DefaultEndpoint is the collector an instance reports to. Released binaries carry it, so it
// never changes.
const DefaultEndpoint = "https://telemetry.shellhub.io"

// SchemaVersion is sent as schema_version on every event. Raise it when a property changes
// meaning or is removed, so the collector can tell old payloads from new ones.
const SchemaVersion = 1

// CommunityWriteKey and EnterpriseWriteKey are the public write keys the Community and
// Enterprise binaries send. They are not secrets: they filter stray traffic and name the edition.
// Cloud's key is a secret its deployment provides.
const (
	CommunityWriteKey  = "shellhub_community_f7319c57cf0f033d6c7128f3eafc4b4b"
	EnterpriseWriteKey = "shellhub_enterprise_277b7aebd2fbd069ce337546bf36adf2"
)

const shutdownTimeout = 3 * time.Second

// Event is one telemetry event. Name is an object and a past-tense verb, written as a literal,
// and never starts with "$", which PostHog reserves.
type Event struct {
	Name       string
	DistinctID string
	Properties map[string]any
	Groups     map[string]string
}

// ClientProperties are the properties posthog-go adds to every event on its own, besides the
// uuid and timestamp. The public telemetry page lists them; $os_version and $os_distro are sent
// only when the host reports them.
var ClientProperties = []string{"$os", "$os_version", "$os_distro", "$go_version", "$is_server", "$geoip_disable"}

// Preview is the event as the collector receives it, without the uuid, the timestamp and the
// [ClientProperties] posthog-go adds when it sends it.
type Preview struct {
	Event      string         `json:"event"`
	DistinctID string         `json:"distinct_id"`
	Options    map[string]any `json:"options"`
	Properties map[string]any `json:"properties"`
}

// Preview returns the event as posthog-go sends it: schema_version added, $groups added when the
// event has groups, and $process_person_profile moved out of Properties into Options.
func (e Event) Preview() Preview {
	properties := e.properties()
	delete(properties, "$process_person_profile")

	return Preview{
		Event:      e.Name,
		DistinctID: e.DistinctID,
		Options:    map[string]any{"process_person_profile": false},
		Properties: properties,
	}
}

func (e Event) properties() map[string]any {
	properties := make(map[string]any, len(e.Properties)+3)
	maps.Copy(properties, e.Properties)

	properties["schema_version"] = SchemaVersion
	properties["$process_person_profile"] = false

	if len(e.Groups) > 0 {
		groups := make(map[string]any, len(e.Groups))
		for name, key := range e.Groups {
			groups[name] = key
		}

		properties["$groups"] = groups
	}

	return properties
}

// Client queues events and sends them to the collector in the background.
type Client struct {
	posthog  posthog.Client
	endpoint string
}

// New returns a client that sends to endpoint, authenticating with the edition's write key. It
// returns an error when posthog-go rejects the configuration.
func New(endpoint, writeKey string) (*Client, error) {
	client, err := posthog.NewWithConfig(writeKey, posthog.Config{
		Endpoint:        endpoint,
		CaptureMode:     posthog.CaptureModeAnalyticsV1,
		Compression:     posthog.CompressionGzip,
		ShutdownTimeout: shutdownTimeout,
		Callback:        callback{},
		Logger:          logger{},
	})
	if err != nil {
		return nil, err
	}

	return &Client{posthog: client, endpoint: endpoint}, nil
}

// Endpoint returns the collector the client sends to.
func (c *Client) Endpoint() string {
	return c.endpoint
}

// Send queues events for delivery. It returns the first error posthog-go reports, such as
// posthog.ErrQueueFull, after trying to queue every event. A queued event the collector drops or
// never receives is logged at debug level, not returned.
func (c *Client) Send(events ...Event) error {
	var first error

	for _, event := range events {
		err := c.posthog.Enqueue(posthog.Capture{
			DistinctId: event.DistinctID,
			Event:      event.Name,
			Properties: event.properties(),
		})
		if err != nil && first == nil {
			first = err
		}
	}

	return first
}

// Close flushes the queued events and stops the client, giving up after a few seconds so a
// collector that does not answer cannot hold up a shutdown. It returns the error posthog-go's
// Close reports when the queue cannot be flushed in time or the client is already closed.
func (c *Client) Close() error {
	return c.posthog.Close()
}

type callback struct{}

func (callback) Success(posthog.APIMessage) {}

func (callback) Failure(message posthog.APIMessage, err error) {
	entry := log.WithError(err)
	if capture, ok := message.(posthog.CaptureInApi); ok {
		entry = entry.WithField("event", capture.Event)
	}

	entry.Debug("the telemetry collector did not accept an event")
}

type logger struct{}

func (logger) Debugf(format string, args ...any) { log.Debugf("telemetry: "+format, args...) }
func (logger) Logf(format string, args ...any)   { log.Debugf("telemetry: "+format, args...) }
func (logger) Warnf(format string, args ...any)  { log.Debugf("telemetry: "+format, args...) }
func (logger) Errorf(format string, args ...any) { log.Debugf("telemetry: "+format, args...) }
