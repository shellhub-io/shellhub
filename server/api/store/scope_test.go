package store_test

import (
	"maps"
	"reflect"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/storetest"
	"github.com/stretchr/testify/assert"
)

var globalMethods = map[string]string{
	"UserCreate":                   "users are global, not namespace-owned",
	"UserDelete":                   "users are global, not namespace-owned",
	"UserGetInfo":                  "users are global, not namespace-owned",
	"UserList":                     "users are global, not namespace-owned",
	"UserResolve":                  "users are global, not namespace-owned",
	"UserRevokeTokens":             "users are global, not namespace-owned",
	"UserUpdate":                   "users are global, not namespace-owned",
	"UserUpdateOnboardingSurvey":   "users are global, not namespace-owned",
	"UserUpdatePreferredNamespace": "users are global, not namespace-owned",
	"UserUpdateRevokingTokens":     "users are global, not namespace-owned",
	"UserMembershipInvitationList": "a user's invitations span the namespaces inviting them",
	"UserInvitationGet":            "user invitations are global, not namespace-owned",
	"UserInvitationUpdate":         "user invitations are global, not namespace-owned",
	"UserInvitationsUpsert":        "user invitations are global, not namespace-owned",
	"InstanceAPIKeyCreate":         "instance API keys belong to the instance, not a namespace",
	"InstanceAPIKeyDelete":         "instance API keys belong to the instance, not a namespace",
	"InstanceAPIKeyList":           "instance API keys belong to the instance, not a namespace",
	"InstanceAPIKeyResolve":        "instance API keys belong to the instance, not a namespace",
	"SystemGet":                    "the system row belongs to the instance",
	"SystemLock":                   "the system row belongs to the instance",
	"SystemSet":                    "the system row belongs to the instance",
	"InstanceUsage":                "instance statistics aggregate every namespace",
	"NamespaceUsages":              "instance statistics aggregate every namespace",
	"EnrollmentCallbackRedeem":     "an enrollment callback is redeemed by its token before any namespace is known",
	"EnrollmentCallbackCleanup":    "enrollment callbacks expire by age, whatever namespace they enroll into",
	"NamespaceCreate":              "the namespace root addresses namespaces themselves",
	"NamespaceList":                "the namespace root addresses namespaces themselves",
	"NamespaceConflicts":           "the namespace root addresses namespaces themselves",
	"NamespaceResolve":             "the namespace root addresses namespaces themselves",
	"NamespaceGetPreferred":        "the namespace root addresses namespaces themselves",
	"NamespaceLockOwner":           "the namespace root addresses namespaces themselves",
	"Options":                      "query options plumbing",
	"WithTransaction":              "transaction plumbing",
	"WithSavepoint":                "transaction plumbing",
}

var unconvertedMethods = map[string]string{ //nolint:gosec // method names and the slice converting them, not credentials
	"SessionCreate":                    "slice 2: session raw-tenant methods",
	"ActiveSessionCreate":              "slice 2: session raw-tenant methods",
	"ActiveSessionResolve":             "slice 2: session raw-tenant methods",
	"ActiveSessionCleanup":             "slice 2: session raw-tenant methods",
	"SessionEventsCreate":              "slice 2: session raw-tenant methods",
	"SessionEventsList":                "slice 2: session raw-tenant methods",
	"SessionEventsDelete":              "slice 2: session raw-tenant methods",
	"SessionEventsTimeline":            "slice 2: session raw-tenant methods",
	"SessionUpdateDeviceUID":           "slice 2: session raw-tenant methods",
	"SessionListExpired":               "slice 2: session raw-tenant methods",
	"SessionDeleteMany":                "slice 2: session raw-tenant methods",
	"TagCreate":                        "slice 2: tag raw-tenant methods",
	"TagDelete":                        "slice 2: tag raw-tenant methods",
	"TagPullFromTarget":                "slice 2: tag raw-tenant methods",
	"TagPushToTarget":                  "slice 2: tag raw-tenant methods",
	"TagUpdate":                        "slice 2: tag raw-tenant methods",
	"SSHIdentityConsume":               "slice 2: SSH identity raw-tenant methods",
	"SSHIdentityCreate":                "slice 2: SSH identity raw-tenant methods",
	"SSHIdentityDelete":                "slice 2: SSH identity raw-tenant methods",
	"SSHIdentityTouchLastUsed":         "slice 2: SSH identity raw-tenant methods",
	"SSHIdentityTouchReauth":           "slice 2: SSH identity raw-tenant methods",
	"SSHIdentityUpdate":                "slice 2: SSH identity raw-tenant methods",
	"NamespaceGetWithoutMembers":       "slice 2: namespace raw-tenant methods",
	"NamespaceGetDeviceLimit":          "slice 2: namespace raw-tenant methods",
	"NamespaceUpgrade":                 "slice 2: namespace raw-tenant methods",
	"NamespaceDeleteMany":              "slice 2: namespace raw-tenant methods",
	"APIKeyCreate":                     "slice 3: API key model-taking writes",
	"APIKeyDelete":                     "slice 3: API key model-taking writes",
	"APIKeyUpdate":                     "slice 3: API key model-taking writes",
	"AccessPolicyCreate":               "slice 3: access policy model-taking writes",
	"AccessPolicyDelete":               "slice 3: access policy model-taking writes",
	"AccessPolicyUpdate":               "slice 3: access policy model-taking writes",
	"MembershipInvitationCreate":       "slice 3: membership invitation model-taking writes",
	"MembershipInvitationUpdate":       "slice 3: membership invitation model-taking writes",
	"MembershipInvitationDelete":       "slice 3: membership invitation model-taking writes",
	"MembershipInvitationResolveBySig": "slice 3: membership invitation model-taking writes",
	"NamespaceUpdate":                  "slice 3: namespace model-taking writes and the device count sync",
	"NamespaceDelete":                  "slice 3: namespace model-taking writes and the device count sync",
	"NamespaceSyncDeviceCounts":        "slice 3: namespace model-taking writes and the device count sync",
	"ProvisioningKeyCreate":            "slice 3: provisioning key model-taking writes",
	"ProvisioningKeyEventCreate":       "slice 3: provisioning key model-taking writes",
	"ProvisioningKeyIncrementUsage":    "slice 3: provisioning key model-taking writes",
	"ProvisioningKeyUpdate":            "slice 3: provisioning key model-taking writes",
	"PublicKeyCreate":                  "slice 3: public key model-taking writes",
	"PublicKeyDelete":                  "slice 3: public key model-taking writes",
	"PublicKeyUpdate":                  "slice 3: public key model-taking writes",
	"SSHApprovalCreate":                "slice 3: SSH approval methods",
	"SSHApprovalGet":                   "slice 3: SSH approval methods",
	"SSHApprovalDecide":                "slice 3: SSH approval methods",
	"SSHApprovalCleanup":               "slice 3: SSH approval methods",
}

func TestEveryNamespaceOwnedStoreMethodTakesAScope(t *testing.T) {
	exempt := make(map[string]string, len(globalMethods)+len(unconvertedMethods))
	maps.Copy(exempt, globalMethods)
	maps.Copy(exempt, unconvertedMethods)

	assert.Empty(t, storetest.UnscopedMethods(reflect.TypeFor[store.Store](), exempt))
}
