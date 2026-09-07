package storetest

// Suite holds the generic store tests Groups registers, and the provider they run against
type Suite struct {
	provider StoreProvider
}

// NewSuite creates a new test suite with the given provider
func NewSuite(provider StoreProvider) *Suite {
	return &Suite{provider: provider}
}
