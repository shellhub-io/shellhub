package services

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// SignInService is the only issuer of user tokens. A user is signed in through
// [SignInService.AuthLocalUser] or [SignInService.SignIn], which both refuse the user the
// instance's or the account's sign-in rules exclude; a bearer is replaced through
// [SignInService.RenewUserToken]; an instance administrator acts as a user through
// [SignInService.Impersonate].
type SignInService interface {
	// AuthLocalUser signs in a user with origin [models.UserOriginLocal] by password. A user is
	// locked out after 3 wrong passwords from one source IP, and a user with MFA enabled is handed
	// a challenge instead of a token.
	//
	// The token carries the user's preferred namespace, or the first one the user was added to,
	// and no tenant when the user is a member of none. It returns a timestamp when the lockout
	// ends if the user is locked out, and, when MFA is enabled, a token keying a
	// [models.PendingMFASignIn] in the cache under mfa-signin={token} for 30 minutes.
	//
	// It returns the store's error as is when the system settings cannot be read, and
	// [ErrAuthMethodNotAllowed] when local authentication is off. It returns
	// [ErrAuthUnathorized] for an unknown identifier, a user without local authentication, a
	// lockout or a wrong password, [ErrUserNotConfirmed] for an unconfirmed user and
	// [ErrUserAwaitingApproval] for a user an admin has not approved yet, both only once the
	// password matched. It returns [ErrTokenSigned] when the token cannot be signed, and
	// [ErrUserUpdate] when the user's last login or preferred namespace cannot be saved.
	AuthLocalUser(ctx context.Context, req *requests.AuthLocalUser, sourceIP string) (res *models.UserAuthResponse, lockout int64, mfaToken string, err error)
	// SignIn signs in the user userID names, whom the caller has authenticated by method, under
	// the same rules as [SignInService.AuthLocalUser]: the instance must have method on, the user
	// must sign in with it, and the account must be confirmed and approved. It records the sign-in
	// as the user's last login, and the token carries the namespace AuthLocalUser's would.
	//
	// It returns [ErrAuthMethodNotAllowed] when the instance has method off, or when method is
	// not local and no edition allowed it through [SignInService.AllowSignInMethod]; the error
	// that edition's check returns; the store's error as is when the system settings cannot be
	// read; [ErrUserNotFound] when the user cannot be resolved; [ErrAuthUnathorized] when the user
	// does not sign in with method; [ErrUserNotConfirmed] and [ErrUserAwaitingApproval] as
	// AuthLocalUser does; [ErrTokenSigned] when the token cannot be signed; and [ErrUserUpdate]
	// when the last login or the preferred namespace cannot be saved. ctx bounds the store and
	// cache calls and is handed to the method's check.
	SignIn(ctx context.Context, userID string, method models.UserAuthMethod) (*models.UserAuthResponse, error)
	// SignInAt signs in like [SignInService.SignIn] a user whose credential the caller checked
	// while the user's tokens were at tokenVersion, such as the password step of an MFA sign-in.
	// The token carries tokenVersion, so a revocation that lands after the check still refuses
	// it. Besides what SignIn returns, it returns [ErrAuthUnathorized] when the user's token
	// version is no longer tokenVersion, as after a revocation. ctx is used as SignIn uses it.
	SignInAt(ctx context.Context, userID string, method models.UserAuthMethod, tokenVersion int) (*models.UserAuthResponse, error)
	// RenewUserToken replaces the bearer of an already signed-in user, optionally for the
	// namespace req.TenantID names, which then becomes the user's preferred one. Without a tenant
	// it uses the preferred namespace or the first one the user was added to. The token carries
	// req.TokenVersion when it is set, so renewing a bearer never yields one newer than the
	// bearer, and stays impersonated when req.Impersonated is set.
	//
	// Unless req.Impersonated is set, the user must still be admitted to sign in: the account
	// confirmed and approved, and at least one of the user's methods on for the instance, as
	// [SignInService.SignIn] checks it.
	//
	// It returns [ErrUserNotFound] when the user cannot be resolved, and [ErrAuthUnathorized],
	// wrapping the reason, when the user is no longer admitted to sign in. It returns the store's
	// error as is when the system settings cannot be read, and an edition's check error other than
	// [ErrAuthMethodNotAllowed] as is. It returns [ErrNamespaceNotFound] when req.TenantID names no
	// namespace, [ErrNamespaceMemberNotFound] when the user is not a member of the namespace,
	// [ErrUserUpdate] when the preferred namespace cannot be saved, and [ErrTokenSigned] when the
	// token cannot be signed.
	RenewUserToken(ctx context.Context, req *requests.RenewUserToken) (*models.UserAuthResponse, error)
	// Impersonate issues the token an instance administrator acts as the user userID names with.
	// The token is marked impersonated, and every token renewed from it stays so. It returns
	// what [SignInService.RenewUserToken] returns for a request without a tenant. ctx bounds the
	// store and cache calls.
	Impersonate(ctx context.Context, userID string) (*models.UserAuthResponse, error)
	// AllowSignInMethod lets [SignInService.SignIn] sign users in by method, a method other than
	// local that an edition implements, while enabled returns nil. enabled is called with the
	// caller's ctx on each sign-in and each renewal it gates. It must be called before the server
	// serves requests; a method no edition allowed is refused.
	AllowSignInMethod(method models.UserAuthMethod, enabled func(ctx context.Context) error)
}

