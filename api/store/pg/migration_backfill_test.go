package pg_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/store/storetest/pgprovider"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// backfillSQL reads migration 014's up statement from disk so the test exercises the exact SQL that
// ships, not a copy that could drift from it.
func backfillSQL(t *testing.T) string {
	t.Helper()
	_, self, _, ok := runtime.Caller(0)
	require.True(t, ok)

	path := filepath.Join(filepath.Dir(self), "migrations", "014_backfill_install_key_events.tx.up.sql")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

// TestInstallKeyEventBackfillMigration covers migration 014: it writes exactly one registration event
// for every device attributed to a key but missing one, freezing the decision for already accepted or
// rejected devices while leaving pending ones open (NULL decided_status, which is what surfaces the
// accept control), and it never touches a device that already has an event or one with no key.
func TestInstallKeyEventBackfillMigration(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { provider.Close(t) })

	st := provider.Store()

	owner, err := st.UserCreate(ctx, &models.User{
		Origin:        models.UserOriginLocal,
		Status:        models.UserStatusConfirmed,
		MaxNamespaces: -1,
		UserData:      models.UserData{Name: "owner", Email: "owner@example.com", Username: "owner"},
		Password:      models.UserPassword{Hash: "hash"},
	})
	require.NoError(t, err)

	// NamespaceCreate seeds the namespace's legacy (system) key; the test devices enroll against it.
	tenant, err := st.NamespaceCreate(ctx, &models.Namespace{
		Name:       "ns",
		Owner:      owner,
		MaxDevices: -1,
		Members:    []models.Member{{ID: owner, Role: authorizer.RoleOwner}},
		Settings:   &models.NamespaceSettings{},
	})
	require.NoError(t, err)

	legacy, err := st.InstallKeyResolveSystem(ctx, tenant)
	require.NoError(t, err)

	now := clock.Now()
	mkDevice := func(uidSeed, mac string, status models.DeviceStatus, keyID string) string {
		t.Helper()
		uid := uidSeed + strings.Repeat("0", 64-len(uidSeed))
		_, err := st.DeviceCreate(ctx, &models.Device{
			UID:             uid,
			TenantID:        tenant,
			Name:            uidSeed,
			Identity:        &models.DeviceIdentity{MAC: mac},
			Info:            &models.DeviceInfo{ID: "arch", PrettyName: "Arch Linux", Version: "v1.2.3", Arch: "amd64", Platform: "docker"},
			PublicKey:       "pk-" + uidSeed,
			Status:          status,
			StatusUpdatedAt: now,
			InstallKeyID:    keyID,
		})
		require.NoError(t, err)

		return uid
	}

	pendingUID := mkDevice("aa", "aa:bb:cc:dd:ee:01", models.DeviceStatusPending, legacy.ID)
	acceptedUID := mkDevice("bb", "aa:bb:cc:dd:ee:02", models.DeviceStatusAccepted, legacy.ID)
	withEventUID := mkDevice("cc", "aa:bb:cc:dd:ee:03", models.DeviceStatusAccepted, legacy.ID)
	keylessUID := mkDevice("dd", "aa:bb:cc:dd:ee:04", models.DeviceStatusPending, "")

	// The device that already has an event must be left untouched by the backfill.
	require.NoError(t, st.InstallKeyEventCreate(ctx, &models.InstallKeyEvent{
		InstallKeyID: legacy.ID,
		TenantID:     tenant,
		DeviceUID:    withEventUID,
		Hostname:     "cc",
	}))

	_, err = provider.DB().ExecContext(ctx, backfillSQL(t))
	require.NoError(t, err)

	type row struct {
		DeviceUID  string `bun:"device_uid"`
		Decided    string `bun:"decided_status"`
		InfoID     string `bun:"info_id"`
		InfoPretty string `bun:"info_pretty_name"`
		InfoArch   string `bun:"info_arch"`
	}
	events := make(map[string][]row)
	var rows []row
	require.NoError(t, provider.DB().
		NewRaw("SELECT device_uid, coalesce(decided_status, '') AS decided_status, coalesce(info_id, '') AS info_id, coalesce(info_pretty_name, '') AS info_pretty_name, coalesce(info_arch, '') AS info_arch FROM install_key_events").
		Scan(ctx, &rows))
	for _, r := range rows {
		events[r.DeviceUID] = append(events[r.DeviceUID], r)
	}

	require.Len(t, events[pendingUID], 1, "a pending device without an event gets one")
	assert.Empty(t, events[pendingUID][0].Decided, "a pending device's backfilled event stays open so the accept control shows")
	// The device's OS facts must ride along so the registration activity shows the distro icon and name,
	// not a blank generic row: info_id drives the icon, info_pretty_name/arch the labels.
	assert.Equal(t, "arch", events[pendingUID][0].InfoID, "the device's distro id (identifier) is copied so the icon renders")
	assert.Equal(t, "Arch Linux", events[pendingUID][0].InfoPretty)
	assert.Equal(t, "amd64", events[pendingUID][0].InfoArch)

	require.Len(t, events[acceptedUID], 1, "an accepted device without an event gets one")
	assert.Equal(t, "accepted", events[acceptedUID][0].Decided, "an accepted device's decision is frozen on the event")
	assert.Equal(t, "arch", events[acceptedUID][0].InfoID, "OS facts ride along regardless of decision")

	require.Len(t, events[withEventUID], 1, "a device that already had an event is not given a second one")

	assert.Empty(t, events[keylessUID], "a device with no key is skipped")
}
