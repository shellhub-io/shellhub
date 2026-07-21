package services

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

type AccessPolicyService interface {
	// Authorize decides whether the user may reach the device as the given login,
	// connecting from sourceIP, under the namespace's Access Policies. It is
	// default-deny and fail-closed: access is granted iff at least one policy
	// grants it, and any store failure denies. It is the authorization model for
	// the identity-based SSH access mode; the gateway calls it at the
	// ephemeral-key mint point.
	Authorize(ctx context.Context, tenantID, userID string, device *models.Device, login, sourceIP string) (*models.Decision, error)

	// ListAccessPolicies returns every access policy in the namespace.
	ListAccessPolicies(ctx context.Context, tenantID string) ([]models.AccessPolicy, error)

	// NamespaceHasAccessPolicies reports whether the namespace has any access
	// policy. The gateway uses it to refuse an identity-mode login before minting
	// an approval when no policy could ever grant access (default-deny), so the
	// user is not asked to approve a login that is certain to be denied.
	NamespaceHasAccessPolicies(ctx context.Context, tenantID string) (bool, error)

	// GetAccessPolicy returns a single access policy by id within the namespace.
	GetAccessPolicy(ctx context.Context, req *requests.AccessPolicyGet) (*models.AccessPolicy, error)

	// CreateAccessPolicy creates a new access policy in the namespace.
	CreateAccessPolicy(ctx context.Context, req *requests.AccessPolicyCreate) (*models.AccessPolicy, error)

	// UpdateAccessPolicy updates an existing access policy in the namespace.
	UpdateAccessPolicy(ctx context.Context, req *requests.AccessPolicyUpdate) (*models.AccessPolicy, error)

	// DeleteAccessPolicy removes an access policy from the namespace.
	DeleteAccessPolicy(ctx context.Context, req *requests.AccessPolicyDelete) error
}

func (s *service) Authorize(ctx context.Context, tenantID, userID string, device *models.Device, login, sourceIP string) (*models.Decision, error) {
	// Resolve the device from the store so the filter matches against the
	// authoritative name and tag ids, not whatever the caller supplied.
	dev, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, device.UID)
	if err != nil {
		return nil, NewErrDeviceNotFound(models.UID(device.UID), err)
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(tenantID, err)
	}

	member, ok := namespace.FindMember(userID)
	if !ok {
		return &models.Decision{Allowed: false, Reason: "user is not a member of the namespace"}, nil
	}

	policies, _, err := s.store.AccessPolicyList(ctx, s.store.Options().InNamespace(tenantID))
	if err != nil {
		return nil, err
	}

	// Deny wins: a matching deny blocks access before any allow is considered,
	// however specific the allow. It is fail-closed — a deny whose filter cannot
	// be evaluated denies rather than silently opening access.
	for _, policy := range policies {
		if policy.Effect != models.PolicyEffectDeny {
			continue
		}

		matched, err := policyApplies(policy, dev, userID, member.Role, login, sourceIP)
		if err != nil {
			log.WithError(err).WithField("access_policy", policy.ID).
				Warn("deny access policy failed to evaluate; denying")

			return &models.Decision{Allowed: false, Reason: fmt.Sprintf("denied: policy %q could not be evaluated", policy.Name)}, nil
		}

		if matched {
			return &models.Decision{Allowed: false, Reason: fmt.Sprintf("denied by policy %q", policy.Name)}, nil
		}
	}

	// Allow: the first allow policy that matches grants access. A malformed
	// hostname regexp is treated as a non-match so one broken policy never blocks
	// a valid one, and the default stays deny.
	for _, policy := range policies {
		if policy.Effect == models.PolicyEffectDeny {
			continue
		}

		matched, err := policyApplies(policy, dev, userID, member.Role, login, sourceIP)
		if err != nil {
			log.WithError(err).WithField("access_policy", policy.ID).
				Warn("access policy failed to evaluate; treating as non-match")

			continue
		}

		if matched {
			return &models.Decision{Allowed: true, RequireStepUp: policy.RequireStepUp}, nil
		}
	}

	return &models.Decision{Allowed: false, Reason: fmt.Sprintf("no policy grants %q on this device", login)}, nil
}

