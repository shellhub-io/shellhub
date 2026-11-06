package main

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessPolicySelectorTags covers who may change which devices an access policy selects: a tag
// a policy's filter holds is moved on or off a device only by a role that may edit the policy.
func TestAccessPolicySelectorTags(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, newNamespace)
	_, device := startAcceptedAgent(t, ctx, compose)

	sshid := deviceSSHID(device)

	t.Run("an operator does not attach a tag an allow policy selects by", func(t *testing.T) {
		member, signer := memberWithKey(t, compose, "attacher", authorizer.RoleOperator)

		createTag(t, compose, "selected")
		t.Cleanup(func() { deleteTag(t, compose, "selected") })

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "selected",
			Subject: userSubject(member.ID),
			Filter:  requests.AccessPolicyFilter{Tags: []string{"selected"}},
			Logins:  []string{"*"},
		})

		resp, err := asBearer(t, compose, member.Token).Post("/api/devices/" + device.UID + "/tags/selected")
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = compose.R(context.WithoutCancel(t.Context())).Delete("/api/devices/" + device.UID + "/tags/selected")
		})

		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, sshid, signer)
		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonNoGrant), "user="+member.ID)
	})

	t.Run("an operator does not detach a tag a deny policy selects by", func(t *testing.T) {
		member, signer := memberWithKey(t, compose, "detacher", authorizer.RoleOperator)

		createTag(t, compose, "fenced")
		t.Cleanup(func() { deleteTag(t, compose, "fenced") })
		tagDevice(t, compose, device.UID, "fenced")

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "operators",
			Subject: roleSubject(authorizer.RoleOperator),
			Filter:  requests.AccessPolicyFilter{All: true},
			Logins:  []string{"*"},
		})
		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "fenced",
			Subject: userSubject(member.ID),
			Filter:  requests.AccessPolicyFilter{Tags: []string{"fenced"}},
			Logins:  []string{"*"},
			Action:  string(models.PolicyActionDeny),
		})

		resp, err := asBearer(t, compose, member.Token).Delete("/api/devices/" + device.UID + "/tags/fenced")
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = compose.R(context.WithoutCancel(t.Context())).Post("/api/devices/" + device.UID + "/tags/fenced")
		})

		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, sshid, signer)
		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonDeniedByPolicy), "user="+member.ID)
	})

	t.Run("an operator attaches and detaches a tag no rule selects by", func(t *testing.T) {
		member := newMember(t, compose, "labeller", authorizer.RoleOperator)

		createTag(t, compose, "label")
		t.Cleanup(func() { deleteTag(t, compose, "label") })

		resp, err := asBearer(t, compose, member.Token).Post("/api/devices/" + device.UID + "/tags/label")
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())
		requireDeviceTags(t, compose, device.UID, "label")

		resp, err = asBearer(t, compose, member.Token).Delete("/api/devices/" + device.UID + "/tags/label")
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())
		requireDeviceTags(t, compose, device.UID)
	})
}

// TestPublicKeySelectorHostname covers who may rename a device a public key selects by name: a
// rename that moves the device out of the key's pattern needs a role that may edit the key, and
// one between two names the pattern matches does not.
func TestPublicKeySelectorHostname(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, legacyNamespace)
	_, device := startAcceptedAgent(t, ctx, compose)

	kept := device.Name + "-kept"
	pattern := regexp.QuoteMeta(device.Name) + "|" + regexp.QuoteMeta(kept)
	signer := registerPublicKey(t, compose, ".*", requests.PublicKeyFilter{Hostname: pattern})
	member := newMember(t, compose, "renamer", authorizer.RoleOperator)

	requireDeviceName := func(t *testing.T, name string) {
		t.Helper()

		current, resp, err := compose.GetDevice(t.Context(), device.UID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, name, current.Name)
	}

	rename := func(t *testing.T, name string) int {
		t.Helper()

		resp, err := asBearer(t, compose, member.Token).SetBody(map[string]string{"name": name}).Put("/api/devices/" + device.UID)
		require.NoError(t, err)

		return resp.StatusCode()
	}

	t.Run("an operator does not rename a device out of a key's name pattern", func(t *testing.T) {
		t.Cleanup(func() {
			_, _ = compose.R(context.WithoutCancel(t.Context())).SetBody(map[string]string{"name": device.Name}).Put("/api/devices/" + device.UID)
		})

		assert.Equal(t, http.StatusForbidden, rename(t, "renamed-away"))
		requireDeviceName(t, device.Name)
		requireKeyLogsIn(t, compose, deviceSSHID(device), signer)
	})

	t.Run("an operator renames a device within a key's name pattern", func(t *testing.T) {
		t.Cleanup(func() {
			_, _ = compose.R(context.WithoutCancel(t.Context())).SetBody(map[string]string{"name": device.Name}).Put("/api/devices/" + device.UID)
		})

		require.Equal(t, http.StatusOK, rename(t, kept))
		requireDeviceName(t, kept)
	})
}
