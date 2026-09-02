package services

import (
	"crypto/rsa"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// APIService is the concrete implementation of [Service]. It is a thin wrapper so that the
// enterprise build can embed it and override individual methods.
type APIService struct {
	*service
}

var _ Service = (*APIService)(nil)

type service struct {
	store             store.Store
	privKey           *rsa.PrivateKey
	pubKey            *rsa.PublicKey
	cache             cache.Cache
	locator           geoip.Locator
	validator         *validator.Validator
	billing           BillingProvider
	licenseEvaluator  LicenseEvaluator
	firewallEvaluator FirewallEvaluator
	recordingPruner   SessionRecordingPruner
}

// Service is the whole service layer, composed from the per-resource interfaces above. A
// handler takes this and asserts down to what it needs.
type Service interface {
	TagsService
	DeviceService
	DeviceLoginCodeService
	DevicePairingService
	SSHApprovalService
	AccessPolicyService
	SSHIdentityService
	ServiceAccountService
	WebReauthService
	UserService
	SSHKeysService
	SessionService
	NamespaceService
	MemberService
	InvitationService
	AuthService
	StatsService
	SetupService
	SystemService
	APIKeyService
	InstanceAPIKeyService
	InstallKeyService
	FirewallService
	LicenseService
	BillingService

	// Store returns the underlying store instance.
	//
	// IMPORTANT: Extensions should use Store() ONLY for:
	// 1. Creating their own service layers that need database access
	// 2. Querying data not exposed by core Service methods
	//
	// Extensions MUST NOT:
	// 1. Modify core data (devices, users, namespaces) directly via Store()
	// 2. Bypass core business logic or validation rules
	// 3. Create transactions that span both core and extension operations
	//
	// Violating these rules may cause data inconsistency or break core functionality.
	Store() store.Store
}

// Option supplies a capability the community build has no implementation for. Each is
// optional, and the enterprise binary is what passes them.
type Option func(service *APIService)

// WithLocator resolves IP addresses to locations, for session geolocation.
func WithLocator(locator geoip.Locator) Option {
	return func(service *APIService) {
		service.locator = locator
	}
}

// WithBilling attaches the billing system, without which billing is never consulted.
func WithBilling(billing BillingProvider) Option {
	return func(service *APIService) {
		service.billing = billing
	}
}

// WithLicenseEvaluator attaches licence enforcement, without which nothing is licence-gated.
func WithLicenseEvaluator(le LicenseEvaluator) Option {
	return func(service *APIService) {
		service.licenseEvaluator = le
	}
}

// WithFirewallEvaluator attaches firewall evaluation, without which no rule is applied.
func WithFirewallEvaluator(fe FirewallEvaluator) Option {
	return func(service *APIService) {
		service.firewallEvaluator = fe
	}
}

// WithSessionRecordingPruner attaches the pruner that enforces recording retention.
func WithSessionRecordingPruner(rp SessionRecordingPruner) Option {
	return func(service *APIService) {
		service.recordingPruner = rp
	}
}

// NewService builds the service layer over store, signing tokens with the given key pair. A
// nil pair is loaded from the configured files, which is how the server normally starts.
func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cache cache.Cache, options ...Option) *APIService {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = LoadKeys()
		if err != nil {
			panic(err)
		}
	}

	service := &APIService{
		service: &service{
			store:             store,
			privKey:           privKey,
			pubKey:            pubKey,
			cache:             cache,
			locator:           geoip.NewNullGeoLite(),
			validator:         validator.New(),
			billing:           nil, // injected via WithBilling option
			licenseEvaluator:  nil, // injected via WithLicenseEvaluator option
			firewallEvaluator: nil, // injected via WithFirewallEvaluator option
			recordingPruner:   nil, // injected via WithSessionRecordingPruner option
		},
	}

	for _, option := range options {
		option(service)
	}

	return service
}

// Store returns the underlying store instance.
// This allows route extensions (enterprise/cloud) to access the same database connection.
func (s *APIService) Store() store.Store {
	return s.store
}
