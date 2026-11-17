package pg_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScrubSessionEventPayloadsMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	provider, err := pgprovider.NewProviderAt(ctx, 59)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	db := provider.DB()

	const (
		tenant    = "22222222-2222-4222-8222-222222222222"
		ownerID   = "11111111-1111-4111-8111-111111111111"
		deviceID  = "33333333-3333-4333-8333-333333333333"
		sessionID = "session"
	)

	seedUser(t, ctx, db, ownerID, "owner")
	seedNamespace(t, ctx, db, tenant, "ns", ownerID, fixtureTime)
	seedDevice(t, ctx, db, deviceID, tenant, "dev", "00:00:00:00:00:01", "accepted", fixtureTime)

	execSQL(t, ctx, db, `
		INSERT INTO sessions (id, namespace_id, device_id, username, ip_address,
		                      started_at, seen_at, created_at, updated_at)
		VALUES (?, ?, ?, 'root', '10.0.0.1', now(), now(), now(), now())
	`, sessionID, tenant, deviceID)

	kept := map[string]string{
		"pty-output":                 `{"output":"hello"}`,
		"pty-req":                    `{"term":"xterm"}`,
		"window-change":              `{"columns":80}`,
		"shell":                      `""`,
		"exec":                       `{"command":"uptime"}`,
		"subsystem":                  `{"command":"sftp"}`,
		"signal":                     `"AAAABElOVA=="`,
		"exit-status":                `{"status":0}`,
		"exit-signal":                `{"signal":"KILL"}`,
		"direct-tcpip":               `{"dest_addr":"10.0.0.2"}`,
		"auth-agent-req@openssh.com": `""`,
	}

	event := func(id, eventType, data string) {
		t.Helper()

		execSQL(t, ctx, db, `
			INSERT INTO session_events (id, session_id, type, seat, data, created_at)
			VALUES (?, ?, ?, 0, ?, now())
		`, id, sessionID, eventType, data)
	}

	n := 0
	for eventType, data := range kept {
		n++
		event(fmt.Sprintf("00000000-0000-4000-8000-%012d", n), eventType, data)
	}

	const (
		envID     = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
		unknownID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	)

	event(envID, "env", `"AAAABVRPS0VOAAAABnNlY3JldA=="`)
	event(unknownID, "x11-req", `"YW4geDExIGNvb2tpZQ=="`)

	require.NoError(t, provider.ApplyNext(ctx), "060 must apply cleanly over release-shaped data")

	var envEvents int
	require.NoError(t, db.NewRaw("SELECT count(*) FROM session_events WHERE type = 'env'").Scan(ctx, &envEvents))
	assert.Zero(t, envEvents, "every env event must be deleted")

	var unknownData string
	require.NoError(t, db.NewRaw("SELECT coalesce(data, '') FROM session_events WHERE id = ?", unknownID).Scan(ctx, &unknownData))
	assert.Empty(t, unknownData, "an event of a type the gateway does not decode must lose its payload")

	type row struct {
		Type string `bun:"type"`
		Data string `bun:"data"`
	}

	var rows []row
	require.NoError(t, db.NewRaw("SELECT type, data FROM session_events WHERE type <> 'x11-req'").Scan(ctx, &rows))

	got := make(map[string]string, len(rows))
	for _, r := range rows {
		got[r.Type] = r.Data
	}

	assert.Equal(t, kept, got, "the payloads the gateway decodes must survive")
}
