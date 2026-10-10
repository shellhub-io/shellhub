package environment

import (
	"testing"
	"time"
)

// ExpireAPIKeyIn moves the expiry of the API key named name to ttl from now, failing t unless
// exactly that key changed. The API sets an expiry only in whole days ahead, so it writes the row
// directly, standing in for the days a real key waits to expire. A key the server has already
// cached keeps the expiry it was cached with, so call it before the key's first use.
func (dc *DockerCompose) ExpireAPIKeyIn(t *testing.T, name string, ttl time.Duration) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE api_keys SET expires_in = extract(epoch FROM now() + :'ttl'::interval)::bigint WHERE name = :'name'",
		map[string]string{"name": name, "ttl": interval(ttl)})
}

// ExpireInstanceAPIKey moves the expiry of the instance API key named name one minute into the
// past, failing t unless exactly that key changed. The admin API sets an expiry only in whole days
// ahead, so it writes the row directly, standing in for the days a real key waits to expire.
func (dc *DockerCompose) ExpireInstanceAPIKey(t *testing.T, name string) {
	t.Helper()

	dc.updateOne(t,
		"UPDATE instance_api_keys SET expires_at = now() - interval '1 minute' WHERE name = :'name'",
		map[string]string{"name": name})
}
