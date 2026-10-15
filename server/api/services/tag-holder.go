package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// TagHolderFactoryFunc constructs a TagHolder from the core store and cache. Enterprise packages
// register a factory via RegisterTagHolder in their init() functions; it runs during server setup.
type TagHolderFactoryFunc func(ctx context.Context, store store.Store, cache cache.Cache) (TagHolder, error)

var tagHolderFactory TagHolderFactoryFunc

// RegisterTagHolder registers the factory that creates the tag holder. It must be called before
// the server's Setup() runs.
func RegisterTagHolder(f TagHolderFactoryFunc) {
	tagHolderFactory = f
}

// TagHolderFactory returns the registered factory, or nil in Community Edition builds.
func TagHolderFactory() TagHolderFactoryFunc {
	return tagHolderFactory
}

// TagHolder names what holds a tag in a device filter the core store does not know of, such as
// the enterprise firewall rules, so that deleting the tag is refused as it is for a public key or
// an access policy holding it.
type TagHolder interface {
	// TagHolders returns one name per thing in sc whose device filter holds the tag tagID, each
	// in the form the conflict reports it, or an error when they cannot be listed. DeleteTag calls
	// it inside its transaction, so a store read through ctx joins that transaction, and an error
	// fails the deletion.
	TagHolders(ctx context.Context, sc scope.Scope, tagID string) ([]string, error)
}
