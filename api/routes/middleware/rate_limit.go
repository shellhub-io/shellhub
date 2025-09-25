package middleware

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type Expirable[D any] struct {
	data     *D
	duration time.Duration
	lastSeen *time.Time
}

// IsExpired checks if the structure is expired.
func (n *Expirable[D]) IsExpired() bool {
	if n.lastSeen == nil {
		return false
	}

	if n.data == nil {
		return false
	}

	return time.Since(*n.lastSeen) > n.duration
}

func (n *Expirable[D]) Get() *D {
	if n.IsExpired() {
		return nil
	}

	return n.data
}

const (
	// DefaultNamespaceCacheDuration is the default duration for which a namespace is cached.
	DefaultNamespaceCacheDuration = 30 * time.Minute
	// DefaultNamespaceRateLimit defines the rate at which tokens are replenished into the bucket for the rate limiter.
	DefaultNamespaceRateLimit = 1000
	// DefaultNamespaceBurst defines the maximun size of the bucket for the rate limiter.
	DefaultNamespaceBurst = 1000
)

func NewNamespaceCached(namespace *models.Namespace, duration time.Duration) *Expirable[models.Namespace] {
	if duration <= 0 {
		duration = DefaultNamespaceCacheDuration
	}

	t := time.Now()

	return &Expirable[models.Namespace]{
		data:     namespace,
		duration: duration,
		lastSeen: &t,
	}
}

type NamespaceRateLimitOptions struct {
	// cacheDuration specifies how long the namespace cache should be valid.
	cacheDuration time.Duration
	// rate specify how many requests per second are allowed.
	rate int
	// burst specifies the maximum burst size for the rate limiter.
	burst int
}

func DefaultNamespaceRateLimitOptions() *NamespaceRateLimitOptions {
	return &NamespaceRateLimitOptions{
		cacheDuration: DefaultNamespaceCacheDuration,
	}
}

type NamespaceRateLimitOption func(*NamespaceRateLimitOptions) *NamespaceRateLimitOptions

// NamespaceRateLimitWithCacheDuration sets the duration for which the namespace cache is valid.
func NamespaceRateLimitWithCacheDuration(duration time.Duration) NamespaceRateLimitOption {
	return func(options *NamespaceRateLimitOptions) *NamespaceRateLimitOptions {
		options.cacheDuration = duration

		return options
	}
}

// NamespaceRateLimitWithRate sets the rate limit of requests per second for the rate limiter.
func NamespaceRateLimitWithRate(rate int) NamespaceRateLimitOption {
	return func(options *NamespaceRateLimitOptions) *NamespaceRateLimitOptions {
		options.rate = rate

		return options
	}
}

// NamespaceRateLimitWithBurst sets the burst size for the rate limiter.
func NamespaceRateLimitWithBurst(burst int) NamespaceRateLimitOption {
	return func(options *NamespaceRateLimitOptions) *NamespaceRateLimitOptions {
		options.burst = burst

		return options
	}
}

type NamespaceRateLimit struct {
	config *NamespaceRateLimitOptions

	mutex   *sync.Mutex
	mutexts map[string]*sync.Mutex

	services services.Service

	cached   map[string]*Expirable[models.Namespace]
	limiters map[string]*rate.Limiter
}

func NewNamespaceRateLimit(svc any, options ...NamespaceRateLimitOption) *NamespaceRateLimit {
	s, _ := svc.(services.Service)

	config := &NamespaceRateLimitOptions{
		cacheDuration: DefaultNamespaceCacheDuration,
		rate:          DefaultNamespaceRateLimit,
		burst:         DefaultNamespaceBurst,
	}

	for _, option := range options {
		config = option(config)
	}

	return &NamespaceRateLimit{
		config: config,

		mutex:   new(sync.Mutex),
		mutexts: make(map[string]*sync.Mutex),

		services: s,

		cached:   make(map[string]*Expirable[models.Namespace]),
		limiters: make(map[string]*rate.Limiter),
	}
}

// getTenantMutex gets or creates a mutex for the given tenant in a thread-safe way
func (l *NamespaceRateLimit) getTenantMutex(tenant string) *sync.Mutex {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	mutex, exists := l.mutexts[tenant]
	if !exists {
		mutex = &sync.Mutex{}
		l.mutexts[tenant] = mutex
	}

	return mutex
}

