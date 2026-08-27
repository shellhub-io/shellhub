package dialer

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	type Expected struct {
		result any
		status bool
	}

	cases := []struct {
		title    string
		setup    func() *SyncSliceMap
		key      string
		expected Expected
	}{
		{
			title: "fails when loading from an empty map",
			setup: func() *SyncSliceMap {
				return &SyncSliceMap{}
			},
			key: "",
			expected: Expected{
				result: nil,
				status: false,
			},
		},
		{
			title: "loads the last value when the key holds multiple",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("keys", "value1")
				ssm.Store("keys", "value2")
				ssm.Store("keys", "value3")

				return ssm
			},
			key: "keys",
			expected: Expected{
				result: "value3",
				status: true,
			},
		},
		{
			title: "fails when the key held a value that was deleted",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key", "value1")
				ssm.Delete("key", "value1")

				return ssm
			},
			key: "key",
			expected: Expected{
				result: nil,
				status: false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ssm := tc.setup()
			result, ok := ssm.Load(tc.key)

			assert.Equal(t, tc.expected, Expected{result, ok})
		})
	}
}

func TestStore(t *testing.T) {
	cases := []struct {
		title             string
		setup             func() *SyncSliceMap
		key               string
		value             any
		expectedDisplaced []any
		expectedSize      int
	}{
		{
			title: "reports nothing displaced when storing under a new key",
			setup: func() *SyncSliceMap {
				return &SyncSliceMap{}
			},
			key:               "key",
			value:             "value1",
			expectedDisplaced: nil,
			expectedSize:      1,
		},
		{
			title: "reports the values already held by the key",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key", "value1")
				ssm.Store("key", "value2")

				return ssm
			},
			key:               "key",
			value:             "value3",
			expectedDisplaced: []any{"value1", "value2"},
			expectedSize:      3,
		},
		{
			title: "reports nothing displaced for a key whose values were deleted",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key", "value1")
				ssm.Delete("key", "value1")

				return ssm
			},
			key:               "key",
			value:             "value2",
			expectedDisplaced: nil,
			expectedSize:      1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ssm := tc.setup()

			assert.Equal(t, tc.expectedDisplaced, ssm.Store(tc.key, tc.value))
			assert.Equal(t, tc.expectedSize, ssm.Size(tc.key))

			value, ok := ssm.Load(tc.key)
			assert.True(t, ok)
			assert.Equal(t, tc.value, value)
		})
	}
}

func TestDelete(t *testing.T) {
	cases := []struct {
		title             string
		setup             func() *SyncSliceMap
		key               string
		valueToDelete     any
		expectedRemaining int
	}{
		{
			title: "deletes a value from a key that holds several",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("existingKey", "value1.1")
				ssm.Store("existingKey", "value1.2")
				ssm.Store("existingKey", "value1.3")

				return ssm
			},
			key:               "existingKey",
			valueToDelete:     "value1.2",
			expectedRemaining: 2,
		},
		{
			title: "reports no remaining values when the last one is deleted",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key", "value1")

				return ssm
			},
			key:               "key",
			valueToDelete:     "value1",
			expectedRemaining: 0,
		},
		{
			title: "reports no remaining values for an unknown key",
			setup: func() *SyncSliceMap {
				return &SyncSliceMap{}
			},
			key:               "missing",
			valueToDelete:     "value1",
			expectedRemaining: 0,
		},
		{
			title: "leaves the key untouched when the value is not held by it",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key", "value1")

				return ssm
			},
			key:               "key",
			valueToDelete:     "other",
			expectedRemaining: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ssm := tc.setup()

			assert.Equal(t, tc.expectedRemaining, ssm.Delete(tc.key, tc.valueToDelete))
			assert.Equal(t, tc.expectedRemaining, ssm.Size(tc.key))
		})
	}
}

// TestDeleteDropsTheKey guards the map against growing one entry per device
// ever connected: a key whose last value is gone must leave nothing behind.
func TestDeleteDropsTheKey(t *testing.T) {
	ssm := &SyncSliceMap{}
	ssm.Store("key", "value1")
	ssm.Delete("key", "value1")

	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	assert.NotContains(t, ssm.values, "key")
}

func TestSize(t *testing.T) {
	cases := []struct {
		title        string
		setup        func() *SyncSliceMap
		key          string
		expectedSize int
	}{
		{
			title: "getting size of an empty slice",
			setup: func() *SyncSliceMap {
				return &SyncSliceMap{}
			},
			key:          "",
			expectedSize: 0,
		},
		{
			title: "getting size of a slice with multiple values",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("keys", "value1")
				ssm.Store("keys", "value2")
				ssm.Store("keys", "value3")

				return ssm
			},
			key:          "keys",
			expectedSize: 3,
		},
		{
			title: "getting size of a slice after adding a new value",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key", "newValue")

				return ssm
			},
			key:          "key",
			expectedSize: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ssm := tc.setup()

			size := ssm.Size(tc.key)
			assert.Equal(t, tc.expectedSize, size)
		})
	}
}

