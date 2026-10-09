package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// TestTagsSelectPublicKeyDevices covers a tag as the device selector of a legacy public key. The key
// holds the tag itself, not its name, so a rename keeps the key reaching the devices carrying it.
// A filter left without its only tag would select every device, so the tag cannot be deleted
// while a key's filter holds it.
func TestTagsSelectPublicKeyDevices(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)
	_, tagged := startAcceptedAgent(t, ctx, compose)
	_, untagged := startAcceptedAgent(t, ctx, compose)

	t.Run("a renamed tag keeps selecting the devices carrying it", func(t *testing.T) {
		createTag(t, compose, "keybefore")
		t.Cleanup(func() { deleteTag(t, compose, "keyafter") })
		attachTag(t, compose, tagged.UID, "keybefore")
		signer := registerPublicKey(t, compose, ".*", requests.PublicKeyFilter{Tags: []string{"keybefore"}})

		renameTag(t, compose, "keybefore", "keyafter")

		assert.Equal(t, []string{"keyafter"}, tagNames(publicKey(t, compose, signer).Filter.Tags))
		requireKeyLogsIn(t, compose, deviceSSHID(tagged), signer)
		requireKeyRefused(t, compose, deviceSSHID(untagged), signer, errPublicKeyNotEvaluated)
	})

	t.Run("a tag a key's filter holds is not deleted", func(t *testing.T) {
		createTag(t, compose, "keyheld")
		t.Cleanup(func() { deleteTag(t, compose, "keyheld") })
		attachTag(t, compose, tagged.UID, "keyheld")
		signer := registerPublicKey(t, compose, ".*", requests.PublicKeyFilter{Tags: []string{"keyheld"}})

		requireTagInUse(t, compose, "keyheld", `public key "`+t.Name()+`"`)

		assert.Contains(t, listedTags(t, compose.R(t.Context())), "keyheld")
		assert.Equal(t, []string{"keyheld"}, tagNames(publicKey(t, compose, signer).Filter.Tags))
		requireDeviceTags(t, compose, tagged.UID, "keyheld")
		requireKeyLogsIn(t, compose, deviceSSHID(tagged), signer)
		requireKeyRefused(t, compose, deviceSSHID(untagged), signer, errPublicKeyNotEvaluated)
	})
}

// TestTagsSelectAccessPolicyDevices covers a tag as the device filter of an access policy. The
// policy holds the tag itself, not its name, so a rename keeps it granting the devices carrying the
// tag. A filter left without its only tag would grant every device, so the tag cannot be deleted
// while a policy's filter holds it. Each case logs in as a member only its own policy grants, so the
// owner's starter policy never decides it.
func TestTagsSelectAccessPolicyDevices(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
	_, tagged := startAcceptedAgent(t, ctx, compose)
	_, untagged := startAcceptedAgent(t, ctx, compose)

	t.Run("a renamed tag keeps granting the devices carrying it", func(t *testing.T) {
		member, signer := memberWithKey(t, compose, "renamed", authorizer.RoleOperator)
		createTag(t, compose, "policybefore")
		t.Cleanup(func() { deleteTag(t, compose, "policyafter") })
		attachTag(t, compose, tagged.UID, "policybefore")
		policy := grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "renamed",
			Subject: userSubject(member.ID),
			Filter:  requests.AccessPolicyFilter{Tags: []string{"policybefore"}},
			Logins:  []string{"*"},
		})

		renameTag(t, compose, "policybefore", "policyafter")

		assert.Equal(t, []string{"policyafter"}, tagNames(accessPolicy(t, compose, policy.ID).Filter.Tags))
		requireStraightThrough(t, compose, deviceSSHID(tagged), signer)
		requireRefusedAtAuth(t, compose, deviceSSHID(untagged), signer)
	})

	t.Run("a tag a policy's filter holds is not deleted", func(t *testing.T) {
		member, signer := memberWithKey(t, compose, "held", authorizer.RoleOperator)
		createTag(t, compose, "policyheld")
		t.Cleanup(func() { deleteTag(t, compose, "policyheld") })
		attachTag(t, compose, tagged.UID, "policyheld")
		policy := grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "held",
			Subject: userSubject(member.ID),
			Filter:  requests.AccessPolicyFilter{Tags: []string{"policyheld"}},
			Logins:  []string{"*"},
		})

		requireTagInUse(t, compose, "policyheld", `access policy "held"`)

		assert.Contains(t, listedTags(t, compose.R(t.Context())), "policyheld")
		assert.Equal(t, []string{"policyheld"}, tagNames(accessPolicy(t, compose, policy.ID).Filter.Tags))
		requireDeviceTags(t, compose, tagged.UID, "policyheld")
		requireStraightThrough(t, compose, deviceSSHID(tagged), signer)
		requireRefusedAtAuth(t, compose, deviceSSHID(untagged), signer)
	})
}

// TestTagDeleteDetachesItEverywhere covers a tag no filter holds: deleting it takes it off every
// device carrying it, whatever the device's status, and leaves the devices' other tags in place.
func TestTagDeleteDetachesItEverywhere(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	accepted := enrollDevice(t, compose, "detachaccepted", "02:00:00:00:11:01")
	compose.UpdateDeviceStatus(t, accepted, environment.DeviceActionAccept)
	pending := enrollDevice(t, compose, "detachpending", "02:00:00:00:11:02")

	createTag(t, compose, "detached")
	createTag(t, compose, "kept")

	for _, uid := range []string{accepted, pending} {
		attachTag(t, compose, uid, "detached")
		attachTag(t, compose, uid, "kept")
	}

	deleteTag(t, compose, "detached")

	assert.NotContains(t, listedTags(t, compose.R(t.Context())), "detached")
	requireDeviceTags(t, compose, accepted, "kept")
	requireDeviceTags(t, compose, pending, "kept")
}

