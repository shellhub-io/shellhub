package store

import "github.com/shellhub-io/shellhub/pkg/cache"

// StoreWrapperFactory is a function that wraps a Store with additional
// functionality. Cloud packages register a factory via RegisterStoreWrapper in
// their init() functions so the core server can apply it after creating the
// base store.
type StoreWrapperFactory func(s Store, c cache.Cache) (Store, error)

var storeWrapper StoreWrapperFactory

// RegisterStoreWrapper registers the factory function that wraps the store.
// This must be called before the server's Setup() is invoked — typically from a
// cloud package's init() function.
func RegisterStoreWrapper(f StoreWrapperFactory) {
	storeWrapper = f
}

// StoreWrapper returns the registered StoreWrapperFactory, or nil in CE builds.
func StoreWrapper() StoreWrapperFactory {
	return storeWrapper
}