const pendingMFASignInTTL = 30 * time.Minute

func (s *service) AllowSignInMethod(method models.UserAuthMethod, enabled func(ctx context.Context) error) {
	if s.signInMethods == nil {
		s.signInMethods = map[models.UserAuthMethod]func(ctx context.Context) error{}
	}

	s.signInMethods[method] = enabled
}

func (s *service) AuthLocalUser(ctx context.Context, req *requests.AuthLocalUser, sourceIP string) (*models.UserAuthResponse, int64, string, error) {
	if err := s.signInMethodEnabled(ctx, models.UserAuthMethodLocal); err != nil {
		return nil, 0, "", err
	}

	user, err := store.UserResolveByAuthIdentifier(ctx, s.store, req.Identifier)
	if err != nil {
		return nil, 0, "", NewErrAuthUnathorized(nil)
	}

	if !slices.Contains(user.Preferences.AuthMethods, models.UserAuthMethodLocal) {
		return nil, 0, "", NewErrAuthUnathorized(nil)
	}

	if lockout, attempt, _ := s.cache.HasAccountLockout(ctx, sourceIP, user.ID); lockout > 0 {
		log.
			WithFields(log.Fields{
				"lockout":   lockout,
				"attempt":   attempt,
				"source_ip": sourceIP,
				"user_id":   user.ID,
			}).
			Warn("attempt to login blocked")

		return nil, lockout, "", NewErrAuthUnathorized(nil)
	}

	if !user.Password.Compare(req.Password) {
		lockout, _, err := s.cache.StoreLoginAttempt(ctx, sourceIP, user.ID)
		if err != nil {
			log.WithError(err).
				WithField("source_ip", sourceIP).
				WithField("user_id", user.ID).
				Warn("unable to store login attempt")
		}

		return nil, lockout, "", NewErrAuthUnathorized(nil)
	}

	if err := s.cache.ResetLoginAttempts(ctx, sourceIP, user.ID); err != nil {
		log.WithError(err).
			WithField("source_ip", sourceIP).
			WithField("user_id", user.ID).
			Warn("unable to reset authentication attempts")
	}

	if err := accountAdmitted(user); err != nil {
		return nil, 0, "", err
	}

	if user.MFA.Enabled {
		mfaToken := uuid.Generate()
		pending := &models.PendingMFASignIn{UserID: user.ID, TokenVersion: user.TokenVersion}
		if err := s.cache.Set(ctx, "mfa-signin={"+mfaToken+"}", pending, pendingMFASignInTTL); err != nil {
			log.WithError(err).
				WithField("source_ip", sourceIP).
				WithField("user_id", user.ID).
				Warn("unable to store mfa-token")
		}

		return nil, 0, mfaToken, nil
	}

	if !strings.HasPrefix(user.Password.Hash, "$") {
		if neo, _ := models.HashUserPassword(req.Password); neo.Hash != "" {
			user.Password = neo
		}
	}

	res, err := s.issueSignIn(ctx, user, user.TokenVersion)
	if err != nil {
		return nil, 0, "", err
	}

	return res, 0, "", nil
}