// TestTagsAreNamespaceScoped covers two namespaces using the same tag name. Each namespace sees,
// attaches, renames and deletes only its own tag, so a change in one leaves the other untouched.
func TestTagsAreNamespaceScoped(t *testing.T) {
	const (
		neighbourUsername = "neighbour"
		neighbourTenant   = "11111111-1111-4111-1111-111111111111"
	)

	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	compose.NewUser(t, neighbourUsername, "neighbour@ossystems.com.br", ShellHubPassword)
	compose.NewNamespace(t, neighbourUsername, "neighbourhood", neighbourTenant, models.SSHAccessModeLegacy)
	neighbour := compose.AuthUser(t, neighbourUsername, ShellHubPassword)
	asNeighbour := func(ctx context.Context) *resty.Request {
		return compose.Anonymous(ctx).SetAuthToken(neighbour.Token)
	}

	own := enrollDevice(t, compose, "owndevice", "02:00:00:00:12:01")
	createTag(t, compose, "shared")
	attachTag(t, compose, own, "shared")

	neighbourDevice := newDeviceAuthRequest(t, "neighbourdevice", "02:00:00:00:12:02")
	neighbourDevice.TenantID = neighbourTenant
	theirs := authDevice(t, compose, neighbourDevice).UID

	t.Run("another namespace does not list the tag", func(t *testing.T) {
		assert.NotContains(t, listedTags(t, asNeighbour(t.Context())), "shared")
	})

	t.Run("another namespace cannot attach, rename or delete the tag", func(t *testing.T) {
		attempts := []struct {
			method string
			path   string
			body   any
		}{
			{method: http.MethodPost, path: "/api/devices/" + theirs + "/tags/shared"},
			{method: http.MethodPatch, path: "/api/tags/shared", body: map[string]string{"name": "taken"}},
			{method: http.MethodDelete, path: "/api/tags/shared"},
		}

		for _, attempt := range attempts {
			req := asNeighbour(t.Context())
			if attempt.body != nil {
				req = req.SetBody(attempt.body)
			}

			resp, err := req.Execute(attempt.method, attempt.path)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "%s %s: %s", attempt.method, attempt.path, resp.String())
		}

		assert.Contains(t, listedTags(t, compose.R(t.Context())), "shared")
		requireDeviceTags(t, compose, own, "shared")
	})

	t.Run("another namespace keeps its own tag of that name through the first one's rename and delete", func(t *testing.T) {
		resp, err := asNeighbour(t.Context()).SetBody(map[string]string{"name": "shared"}).Post("/api/tags")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		resp, err = asNeighbour(t.Context()).Post("/api/devices/" + theirs + "/tags/shared")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		renameTag(t, compose, "shared", "renamed")
		deleteTag(t, compose, "renamed")

		assert.Equal(t, []string{"shared"}, listedTags(t, asNeighbour(t.Context())))

		device := new(models.Device)
		resp, err = asNeighbour(t.Context()).SetResult(device).Get("/api/devices/" + theirs)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, []string{"shared"}, tagNames(device.Tags))
	})
}

func attachTag(t *testing.T, compose *environment.DockerCompose, uid, tag string) {
	t.Helper()

	resp, err := compose.R(t.Context()).Post("/api/devices/" + uid + "/tags/" + tag)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

func renameTag(t *testing.T, compose *environment.DockerCompose, from, to string) {
	t.Helper()

	resp, err := compose.R(t.Context()).SetBody(map[string]string{"name": to}).Patch("/api/tags/" + from)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

func deleteTag(t *testing.T, compose *environment.DockerCompose, name string) {
	t.Helper()

	resp, err := compose.R(context.WithoutCancel(t.Context())).Delete("/api/tags/" + name)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())
}

func requireTagInUse(t *testing.T, compose *environment.DockerCompose, name, holder string) {
	t.Helper()

	refusal := new(responses.Error)

	resp, err := compose.R(t.Context()).SetError(refusal).Delete("/api/tags/" + name)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())
	assert.Equal(t, map[string]string{"name": holder}, refusal.Fields)
}

func listedTags(t *testing.T, req *resty.Request) []string {
	t.Helper()

	tags := []models.Tag{}

	resp, err := req.SetQueryParam("per_page", "100").SetResult(&tags).Get("/api/tags")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return tagNames(tags)
}

func requireDeviceTags(t *testing.T, compose *environment.DockerCompose, uid string, tags ...string) {
	t.Helper()

	device, resp, err := compose.GetDevice(t.Context(), uid)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	assert.ElementsMatch(t, tags, tagNames(device.Tags))
}

func publicKey(t *testing.T, compose *environment.DockerCompose, signer ssh.Signer) models.PublicKey {
	t.Helper()

	keys := []models.PublicKey{}

	resp, err := compose.R(t.Context()).
		SetQueryParam("filter", sessionFilter(t, "fingerprint", "eq", ssh.FingerprintLegacyMD5(signer.PublicKey()))).
		SetResult(&keys).
		Get("/api/sshkeys/public-keys")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.Len(t, keys, 1)

	return keys[0]
}

func accessPolicy(t *testing.T, compose *environment.DockerCompose, id string) models.AccessPolicy {
	t.Helper()

	policy := models.AccessPolicy{}

	resp, err := compose.R(t.Context()).SetResult(&policy).Get("/api/access-policies/" + id)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return policy
}
