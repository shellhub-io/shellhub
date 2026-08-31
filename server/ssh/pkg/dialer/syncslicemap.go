package dialer

import (
	"slices"
	"sync"
)

// SyncSliceMap is a map that holds a slice of values per key, safe for
// concurrent use.
//
// Its mutating operations report what they observed under the lock: Store
// returns the values it displaced and Delete the count that survived. Callers
// need both to decide whether a key has just gained a duplicate or lost its
// last value, and reading them back afterwards would race with the next
// caller.
type SyncSliceMap struct {
	mu     sync.RWMutex
	values map[any][]any

	total int
}

// Stats returns how many keys the map holds and how many values across all of
// them, under one read lock so the two describe the same instant. They differ
// only while some key holds more than one value.
func (ssm *SyncSliceMap) Stats() (keys, values int) {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	return len(ssm.values), ssm.total
}

// Load retrieves the most recently stored value for the key.
func (ssm *SyncSliceMap) Load(key any) (any, bool) {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	if values := ssm.values[key]; len(values) > 0 {
		return values[len(values)-1], true
	}

	return nil, false
}

// Store appends the value to the key's slice and returns the values that were
// already stored under it.
func (ssm *SyncSliceMap) Store(key, value any) []any {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	if ssm.values == nil {
		ssm.values = make(map[any][]any)
	}

	displaced := slices.Clone(ssm.values[key])
	ssm.values[key] = append(ssm.values[key], value)
	ssm.total++

	return displaced
}

// Delete removes the value from the key's slice and returns how many values
// remain under it. The key itself is dropped once its last value is gone, so
// the map does not retain an entry for every device ever seen.
func (ssm *SyncSliceMap) Delete(key, value any) int {
	ssm.mu.Lock()
	defer ssm.mu.Unlock()

	stored := ssm.values[key]
	remaining := slices.DeleteFunc(stored, func(v any) bool {
		return v == value
	})

	ssm.total -= len(stored) - len(remaining)

	if len(remaining) == 0 {
		delete(ssm.values, key)

		return 0
	}

	ssm.values[key] = remaining

	return len(remaining)
}

// Size returns the current size of the slice associated with the key.
func (ssm *SyncSliceMap) Size(key any) int {
	ssm.mu.RLock()
	defer ssm.mu.RUnlock()

	return len(ssm.values[key])
}
