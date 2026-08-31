package services

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

type AccessPolicyService interface {
	// Authorize decides whether the user may reach the device as the given login,
	// connecting from sourceIP, under the namespace's Access Policies. It is
	// default-deny and fail-closed: access is granted iff at least one policy
	// grants it, and any store failure denies. It is the authorization model for
	// the identity-based SSH access mode; the gateway calls it at the
	// ephemeral-key mint point.
	Authorize(ctx context.Context, tenantID, userID, deviceUID, login, sourceIP string) (*models.Decision, error)

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

func (s *service) Authorize(ctx context.Context, tenantID, userID, deviceUID, login, sourceIP string) (*models.Decision, error) {
	sc, err := BoundTo(tenantID)
	if err != nil {
		return nil, err
	}

	dev, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, deviceUID)
	if err != nil {
		return nil, NewErrDeviceNotFound(models.UID(deviceUID), err)
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(tenantID, err)
	}

	member, ok := namespace.FindMember(userID)
	if !ok {
		return &models.Decision{Allowed: false, Reason: models.ReasonNotAMember}, nil
	}

	policies, _, err := s.store.AccessPolicyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	for _, policy := range policies {
		if policy.Action != models.PolicyActionDeny {
			continue
		}

		matched, err := policyApplies(policy, dev, userID, member.Role, member.Type, login, sourceIP)
		if err != nil {
			log.WithError(err).WithField("access_policy", policy.ID).
				Warn("deny access policy failed to evaluate; denying")

			return &models.Decision{Allowed: false, Reason: models.ReasonPolicyUnevaluable, PolicyName: policy.Name}, nil
		}

		if matched {
			return &models.Decision{Allowed: false, Reason: models.ReasonDeniedByPolicy, PolicyName: policy.Name}, nil
		}
	}

	allowed := false
	requireReauth := false

	var reauthPeriod *int

	for _, policy := range policies {
		if policy.Action == models.PolicyActionDeny {
			continue
		}

		matched, err := policyApplies(policy, dev, userID, member.Role, member.Type, login, sourceIP)
		if err != nil {
			log.WithError(err).WithField("access_policy", policy.ID).
				Warn("access policy failed to evaluate; treating as non-match")

			continue
		}

		if !matched {
			continue
		}

		allowed = true

		if policy.RequireReauth {
			if requireReauth {
				reauthPeriod = stricterReauthPeriod(reauthPeriod, policy.ReauthPeriod)
			} else {
				reauthPeriod = policy.ReauthPeriod
			}

			requireReauth = true
		}
	}

	if !allowed {
		return &models.Decision{Allowed: false, Reason: models.ReasonNoGrant, Login: login}, nil
	}

	if member.Type == models.UserTypeService {
		requireReauth = false
		reauthPeriod = nil
	}

	return &models.Decision{Allowed: true, RequireReauth: requireReauth, ReauthPeriod: reauthPeriod}, nil
}

