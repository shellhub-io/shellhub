package connman

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	type Expected struct {
		result interface{}
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
			title: "fails when loading from a map with multiple values",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				key := "keys"
				values := []interface{}{"value1", "value2", "value3"}
				ssm.syncMap.Store(key, values)

				return ssm
			},
			key: "keys",
			expected: Expected{
				result: "value3",
				status: true,
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
	type Expected struct {
		values any
		status bool
	}

	cases := []struct {
		title    string
		setup    func() *SyncSliceMap
		key      string
		value    interface{}
		expected Expected
	}{
		{
			title: "success when storing a value for a new key",
			setup: func() *SyncSliceMap {
				return &SyncSliceMap{}
			},
			key:   "key",
			value: "value1",
			expected: Expected{
				values: []interface{}{"value1"},
				status: true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ssm := tc.setup()
			ssm.Store(tc.key, tc.value)

			result, ok := ssm.syncMap.Load(tc.key)
			assert.Equal(t, tc.expected, Expected{result, ok})
		})
	}
}

func TestDelete(t *testing.T) {
	type Expected struct {
		values any
		status bool
	}

	cases := []struct {
		title         string
		setup         func() *SyncSliceMap
		key           string
		valueToDelete interface{}
		expected      Expected
	}{
		{
			title: "success when try deleting a value from an existing key",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				key := "existingKey"
				values := []interface{}{"value1.1", "value1.2", "value1.3"}
				ssm.syncMap.Store(key, values)

				return ssm
			},
			key:           "existingKey",
			valueToDelete: "value1.2",
			expected: Expected{
				values: []interface{}{"value1.1", "value1.3"},
				status: true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ssm := tc.setup()
			ssm.Delete(tc.key, tc.valueToDelete)

			result, ok := ssm.syncMap.Load(tc.key)
			assert.Equal(t, tc.expected, Expected{result, ok})
		})
	}
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
				key := "keys"
				values := []interface{}{"value1", "value2", "value3"}
				ssm.syncMap.Store(key, values)

				return ssm
			},
			key:          "keys",
			expectedSize: 3,
		},
		{
			title: "getting size of a slice after adding a new value",
			setup: func() *SyncSliceMap {
				ssm := &SyncSliceMap{}
				key := "key"
				valueToAdd := "newValue"
				ssm.syncMap.Store(key, []interface{}{valueToAdd})

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
