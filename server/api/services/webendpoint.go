package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// WebEndpointResolverFactoryFunc constructs a WebEndpointResolver from the core store and
// cache. Enterprise packages register a factory via RegisterWebEndpointResolver in their
// init() functions; it runs during server setup.
type WebEndpointResolverFactoryFunc func(ctx context.Context, store store.Store, cache cache.Cache) (WebEndpointResolver, error)

var webEndpointResolverFactory WebEndpointResolverFactoryFunc

// RegisterWebEndpointResolver registers the factory that creates the web endpoint
// resolver. It must be called before the server's Setup() runs.
func RegisterWebEndpointResolver(f WebEndpointResolverFactoryFunc) {
	webEndpointResolverFactory = f
}

// WebEndpointResolverFactory returns the registered factory, or nil in Community Edition
// builds.
func WebEndpointResolverFactory() WebEndpointResolverFactoryFunc {
	return webEndpointResolverFactory
}

// WebEndpointResolver resolves a public web endpoint address to the device and backend it
// points at.
type WebEndpointResolver interface {
	// LookupWebEndpoint returns the endpoint registered under address, or an error
	// when it does not exist or has expired.
	LookupWebEndpoint(ctx context.Context, address string) (*models.WebEndpoint, error)
}

type WebEndpointService interface {
	// LookupWebEndpoint resolves a public web endpoint address, returning
	// ErrWebEndpointNotAvailable when the feature is not enabled in this edition.
	LookupWebEndpoint(ctx context.Context, address string) (*models.WebEndpoint, error)
}

func (s *service) LookupWebEndpoint(ctx context.Context, address string) (*models.WebEndpoint, error) {
	if s.webEndpoints == nil {
		return nil, ErrWebEndpointNotAvailable
	}

	return s.webEndpoints.LookupWebEndpoint(ctx, address)
}
