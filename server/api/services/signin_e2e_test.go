package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignIn(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	st := provider.Store()
	provider.Clock().Set(now)

	newUser := func(t *testing.T, username string, change func(*models.User)) string {
		t.Helper()

		user := &models.User{
			Origin:        models.UserOriginLocal,
			Status:        models.UserStatusConfirmed,
			MaxNamespaces: -1,
			UserData:      models.UserData{Name: username, Email: username + "@example.com", Username: username},
			Password:      models.UserPassword{Hash: "hash"},
			Preferences:   models.UserPreferences{AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal}},
		}
		if change != nil {
			change(user)
		}

		id, err := st.UserCreate(ctx, user)
		require.NoError(t, err)

		return id
	}

	setLocalSignIn := func(t *testing.T, enabled bool) {
		t.Helper()

		system, err := st.SystemGet(ctx)
		require.NoError(t, err)

		system.Authentication.Local.Enabled = enabled
		require.NoError(t, st.SystemSet(ctx, system))
	}

	samlEnabled := true
	samlBroken := false
	svc := NewService(st, privateKey, publicKey, storecache.NewNullCache(), WithClock(provider.Clock()))
	svc.AllowSignInMethod(models.UserAuthMethodSAML, func(context.Context) error {
		if samlBroken {
			return errors.New("the identity provider's metadata cannot be parsed")
		}

		if !samlEnabled {
			return NewErrAuthMethodNotAllowed(models.UserAuthMethodSAML.String())
		}

		return nil
	})

	cases := []struct {
		description string
		username    string
		change      func(*models.User)
		method      models.UserAuthMethod
		localOff    bool
		samlOff     bool
		expected    error
	}{
		{
			description: "signs a confirmed user in by a method the instance and the user allow",
			username:    "allowed",
			method:      models.UserAuthMethodLocal,
		},
		{
			description: "refuses local sign-in while the instance has it turned off",
			username:    "local-off",
			method:      models.UserAuthMethodLocal,
			localOff:    true,
			expected:    ErrAuthMethodNotAllowed,
		},
		{
			description: "refuses a method the user does not sign in with",
			username:    "saml-only",
			change: func(u *models.User) {
				u.Preferences.AuthMethods = []models.UserAuthMethod{models.UserAuthMethodSAML}
			},
			method:   models.UserAuthMethodLocal,
			expected: ErrAuthUnathorized,
		},
		{
			description: "refuses a user who has not confirmed the account",
			username:    "unconfirmed",
			change:      func(u *models.User) { u.Status = models.UserStatusNotConfirmed },
			method:      models.UserAuthMethodLocal,
			expected:    ErrUserNotConfirmed,
		},
		{
			description: "refuses a user an administrator has not approved",
			username:    "unapproved",
			change:      func(u *models.User) { u.AwaitingApproval = true },
			method:      models.UserAuthMethodLocal,
			expected:    ErrUserAwaitingApproval,
		},
		{
			description: "signs in by a method an edition allowed while it is on",
			username:    "saml-on",
			change: func(u *models.User) {
				u.Preferences.AuthMethods = []models.UserAuthMethod{models.UserAuthMethodSAML}
			},
			method: models.UserAuthMethodSAML,
		},
		{
			description: "refuses a method an edition allowed while it is off",
			username:    "saml-off",
			change: func(u *models.User) {
				u.Preferences.AuthMethods = []models.UserAuthMethod{models.UserAuthMethodSAML}
			},
			method:   models.UserAuthMethodSAML,
			samlOff:  true,
			expected: ErrAuthMethodNotAllowed,
		},
		{
			description: "refuses a method no edition allowed",
			username:    "unknown-method",
			method:      "oidc",
			expected:    ErrAuthMethodNotAllowed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			id := newUser(t, tc.username, tc.change)

			setLocalSignIn(t, !tc.localOff)
			samlEnabled = !tc.samlOff

			res, err := svc.SignIn(ctx, id, tc.method)
			if tc.expected != nil {
				require.ErrorIs(t, err, tc.expected)
				assert.Nil(t, res)

				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, res.Token)
			assert.Equal(t, id, res.ID)

			user, err := st.UserResolve(ctx, store.UserIDResolver, id)
			require.NoError(t, err)
			assert.WithinDuration(t, now, user.LastLogin, time.Microsecond)
		})
	}

	t.Run("signs in at the token version the caller checked", func(t *testing.T) {
		id := newUser(t, "at-version", nil)
		setLocalSignIn(t, true)

		res, err := svc.SignInAt(ctx, id, models.UserAuthMethodLocal, 0)
		require.NoError(t, err)

		claims, err := jwttoken.ClaimsFromBearerToken(publicKey, res.Token)
		require.NoError(t, err)
		userClaims, ok := claims.(*authorizer.UserClaims)
		require.True(t, ok)
		assert.Equal(t, 0, userClaims.TokenVersion)
	})

	t.Run("refuses a sign-in checked before the user's tokens were revoked", func(t *testing.T) {
		id := newUser(t, "revoked-since", nil)
		setLocalSignIn(t, true)
		require.NoError(t, st.UserRevokeTokens(ctx, id))

		res, err := svc.SignInAt(ctx, id, models.UserAuthMethodLocal, 0)
		require.ErrorIs(t, err, ErrAuthUnathorized)
		assert.Nil(t, res)
	})

	renewals := []struct {
		description  string
		username     string
		change       func(*models.User)
		impersonated bool
		localOff     bool
		samlOff      bool
		samlBroken   bool
		expected     error
	}{
		{
			description: "renews the token of a user the sign-in policy still admits",
			username:    "renew-allowed",
		},
		{
			description: "refuses to renew while the instance has the user's only method off",
			username:    "renew-local-off",
			localOff:    true,
			expected:    ErrAuthUnathorized,
		},
		{
			description: "renews while another method of the user is on",
			username:    "renew-saml-on",
			change: func(u *models.User) {
				u.Preferences.AuthMethods = []models.UserAuthMethod{models.UserAuthMethodLocal, models.UserAuthMethodSAML}
			},
			localOff: true,
		},
		{
			description: "refuses to renew while every method of the user is off",
			username:    "renew-all-off",
			change: func(u *models.User) {
				u.Preferences.AuthMethods = []models.UserAuthMethod{models.UserAuthMethodLocal, models.UserAuthMethodSAML}
			},
			localOff: true,
			samlOff:  true,
			expected: ErrAuthUnathorized,
		},
		{
			description: "renews by a later method while an earlier one's check fails",
			username:    "renew-saml-broken",
			change: func(u *models.User) {
				u.Preferences.AuthMethods = []models.UserAuthMethod{models.UserAuthMethodSAML, models.UserAuthMethodLocal}
			},
			samlBroken: true,
		},
		{
			description: "refuses to renew for a user with no sign-in method",
			username:    "renew-no-method",
			change:      func(u *models.User) { u.Preferences.AuthMethods = nil },
			expected:    ErrAuthUnathorized,
		},
		{
			description: "refuses to renew for a user who has not confirmed the account",
			username:    "renew-unconfirmed",
			change:      func(u *models.User) { u.Status = models.UserStatusNotConfirmed },
			expected:    ErrAuthUnathorized,
		},
		{
			description: "refuses to renew for a user an administrator has not approved",
			username:    "renew-unapproved",
			change:      func(u *models.User) { u.AwaitingApproval = true },
			expected:    ErrAuthUnathorized,
		},
		{
			description:  "renews an impersonated token whatever the user's sign-in policy",
			username:     "renew-impersonated",
			change:       func(u *models.User) { u.AwaitingApproval = true },
			impersonated: true,
			localOff:     true,
		},
	}

	for _, tc := range renewals {
		t.Run(tc.description, func(t *testing.T) {
			id := newUser(t, tc.username, tc.change)

			setLocalSignIn(t, !tc.localOff)
			samlEnabled = !tc.samlOff
			samlBroken = tc.samlBroken

			res, err := svc.RenewUserToken(ctx, &requests.RenewUserToken{UserID: id, Impersonated: tc.impersonated})
			if tc.expected != nil {
				require.ErrorIs(t, err, tc.expected)
				assert.Nil(t, res)

				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, res.Token)
			assert.Equal(t, tc.impersonated, res.Impersonated)
		})
	}
}
