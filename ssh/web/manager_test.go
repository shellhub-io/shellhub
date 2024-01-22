package web

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManagerSave(t *testing.T) {
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

			manager := NewManager(test.waitFor)
			manager.Save(test.id, nil)

			assert.EventuallyWithT(t, func(tt *assert.CollectT) {
				_, ok := manager.Get(test.id)
				assert.False(tt, ok)

				// NOTICE: we are waiting for two times the defined time, verifying each 10 Millisecond if the condition
				// met. It means that the wait time multiplied by two is the max time, but the codition can be met until
				// its end.
			}, 2*test.waitFor, 10*time.Millisecond)
		})
	}
}