func (s *service) SignIn(ctx context.Context, userID string, method models.UserAuthMethod) (*models.UserAuthResponse, error) {
	return s.signIn(ctx, userID, method, nil)
}

func (s *service) SignInAt(ctx context.Context, userID string, method models.UserAuthMethod, tokenVersion int) (*models.UserAuthResponse, error) {
	return s.signIn(ctx, userID, method, &tokenVersion)
}

func (s *service) signIn(ctx context.Context, userID string, method models.UserAuthMethod, tokenVersion *int) (*models.UserAuthResponse, error) {
	if err := s.signInMethodEnabled(ctx, method); err != nil {
		return nil, err
	}

	user, err := s.store.UserResolve(ctx, store.UserIDResolver, userID)
	if err != nil {
		return nil, NewErrUserNotFound(userID, err)
	}

	if !slices.Contains(user.Preferences.AuthMethods, method) {
		return nil, NewErrAuthUnathorized(nil)
	}

	if err := accountAdmitted(user); err != nil {
		return nil, err
	}

	version := user.TokenVersion
	if tokenVersion != nil {
		if *tokenVersion != user.TokenVersion {
			return nil, NewErrAuthUnathorized(nil)
		}

		version = *tokenVersion
	}

	return s.issueSignIn(ctx, user, version)
}

func (s *service) signInMethodEnabled(ctx context.Context, method models.UserAuthMethod) error {
	if method == models.UserAuthMethodLocal {
		system, err := s.store.SystemGet(ctx)
		if err != nil {
			return err
		}

		if !system.Authentication.Local.Enabled {
			return NewErrAuthMethodNotAllowed(method.String())
		}

		return nil
	}

	enabled, ok := s.signInMethods[method]
	if !ok {
		return NewErrAuthMethodNotAllowed(method.String())
	}

	return enabled(ctx)
}

func (s *service) stillAdmitted(ctx context.Context, user *models.User) error {
	if err := accountAdmitted(user); err != nil {
		return err
	}

	refused := NewErrAuthMethodNotAllowed("")

	var failure error

	for _, method := range user.Preferences.AuthMethods {
		err := s.signInMethodEnabled(ctx, method)
		if err == nil {
			return nil
		}

		if signInRefused(err) {
			refused = err
		} else {
			failure = err
		}
	}

	if failure != nil {
		return failure
	}

	return refused
}

func signInRefused(err error) bool {
	return errors.Is(err, ErrAuthMethodNotAllowed) ||
		errors.Is(err, ErrUserNotConfirmed) ||
		errors.Is(err, ErrUserAwaitingApproval)
}

func accountAdmitted(user *models.User) error {
	if user.Status == models.UserStatusNotConfirmed {
		return NewErrUserNotConfirmed(nil)
	}

	if user.AwaitingApproval {
		return NewErrUserAwaitingApproval(nil)
	}

	return nil
}

func (s *service) issueSignIn(ctx context.Context, user *models.User, tokenVersion int) (*models.UserAuthResponse, error) {
	tenantID := ""

	var role *authorizer.Role

	if ns, _ := s.store.NamespaceGetPreferred(ctx, user.ID); ns != nil && ns.TenantID != "" {
		if m, _ := ns.FindMember(user.ID); m != nil {
			tenantID = ns.TenantID
			role = &m.Role
		}
	}

	token, err := s.mintUserToken(user, tenantID, tokenVersion, false)
	if err != nil {
		return nil, err
	}

	user.LastLogin = s.clock.Now()
	if err := s.store.UserUpdate(ctx, user); err != nil {
		return nil, NewErrUserUpdate(user, err)
	}

	if user.Preferences.PreferredNamespace != tenantID {
		if err := s.store.UserUpdatePreferredNamespace(ctx, user.ID, tenantID); err != nil {
			return nil, NewErrUserUpdate(user, err)
		}
	}

	if err := s.AuthCacheToken(ctx, tenantID, user.ID, token); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"id": user.ID}).
			Warn("unable to cache the authentication token")
	}

	return userAuthResponse(user, tenantID, role, token, false), nil
}

