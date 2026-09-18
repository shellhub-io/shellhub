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

// AccessPolicyService answers whether a principal may connect to a device. The namespace's
// Access Policies decide which devices and logins, and the member's role decides whether they
// connect at all.
type AccessPolicyService interface {
	// Authorize decides whether the user may reach the device as the given login,
	// connecting from sourceIP, under the namespace's Access Policies. It is
	// default-deny and fail-closed: access is granted iff at least one policy
	// grants it, and any store failure denies. It is the authorization model for
	// the identity-based SSH access mode; the gateway calls it at the
	// ephemeral-key mint point.
	//
	// A member whose role lacks [authorizer.DeviceConnect] is refused with
	// [models.ReasonRoleCannotConnect] before any policy is read, since no policy could
	// grant what the role withholds. Service accounts are exempt by user type:
	// [authorizer.RoleService] holds no permissions by design, and their access comes
	// entirely from the policies.
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

	if !member.IsService() && !member.Role.HasPermission(authorizer.DeviceConnect) {
		return &models.Decision{Allowed: false, Reason: models.ReasonRoleCannotConnect}, nil
	}

	principal := models.Principal{Kind: principalKindOfMember(*member), ID: userID}

	policies, _, err := s.store.AccessPolicyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	for _, policy := range policies {
		if policy.Action != models.PolicyActionDeny {
			continue
		}

		matched, err := policyApplies(policy, dev, principal, member.Role, login, sourceIP)
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

		matched, err := policyApplies(policy, dev, principal, member.Role, login, sourceIP)
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

	if member.IsService() {
		requireReauth = false
		reauthPeriod = nil
	}

	return &models.Decision{Allowed: true, RequireReauth: requireReauth, ReauthPeriod: reauthPeriod}, nil
}

func policyApplies(policy models.AccessPolicy, dev *models.Device, principal models.Principal, role authorizer.Role, login, sourceIP string) (bool, error) {
	if !subjectMatches(policy.Subject, principal, role) {
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

func validateAccessPolicySubject(namespace *models.Namespace, apiKeys []models.APIKey, subject requests.AccessPolicySubject) error {
	switch models.PolicySubjectType(subject.Type) {
	case models.PolicySubjectUser:
		if _, ok := namespace.FindMember(subject.Value); !ok {
			return NewErrAccessPolicyInvalidField(map[string]string{
				"subject.value": "must be a member of this namespace",
			})
		}
	case models.PolicySubjectRole:
		if authorizer.RoleFromString(subject.Value) == authorizer.RoleInvalid {
			return NewErrAccessPolicyInvalidField(map[string]string{
				"subject.value": "must be a role this namespace defines",
			})
		}
	case models.PolicySubjectAPIKey:
		if !apiKeyExists(apiKeys, subject.Value) {
			return NewErrAccessPolicyInvalidField(map[string]string{
				"subject.value": "must be an API key of this namespace",
			})
		}
	case models.PolicySubjectAllMembers:
		if subject.Value != "" {
			return NewErrAccessPolicyInvalidField(map[string]string{
				"subject.value": "must be empty when the subject is every member",
			})
		}
	}

	return nil
}

func apiKeyExists(apiKeys []models.APIKey, id string) bool {
	for _, key := range apiKeys {
		if key.ID == id {
			return true
		}
	}

	return false
}

// subjectMatches reports whether a policy's subject names this principal. Taking the principal
// rather than a bare id is what keeps an API key's id from being compared against a user
// subject: the kind decides which subjects can match at all, and only a person is a member.
func subjectMatches(subject models.PolicySubject, principal models.Principal, role authorizer.Role) bool {
	switch subject.Type {
	case models.PolicySubjectAllMembers:
		return principal.Kind == models.PrincipalUser
	case models.PolicySubjectRole:
		return principal.Kind == models.PrincipalUser && subject.Value == role.String()
	case models.PolicySubjectUser:
		return principal.Kind != models.PrincipalAPIKey && subject.Value == principal.ID
	case models.PolicySubjectAPIKey:
		return principal.Kind == models.PrincipalAPIKey && subject.Value == principal.ID
	default:
		return false
	}
}

func loginMatches(logins []string, login string) bool {
	for _, l := range logins {
		if l == "*" || l == login {
			return true
		}
	}

	return false
}

func defaultAction(action string) models.PolicyAction {
	if action == "" {
		return models.PolicyActionAllow
	}

	return models.PolicyAction(action)
}

func normalizeReauthPeriod(period *int) *int {
	if period != nil && *period == 0 {
		return nil
	}

	return period
}

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

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(tenantID, err)
	}

	apiKeys, _, err := s.store.APIKeyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	policies, _, err := s.store.AccessPolicyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	for i := range policies {
		policies[i].SubjectMatches = subjectMatchesAnyPrincipal(namespace, apiKeys, policies[i].Subject)
	}

	return policies, nil
}

// subjectMatchesAnyPrincipal reports whether anything in the namespace the subject could name
// exists. Leaving the API keys out would render a live SSH-granting policy as matching nobody,
// which is the defect this flag was added to fix, returning in a new shape.
func subjectMatchesAnyPrincipal(namespace *models.Namespace, apiKeys []models.APIKey, subject models.PolicySubject) bool {
	for _, member := range namespace.Members {
		principal := models.Principal{Kind: principalKindOfMember(member), ID: member.ID}
		if subjectMatches(subject, principal, member.Role) {
			return true
		}
	}

	for _, key := range apiKeys {
		principal := models.Principal{Kind: models.PrincipalAPIKey, ID: key.ID}
		if subjectMatches(subject, principal, key.Role) {
			return true
		}
	}

	return false
}

func principalKindOfMember(member models.Member) models.PrincipalKind {
	if member.Type == models.UserTypeService {
		return models.PrincipalService
	}

	return models.PrincipalUser
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

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	apiKeys, _, err := s.store.APIKeyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	policy.SubjectMatches = subjectMatchesAnyPrincipal(namespace, apiKeys, policy.Subject)

	return policy, nil
}

func (s *service) CreateAccessPolicy(ctx context.Context, req *requests.AccessPolicyCreate) (*models.AccessPolicy, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	apiKeys, _, err := s.store.APIKeyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	if err := validateAccessPolicySubject(namespace, apiKeys, req.Subject); err != nil {
		return nil, err
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

	created, err := s.store.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, id)
	if err != nil {
		return nil, err
	}

	created.SubjectMatches = subjectMatchesAnyPrincipal(namespace, apiKeys, created.Subject)

	return created, nil
}

func (s *service) UpdateAccessPolicy(ctx context.Context, req *requests.AccessPolicyUpdate) (*models.AccessPolicy, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	if _, err := s.store.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, req.ID); err != nil {
		return nil, NewErrAccessPolicyNotFound(req.ID, err)
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	apiKeys, _, err := s.store.APIKeyList(ctx, sc)
	if err != nil {
		return nil, err
	}

	if err := validateAccessPolicySubject(namespace, apiKeys, req.Subject); err != nil {
		return nil, err
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

	updated, err := s.store.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, req.ID)
	if err != nil {
		return nil, err
	}

	updated.SubjectMatches = subjectMatchesAnyPrincipal(namespace, apiKeys, updated.Subject)

	return updated, nil
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

func (s *service) resolveAccessPolicyFilter(ctx context.Context, sc scope.Scope, reqFilter requests.AccessPolicyFilter) (models.PublicKeyFilter, error) {
	var filter models.PublicKeyFilter

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
