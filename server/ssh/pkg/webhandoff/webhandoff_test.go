package webhandoff_test

import (
	"sync"
	"testing"

	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_carriesTheDataAcrossTheDial(t *testing.T) {
	tests := []struct {
		description string
		data        webhandoff.Data
	}{
		{
			description: "an ipv4 browser",
			data:        webhandoff.Data{Device: "device-uid", IP: "192.168.0.10"},
		},
		{
			description: "an ipv6 browser",
			data:        webhandoff.Data{Device: "device-uid", IP: "2001:db8::1"},
		},
		{
			description: "identity mode, carrying the logged-in account",
			data:        webhandoff.Data{Device: "device-uid", IP: "192.168.0.10", UserID: "user-id"},
		},
		{
			description: "legacy web session, with no account",
			data:        webhandoff.Data{Device: "device-uid", IP: "192.168.0.10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			store := webhandoff.NewStore()

			store.Put("user@uuid", tt.data)

			got, ok := store.Take("user@uuid")
			require.True(t, ok)
			assert.Equal(t, tt.data, got)
		})
	}
}

func TestStore_takeIsSingleUse(t *testing.T) {
	store := webhandoff.NewStore()

	store.Put("user@uuid", webhandoff.Data{Device: "device-uid", IP: "192.168.0.10"})

	_, ok := store.Take("user@uuid")
	require.True(t, ok)

	_, ok = store.Take("user@uuid")
	assert.False(t, ok)
}

func TestStore_takeReportsAnUnknownID(t *testing.T) {
	store := webhandoff.NewStore()

	data, ok := store.Take("never-put")

	assert.False(t, ok)
	assert.Equal(t, webhandoff.Data{}, data)
}

func TestStore_onlyOneConcurrentTakeWins(t *testing.T) {
	store := webhandoff.NewStore()

	store.Put("user@uuid", webhandoff.Data{Device: "device-uid", IP: "192.168.0.10"})

	const racers = 32

	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		wins  int
		start = make(chan struct{})
	)

	for range racers {
		wg.Go(func() {
			<-start

			if _, ok := store.Take("user@uuid"); ok {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		})
	}

	close(start)
	wg.Wait()

	assert.Equal(t, 1, wins, "exactly one caller must claim a handoff")
}