// policyApplies reports whether the policy's subject, device filter, login, and
// source IP all match the request. The bool is only meaningful when err is nil;
// a non-nil error means a matcher could not be evaluated (a broken filter regexp,
// or a malformed source CIDR / client IP), and the caller decides how to treat it
// (deny fails closed, allow treats it as a non-match).
func policyApplies(policy models.AccessPolicy, dev *models.Device, userID string, role authorizer.Role, userType models.UserType, login, sourceIP string) (bool, error) {
	if !subjectMatches(policy.Subject, userID, role, userType) {
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
	sc, err := BoundTo(tenantID)
	if err != nil {
		return false, err
	}

	_, count, err := s.store.AccessPolicyList(ctx, sc)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// subjectMatches reports whether the policy subject applies to the given principal.
func subjectMatches(subject models.PolicySubject, userID string, role authorizer.Role, userType models.UserType) bool {
	switch subject.Type {
	case models.PolicySubjectAllMembers:
		return userType != models.UserTypeService
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

// defaultAction resolves a request's action, defaulting an omitted value to
// allow so clients need not send it for the common grant case.
func defaultAction(action string) models.PolicyAction {
	if action == "" {
		return models.PolicyActionAllow
	}

	return models.PolicyAction(action)
}

// normalizeReauthPeriod collapses a zero period to nil so "always" has a single
// stored representation: both an omitted period and an explicit 0 mean re-auth
// every session.
func normalizeReauthPeriod(period *int) *int {
	if period != nil && *period == 0 {
		return nil
	}

	return period
}

// stricterReauthPeriod returns the more demanding of two re-auth freshness windows.
// A nil period means "every session" — the strictest — and always wins; between two
// concrete windows the shorter one wins, since it forces re-auth more often.
func stricterReauthPeriod(a, b *int) *int {
	if a == nil || b == nil {
		return nil
	}

	if *b < *a {
		return b
	}

	return a
}

func (s *service) ListAccessPolicies(ctx context.Context, tenantID string) ([]models.AccessPolicy, error) {
	sc, err := BoundTo(tenantID)
	if err != nil {
		return nil, err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID); err != nil {
		return nil, NewErrNamespaceNotFound(tenantID, err)
	}

	policies, _, err := s.store.AccessPolicyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	return policies, nil
}

func (s *service) GetAccessPolicy(ctx context.Context, req *requests.AccessPolicyGet) (*models.AccessPolicy, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	policy, err := s.store.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, req.ID)
	if err != nil {
		return nil, NewErrAccessPolicyNotFound(req.ID, err)
	}

	return policy, nil
}

func (s *service) CreateAccessPolicy(ctx context.Context, req *requests.AccessPolicyCreate) (*models.AccessPolicy, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	filter, err := s.resolveAccessPolicyFilter(ctx, sc, req.Filter)
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
		Action:        defaultAction(req.Action),
		RequireReauth: req.RequireReauth,
		ReauthPeriod:  normalizeReauthPeriod(req.ReauthPeriod),
	}

	id, err := s.store.AccessPolicyCreate(ctx, policy)
	if err != nil {
		return nil, err
	}

	return s.store.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, id)
}

func (s *service) UpdateAccessPolicy(ctx context.Context, req *requests.AccessPolicyUpdate) (*models.AccessPolicy, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	if _, err := s.store.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, req.ID); err != nil {
		return nil, NewErrAccessPolicyNotFound(req.ID, err)
	}

	filter, err := s.resolveAccessPolicyFilter(ctx, sc, req.Filter)
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
		Action:        defaultAction(req.Action),
		RequireReauth: req.RequireReauth,
		ReauthPeriod:  normalizeReauthPeriod(req.ReauthPeriod),
	}

	if err := s.store.AccessPolicyUpdate(ctx, policy); err != nil {
		return nil, err
	}

	return s.store.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, req.ID)
}

func (s *service) DeleteAccessPolicy(ctx context.Context, req *requests.AccessPolicyDelete) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	if _, err := s.store.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, req.ID); err != nil {
		return NewErrAccessPolicyNotFound(req.ID, err)
	}

	return s.store.AccessPolicyDelete(ctx, &models.AccessPolicy{ID: req.ID, TenantID: req.TenantID})
}

// resolveAccessPolicyFilter translates the request's device selector into a
// stored filter, resolving tag names to their ids (mirroring the public-key
// create path).
func (s *service) resolveAccessPolicyFilter(ctx context.Context, sc scope.Scope, reqFilter requests.AccessPolicyFilter) (models.PublicKeyFilter, error) {
	filter := models.PublicKeyFilter{Hostname: reqFilter.Hostname}

	if len(reqFilter.Tags) == 0 {
		return filter, nil
	}

	tags, _, err := s.store.TagList(ctx, sc)
	if err != nil {
		return filter, NewErrTagEmpty(sc.TenantID(), err)
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

// seedAccessPolicy creates the owner starter policy (see NewOwnerAccessPolicy)
// when a namespace switches to identity access mode with no policies yet, so
// default-deny does not lock the owner out.
func (s *service) seedAccessPolicy(ctx context.Context, tenantID, ownerID string) error {
	sc, err := BoundTo(tenantID)
	if err != nil {
		return err
	}

	_, count, err := s.store.AccessPolicyList(ctx, sc)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	if _, err := s.store.AccessPolicyCreate(ctx, models.NewOwnerAccessPolicy(tenantID, ownerID)); err != nil {
		return err
	}

	return nil
}
