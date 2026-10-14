package storetest

import "github.com/shellhub-io/shellhub/pkg/clock"

// Suite holds the generic store tests Groups registers, and the provider they run against
type Suite struct {
	provider StoreProvider
}

// Clock returns the clock the provider's store reads.
func (s *Suite) Clock() *clock.Fixed {
	return s.provider.Clock()
}

// NewSuite creates a new test suite with the given provider
func NewSuite(provider StoreProvider) *Suite {
	return &Suite{provider: provider}
}
