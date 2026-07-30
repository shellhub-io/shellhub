package magickey

import (
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReference(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "success when called twice returns singleton key",
			test: func(t *testing.T) {
				key1 := GetReference()
				key2 := GetReference()
				assert.Same(t, key1, key2)
			},
		},
		{
			name: "success when key is valid RSA 2048",
			test: func(t *testing.T) {
				key := GetReference()
				assert.NotNil(t, key)
				assert.Equal(t, 2048, key.N.BitLen())
				assert.NotNil(t, key.PublicKey)
				assert.Equal(t, 2048, key.PublicKey.N.BitLen())
			},
		},
		{
			name: "success when key is usable for operations",
			test: func(t *testing.T) {
				key := GetReference()
				assert.NotNil(t, key.Primes)
				assert.Len(t, key.Primes, 2)
				assert.NotNil(t, key.Precomputed)
			},
		},
		{
			name: "success when multiple calls return same key",
			test: func(t *testing.T) {
				keys := make([]*rsa.PrivateKey, 10)
				for i := 0; i < 10; i++ {
					keys[i] = GetReference()
				}
				firstKey := keys[0]
				for _, key := range keys {
					assert.Same(t, firstKey, key)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func TestGetReference_Concurrency(t *testing.T) {
	t.Run("success when called concurrently returns singleton", func(t *testing.T) {
		const numGoroutines = 100
		keys := make(chan *rsa.PrivateKey, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				keys <- GetReference()
			}()
		}
		collectedKeys := make([]*rsa.PrivateKey, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			collectedKeys[i] = <-keys
		}
		firstKey := collectedKeys[0]
		for _, key := range collectedKeys {
			assert.Same(t, firstKey, key)
		}
	})
}

func BenchmarkGetReference(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetReference()
	}
}