// policyApplies reports whether the policy's subject, device filter, login, and
// source IP all match the request. The bool is only meaningful when err is nil;
// a non-nil error means a matcher could not be evaluated (a broken filter regexp,
// or a malformed source CIDR / client IP), and the caller decides how to treat it
// (deny fails closed, allow treats it as a non-match).
func policyApplies(policy models.AccessPolicy, dev *models.Device, userID string, role authorizer.Role, login, sourceIP string) (bool, error) {
	if !subjectMatches(policy.Subject, userID, role) {
		return false, nil
	}

	matched, err := policy.Filter.Matches(dev)
	if err != nil {
		return false, err
	}

	if !matched || !loginMatches(policy.Logins, login) {
		return false, nil
	}

	return sourceIPMatches(policy.SourceIP, sourceIP)
}

// normalizeSourceIPs canonicalizes source entries to CIDR form so a bare IP the
// user typed (e.g. "203.0.113.5") is stored and matched as a host route
// ("203.0.113.5/32", or /128 for IPv6). Entries already in CIDR form pass
// through; anything unparseable is left as-is (the handler validates first).
func normalizeSourceIPs(entries []string) []string {
	out := make([]string, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if !strings.Contains(entry, "/") {
			if addr, err := netip.ParseAddr(entry); err == nil {
				entry = netip.PrefixFrom(addr.Unmap(), addr.Unmap().BitLen()).String()
			}
		}

		out = append(out, entry)
	}

	return out
}

// sourceIPMatches reports whether the client IP falls within any of the policy's
// source CIDRs (OR). An empty list matches any IP. A malformed CIDR or an
// unparseable client IP returns an error so the caller can fail closed on deny.
func sourceIPMatches(cidrs []string, clientIP string) (bool, error) {
	if len(cidrs) == 0 {
		return true, nil
	}

	addr, err := netip.ParseAddr(clientIP)
	if err != nil {
		return false, fmt.Errorf("invalid client ip %q: %w", clientIP, err)
	}

	addr = addr.Unmap()

	for _, cidr := range cidrs {
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			return false, fmt.Errorf("invalid source cidr %q: %w", cidr, err)
		}

		if prefix.Contains(addr) {
			return true, nil
		}
	}

	return false, nil
}

func (s *service) NamespaceHasAccessPolicies(ctx context.Context, tenantID string) (bool, error) {
	_, count, err := s.store.AccessPolicyList(ctx, s.store.Options().InNamespace(tenantID))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// subjectMatches reports whether the policy subject applies to the given user
// and role.
func subjectMatches(subject models.PolicySubject, userID string, role authorizer.Role) bool {
	switch subject.Type {
	case models.PolicySubjectAllMembers:
		return true
	case models.PolicySubjectRole:
		return subject.Value == role.String()
	case models.PolicySubjectUser:
		return subject.Value == userID
	default:
		return false
	}
}

// loginMatches reports whether the login is covered by the policy's login list:
// an exact match, or a wildcard entry.
func loginMatches(logins []string, login string) bool {
	for _, l := range logins {
		if l == "*" || l == login {
			return true
		}
	}

	return false
}

// defaultEffect resolves a request's effect, defaulting an omitted value to
// allow so clients need not send it for the common grant case.
func defaultEffect(effect string) models.PolicyEffect {
	if effect == "" {
		return models.PolicyEffectAllow
	}

	return models.PolicyEffect(effect)
}

func (s *service) ListAccessPolicies(ctx context.Context, tenantID string) ([]models.AccessPolicy, error) {
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID); err != nil {
		return nil, NewErrNamespaceNotFound(tenantID, err)
	}

	policies, _, err := s.store.AccessPolicyList(ctx, s.store.Options().InNamespace(tenantID))
	if err != nil {
		return nil, err
	}

	return policies, nil
}

func (s *service) GetAccessPolicy(ctx context.Context, req *requests.AccessPolicyGet) (*models.AccessPolicy, error) {
	policy, err := s.store.AccessPolicyResolve(ctx, store.AccessPolicyIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		return nil, NewErrAccessPolicyNotFound(req.ID, err)
	}

	return policy, nil
}