func (l *NamespaceRateLimit) Allow(tenant string) (bool, error) {
	if l.services == nil {
		log.Warn("rate limiter service is not configured - allowing request")

		return true, nil
	}

	if strings.TrimSpace(tenant) == "" {
		log.Error("tenant ID cannot be empty")

		return false, fmt.Errorf("tenant ID cannot be empty")
	}

	mu := l.getTenantMutex(tenant)

	mu.Lock()
	defer mu.Unlock()

	cached, exists := l.cached[tenant]

	needsRefresh := !exists || (cached != nil && cached.IsExpired())
	if needsRefresh {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		namespace, err := l.services.GetNamespace(ctx, tenant)
		if err != nil {
			log.WithFields(log.Fields{
				"tenant": tenant,
			}).WithError(err).Error("failed to fetch namespace for rate limiter")

			return false, fmt.Errorf("failed to fetch namespace: %w", err)
		}

		if namespace == nil {
			return false, fmt.Errorf("namespace not found for tenant: %s", tenant)
		}

		// TODO: We'll increase or decrease the rate dynamically based on the namespace characteristics in the future.
		l.cached[tenant] = NewNamespaceCached(namespace, DefaultNamespaceCacheDuration)
		l.limiters[tenant] = rate.NewLimiter(rate.Limit(l.config.rate), l.config.burst)

		log.WithFields(log.Fields{
			"tenant":    tenant,
			"namespace": namespace.Name,
		}).Debug("namespace cache refreshed for rate limiter")
	}

	limiter, exists := l.limiters[tenant]
	if !exists {
		log.WithField("tenant", tenant).Error("rate limiter visitor not found after cache refresh")

		return false, fmt.Errorf("rate limiter not configured for tenant: %s", tenant)
	}

	allowed := limiter.Allow()

	log.WithFields(log.Fields{
		"tenant":  tenant,
		"allowed": allowed,
		"tokens":  limiter.Tokens(),
	}).Debug("rate limiter check completed")

	return allowed, nil
}

// CleanupExpiredEntries removes expired cache entries (call this periodically)
func (l *NamespaceRateLimit) CleanupExpiredEntries() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for tenant, cached := range l.cached {
		if cached != nil && cached.IsExpired() {
			delete(l.cached, tenant)
			delete(l.limiters, tenant)
			delete(l.mutexts, tenant)

			log.WithField("tenant", tenant).Debug("cleaned up expired rate limiter cache entry")
		}
	}
}

// SkipperNamespace is a function that checks if the context contains a valid tenant ID.
var SkipperNamespace = func(context echo.Context) bool {
	c, ok := context.(*gateway.Context)
	if !ok {
		log.Error("context is not of type gateway.Context for rate limiting")

		return true
	}

	tenant, ok := c.GetTennat()
	if !ok || tenant == "" {
		log.Error("tenant ID cannot be empty in request context for rate limiting")

		return true
	}

	return false
}

// NewNamespaceRateLimitMiddleware creates a middleware that limits the rate of requests based on the tenant ID
// extracted from the request context.
func NewNamespaceRateLimitMiddleware(service any, options ...NamespaceRateLimitOption) echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: SkipperNamespace,
		IdentifierExtractor: func(context echo.Context) (string, error) {
			c, ok := context.(*gateway.Context)
			if !ok {
				return "", fmt.Errorf("context is not of type gateway.Context")
			}

			tenant, ok := c.GetTennat()
			if !ok || tenant == "" {
				log.Error("tenant ID cannot be empty in request context for rate limiting")

				return "", fmt.Errorf("tenant ID cannot be empty in request context for rate limiting")
			}

			return tenant, nil
		},
		Store: NewNamespaceRateLimit(service, options...),
		ErrorHandler: func(c echo.Context, err error) error {
			return &echo.HTTPError{
				Code:     middleware.ErrRateLimitExceeded.Code,
				Message:  middleware.ErrRateLimitExceeded.Message,
				Internal: err,
			}
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return &echo.HTTPError{
				Code:     middleware.ErrRateLimitExceeded.Code,
				Message:  middleware.ErrRateLimitExceeded.Message,
				Internal: err,
			}
		},
	})
}