func (s *service) RenewUserToken(ctx context.Context, req *requests.RenewUserToken) (*models.UserAuthResponse, error) {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.UserID)
	if err != nil {
		return nil, NewErrUserNotFound(req.UserID, err)
	}

	if !req.Impersonated {
		if err := s.stillAdmitted(ctx, user); err != nil {
			if signInRefused(err) {
				return nil, NewErrAuthUnathorized(err)
			}

			return nil, err
		}
	}

	tenantID := ""

	var role *authorizer.Role

	switch req.TenantID {
	case "":
		namespace, err := s.store.NamespaceGetPreferred(ctx, user.ID)
		if err != nil {
			break
		}

		member, ok := namespace.FindMember(user.ID)
		if !ok {
			return nil, NewErrNamespaceMemberNotFound(user.ID, nil)
		}

		tenantID = namespace.TenantID
		role = &member.Role
	default:
		namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
		if err != nil {
			return nil, NewErrNamespaceNotFound(req.TenantID, err)
		}

		member, ok := namespace.FindMember(user.ID)
		if !ok {
			return nil, NewErrNamespaceMemberNotFound(user.ID, nil)
		}

		tenantID = namespace.TenantID
		role = &member.Role

		if user.Preferences.PreferredNamespace != namespace.TenantID {
			if err := s.store.UserUpdatePreferredNamespace(ctx, user.ID, tenantID); err != nil {
				return nil, NewErrUserUpdate(user, err)
			}
		}
	}

	tokenVersion := user.TokenVersion
	if req.TokenVersion != nil {
		tokenVersion = *req.TokenVersion
	}

	token, err := s.mintUserToken(user, tenantID, tokenVersion, req.Impersonated)
	if err != nil {
		return nil, err
	}

	if err := s.AuthCacheToken(ctx, tenantID, user.ID, token); err != nil {
		log.WithError(err).Warn("unable to cache the user's auth token")
	}

	return userAuthResponse(user, tenantID, role, token, req.Impersonated), nil
}

func (s *service) Impersonate(ctx context.Context, userID string) (*models.UserAuthResponse, error) {
	return s.RenewUserToken(ctx, &requests.RenewUserToken{UserID: userID, Impersonated: true})
}

func (s *service) mintUserToken(user *models.User, tenantID string, tokenVersion int, impersonated bool) (string, error) {
	claims := authorizer.UserClaims{
		ID:           user.ID,
		Origin:       user.Origin.String(),
		TenantID:     tenantID,
		Username:     user.Username,
		MFA:          user.MFA.Enabled,
		Admin:        user.Admin,
		TokenVersion: tokenVersion,
		Impersonated: impersonated,
	}

	token, err := jwttoken.EncodeUserClaims(claims, s.issuer, s.privKey, s.clock.Now())
	if err != nil {
		return "", NewErrTokenSigned(err)
	}

	return token, nil
}

func userAuthResponse(user *models.User, tenantID string, role *authorizer.Role, token string, impersonated bool) *models.UserAuthResponse {
	res := &models.UserAuthResponse{
		ID:            user.ID,
		Origin:        user.Origin.String(),
		AuthMethods:   user.Preferences.AuthMethods,
		User:          user.Username,
		Name:          user.Name,
		Email:         user.Email,
		RecoveryEmail: user.RecoveryEmail,
		MFA:           user.MFA.Enabled,
		Role:          role,
		Token:         token,
		MaxNamespaces: user.MaxNamespaces,
		Admin:         user.Admin,

		OnboardingSurvey: user.Preferences.OnboardingSurvey,
		Impersonated:     impersonated,
	}

	if tenantID != "" {
		res.Tenant = &tenantID
	}

	return res
}
