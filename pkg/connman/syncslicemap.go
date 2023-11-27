package connman

import "sync"

// SyncSliceMap is a struct that uses sync.Map and stores values in a slice.
type SyncSliceMap struct {
	syncMap sync.Map
}

// Load retrieves the last value associated with the key in the slice along with the size of the slice.
func (ssm *SyncSliceMap) Load(key interface{}) (interface{}, bool) {
	if values, ok := ssm.syncMap.Load(key); ok {
		if valSlice := values.([]interface{}); len(valSlice) > 0 {
			return valSlice[len(valSlice)-1], true
		}
	}

	return nil, false
}

// Store appends the value to the slice associated with the key and updates the value in sync.Map.
// It returns the size of the slice after the append operation.
func (ssm *SyncSliceMap) Store(key, value interface{}) {
	ssm.syncMap.LoadOrStore(key, []interface{}{})
	ssm.syncMap.Store(key, append(ssm.getValues(key), value))
}

// Delete removes the value from the slice associated with the key and updates the value in sync.Map.
func (ssm *SyncSliceMap) Delete(key, value interface{}) {
	if values, ok := ssm.syncMap.Load(key); ok {
		var updatedValues []interface{}
		for _, v := range values.([]interface{}) {
			if v != value {
				updatedValues = append(updatedValues, v)
			}
		}

		ssm.syncMap.Store(key, updatedValues)
	}
}

// Size returns the current size of the slice associated with the key.
func (ssm *SyncSliceMap) Size(key interface{}) int {
	return len(ssm.getValues(key))
}

// getValues returns the slice of values associated with the key.
func (ssm *SyncSliceMap) getValues(key interface{}) []interface{} {
	if values, ok := ssm.syncMap.Load(key); ok {
		return values.([]interface{})
	}

	return nil
}