func (s *service) CreateAccessPolicy(ctx context.Context, req *requests.AccessPolicyCreate) (*models.AccessPolicy, error) {
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	filter, err := s.resolveAccessPolicyFilter(ctx, req.TenantID, req.Filter)
	if err != nil {
		return nil, err
	}

	policy := &models.AccessPolicy{
		TenantID:      req.TenantID,
		Name:          req.Name,
		Subject:       models.PolicySubject{Type: models.PolicySubjectType(req.Subject.Type), Value: req.Subject.Value},
		Filter:        filter,
		Logins:        req.Logins,
		SourceIP:      normalizeSourceIPs(req.SourceIP),
		Effect:        defaultEffect(req.Effect),
		RequireStepUp: req.RequireStepUp,
	}

	id, err := s.store.AccessPolicyCreate(ctx, policy)
	if err != nil {
		return nil, err
	}

	return s.store.AccessPolicyResolve(ctx, store.AccessPolicyIDResolver, id, s.store.Options().InNamespace(req.TenantID))
}

func (s *service) UpdateAccessPolicy(ctx context.Context, req *requests.AccessPolicyUpdate) (*models.AccessPolicy, error) {
	if _, err := s.store.AccessPolicyResolve(ctx, store.AccessPolicyIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID)); err != nil {
		return nil, NewErrAccessPolicyNotFound(req.ID, err)
	}

	filter, err := s.resolveAccessPolicyFilter(ctx, req.TenantID, req.Filter)
	if err != nil {
		return nil, err
	}

	policy := &models.AccessPolicy{
		ID:            req.ID,
		TenantID:      req.TenantID,
		Name:          req.Name,
		Subject:       models.PolicySubject{Type: models.PolicySubjectType(req.Subject.Type), Value: req.Subject.Value},
		Filter:        filter,
		Logins:        req.Logins,
		SourceIP:      normalizeSourceIPs(req.SourceIP),
		Effect:        defaultEffect(req.Effect),
		RequireStepUp: req.RequireStepUp,
	}

	if err := s.store.AccessPolicyUpdate(ctx, policy); err != nil {
		return nil, err
	}

	return s.store.AccessPolicyResolve(ctx, store.AccessPolicyIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID))
}

func (s *service) DeleteAccessPolicy(ctx context.Context, req *requests.AccessPolicyDelete) error {
	if _, err := s.store.AccessPolicyResolve(ctx, store.AccessPolicyIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID)); err != nil {
		return NewErrAccessPolicyNotFound(req.ID, err)
	}

	return s.store.AccessPolicyDelete(ctx, &models.AccessPolicy{ID: req.ID, TenantID: req.TenantID})
}

// resolveAccessPolicyFilter translates the request's device selector into a
// stored filter, resolving tag names to their ids (mirroring the public-key
// create path).
func (s *service) resolveAccessPolicyFilter(ctx context.Context, tenantID string, reqFilter requests.AccessPolicyFilter) (models.PublicKeyFilter, error) {
	filter := models.PublicKeyFilter{Hostname: reqFilter.Hostname}

	if len(reqFilter.Tags) == 0 {
		return filter, nil
	}

	tags, _, err := s.store.TagList(ctx, s.store.Options().InNamespace(tenantID))
	if err != nil {
		return filter, NewErrTagEmpty(tenantID, err)
	}

	tagIDs := make([]string, 0, len(reqFilter.Tags))
	for _, tagName := range reqFilter.Tags {
		found := false
		for _, tag := range tags {
			if tagName == tag.Name {
				tagIDs = append(tagIDs, tag.ID)
				found = true

				break
			}
		}

		if !found {
			return filter, NewErrTagNotFound(tagName, nil)
		}
	}

	filter.TagIDs = tagIDs

	return filter, nil
}

// seedAccessPolicy creates the permissive starter policy when a namespace
// switches to identity access mode with no policies yet, so default-deny does
// not silently lock everyone out. The seed grants every member every login on
// every device; it is visible in the console as the thing to tighten.
func (s *service) seedAccessPolicy(ctx context.Context, tenantID string) error {
	policies, count, err := s.store.AccessPolicyList(ctx, s.store.Options().InNamespace(tenantID))
	if err != nil {
		return err
	}

	if count > 0 || len(policies) > 0 {
		return nil
	}

	seed := &models.AccessPolicy{
		TenantID: tenantID,
		Name:     "Default access",
		Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
		Filter:   models.PublicKeyFilter{},
		Logins:   []string{"*"},
		Effect:   models.PolicyEffectAllow,
	}

	if _, err := s.store.AccessPolicyCreate(ctx, seed); err != nil {
		return err
	}

	return nil
}