// TestConcurrentStoreKeepsEveryValue is the reason this type is not a bare
// sync.Map: a read-modify-write of the key's slice loses one of two concurrent
// stores, and a connection the manager never recorded is one it never tears
// down. Run under -race.
func TestConcurrentStoreKeepsEveryValue(t *testing.T) {
	const goroutines = 50

	ssm := &SyncSliceMap{}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func() {
			defer wg.Done()
			ssm.Store("key", fmt.Sprintf("value%d", i))
		}()
	}

	wg.Wait()

	assert.Equal(t, goroutines, ssm.Size("key"))
}

// TestConcurrentStoreAndDeleteReportOneWinner pins the contract the manager
// relies on to decide a device went offline: across concurrent registrations
// and teardowns of one key, exactly one Delete may observe an empty key.
func TestConcurrentStoreAndDeleteReportOneWinner(t *testing.T) {
	const goroutines = 50

	ssm := &SyncSliceMap{}

	values := make([]string, goroutines)
	for i := range values {
		values[i] = fmt.Sprintf("value%d", i)
		ssm.Store("key", values[i])
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		emptied int
	)

	wg.Add(goroutines)

	for _, value := range values {
		go func() {
			defer wg.Done()

			if ssm.Delete("key", value) == 0 {
				mu.Lock()
				emptied++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, 1, emptied, "exactly one teardown must see the key empty")
	assert.Equal(t, 0, ssm.Size("key"))
}

func TestStats(t *testing.T) {
	type Expected struct {
		keys   int
		values int
	}

	cases := []struct {
		title    string
		setup    func() *SyncSliceMap
		expected Expected
	}{
		{
			title: "reports an empty map as zero",
			setup: func() *SyncSliceMap {
				return &SyncSliceMap{}
			},
			expected: Expected{keys: 0, values: 0},
		},
		{
			title: "counts every value held, across keys",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key1", "value1.1")
				ssm.Store("key1", "value1.2")
				ssm.Store("key2", "value2.1")

				return ssm
			},
			expected: Expected{keys: 2, values: 3},
		},
		{
			title: "drops both counts when a key loses its last value",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key1", "value1.1")
				ssm.Store("key2", "value2.1")
				ssm.Delete("key2", "value2.1")

				return ssm
			},
			expected: Expected{keys: 1, values: 1},
		},
		{
			title: "keeps the key when only one of its values goes",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key", "value1")
				ssm.Store("key", "value2")
				ssm.Delete("key", "value1")

				return ssm
			},
			expected: Expected{keys: 1, values: 1},
		},
		{
			title: "ignores a delete of a value the key never held",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key", "value1")
				ssm.Delete("key", "other")
				ssm.Delete("missing", "value1")

				return ssm
			},
			expected: Expected{keys: 1, values: 1},
		},
		{
			// Delete drops every match, so the total must fall by more than one.
			title: "subtracts every copy a single delete removes",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				ssm.Store("key", "value")
				ssm.Store("key", "value")
				ssm.Store("key", "survivor")
				ssm.Delete("key", "value")

				return ssm
			},
			expected: Expected{keys: 1, values: 1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			keys, values := tc.setup().Stats()

			assert.Equal(t, tc.expected, Expected{keys, values})
		})
	}
}

// TestStatsSurvivesConcurrentMutation guards the running total against the map
// it summarizes: a count maintained outside the critical sections drifts, and a
// gauge that drifts is worse than no gauge at all. Run under -race.
func TestStatsSurvivesConcurrentMutation(t *testing.T) {
	const (
		keys       = 20
		perKey     = 10
		goroutines = keys * perKey
	)

	ssm := &SyncSliceMap{}

	mutate := func(op func(key, value any)) {
		var wg sync.WaitGroup

		wg.Add(goroutines)

		for k := range keys {
			for v := range perKey {
				go func() {
					defer wg.Done()
					op(fmt.Sprintf("key%d", k), fmt.Sprintf("value%d.%d", k, v))
				}()
			}
		}

		wg.Wait()
	}

	mutate(func(key, value any) { ssm.Store(key, value) })

	gotKeys, gotValues := ssm.Stats()
	assert.Equal(t, keys, gotKeys)
	assert.Equal(t, goroutines, gotValues)

	mutate(func(key, value any) { ssm.Delete(key, value) })

	gotKeys, gotValues = ssm.Stats()
	assert.Equal(t, 0, gotKeys)
	assert.Equal(t, 0, gotValues)
}
