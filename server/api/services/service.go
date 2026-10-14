package services

import (
	"crypto/hkdf"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"net/netip"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
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
	issuer            string
	challengeKey      []byte
	clock             clock.Clock

	requireVerifiedDeviceKey bool

	edition      envs.Edition
	version      string
	sshPort      string
	development  bool
	webhookCIDRs []netip.Prefix
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
	ProvisioningKeyService
	FirewallService
	LicenseService
	BillingService
}

// Option configures the service at construction. Most supply a capability the community build
// has no implementation for, which the enterprise binary passes; the rest carry a value the
// server knows and the service layer does not, such as the token issuer.
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

// WithClock makes the service read the time from c. Without it the service reads the wall clock.
func WithClock(c clock.Clock) Option {
	return func(service *APIService) {
		service.clock = c
	}
}

// WithIssuer sets the identifier written to the iss claim of every token the service signs.
// Without it tokens are signed with an empty issuer.
func WithIssuer(issuer string) Option {
	return func(service *APIService) {
		service.issuer = issuer
	}
}

// WithVerifiedDeviceKeyRequired refuses every device authentication that does not prove possession
// of the device's key, which agents released before device key verification cannot send. Without it
// such an authentication is still accepted for a device that has never verified its key.
func WithVerifiedDeviceKeyRequired() Option {
	return func(service *APIService) {
		service.requireVerifiedDeviceKey = true
	}
}

// WithEdition sets the edition the service gates its features on. Without it the service runs as
// Community.
func WithEdition(edition envs.Edition) Option {
	return func(service *APIService) {
		service.edition = edition
	}
}

// WithSystemInfo sets the version and the SSH port the service reports in the system info.
func WithSystemInfo(version, sshPort string) Option {
	return func(service *APIService) {
		service.version = version
		service.sshPort = sshPort
	}
}

// WithDevelopmentMode makes setup give the dev namespace its fixed tenant ID. Never pass it to
// relax a security decision; it exists for developer conveniences only.
func WithDevelopmentMode(development bool) Option {
	return func(service *APIService) {
		service.development = development
	}
}

// WithEnrollmentWebhookCIDRs lets a provisioning key webhook reach the given prefixes, which the
// SSRF guard refuses otherwise. See ParseEnrollmentWebhookCIDRs.
func WithEnrollmentWebhookCIDRs(prefixes []netip.Prefix) Option {
	return func(service *APIService) {
		service.webhookCIDRs = prefixes
	}
}

// NewService builds the service layer over store, signing tokens with the given key pair, which
// the caller loads with LoadKeys. It panics when privKey is nil.
func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey, cache cache.Cache, options ...Option) *APIService {
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
			edition:           envs.Community,
			clock:             clock.New(),
		},
	}

	for _, option := range options {
		option(service)
	}

	service.challengeKey = deviceAuthChallengeKey(privKey)

	return service
}

func deviceAuthChallengeKey(privKey *rsa.PrivateKey) []byte {
	key, err := hkdf.Key(sha256.New, x509.MarshalPKCS1PrivateKey(privKey), nil, "shellhub-device-auth-challenge-v1", sha256.Size)
	if err != nil {
		panic(err)
	}

	return key
}
