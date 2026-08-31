package web

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManagerSave(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description string
		id          string
		waitFor     time.Duration
		data        *Credentials
	}{
		{
			description: "insert credential on manager and delete after 1 second",
			id:          "foo",
			waitFor:     1 * time.Second,
			data:        nil,
		},
		{
			description: "insert credential on manager and delete after 2 seconds",
			id:          "bar",
			waitFor:     2 * time.Second,
			data:        nil,
		},
	}

	for _, ts := range tests {
		test := ts

		t.Run(test.description, func(t *testing.T) {
			t.Parallel()

			manager := newManager(test.waitFor)
			manager.save(test.id, nil)

			assert.EventuallyWithT(t, func(tt *assert.CollectT) {
				_, ok := manager.get(test.id)
				assert.False(tt, ok)
			}, 2*test.waitFor, 10*time.Millisecond)
		})
	}
}
