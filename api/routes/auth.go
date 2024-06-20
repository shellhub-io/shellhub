package routes

import (
	"errors"
	"net/http"
	"strconv"

	jwt "github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mitchellh/mapstructure"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/api/routes/errors"
	svc "github.com/shellhub-io/shellhub/api/services"
	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	AuthRequestURL           = "/auth"
	AuthDeviceURL            = "/devices/auth"
	AuthDeviceURLV2          = "/auth/device"
	AuthUserURL              = "/login"
	AuthUserURLV2            = "/auth/user"
	AuthUserTokenInternalURL = "/auth/token/:id"     //nolint:gosec
	AuthUserTokenPublicURL   = "/auth/token/:tenant" //nolint:gosec
	AuthPublicKeyURL         = "/auth/ssh"
	AuthMFAURL               = "/auth/mfa"
)

const (
	// AuthRequestUserToken is the type of the token used to authenticate a user.
	AuthRequestUserToken = "user"
	// AuthRequestDeviceToken is the type of the token used to authenticate a device.
	AuthRequestDeviceToken = "device"
)

// AuthRequest checks the user and device authentication token.
//
// This route is a special route and it is called every time a user tries to access a route which requires
// authentication. It gets the JWT token sent, unwraps it and sets the information, like tenant, user, etc., as headers
// of the response to be got in the subsequent through the [gateway.Context].
func (h *Handler) AuthRequest(c gateway.Context) error {
	if key := c.Request().Header.Get("X-API-Key"); key != "" {
		apiKey, err := h.service.AuthAPIKey(c.Ctx(), key)
		if err != nil {
			return err
		}

		c.Response().Header().Set("X-Tenant-ID", apiKey.TenantID)
		c.Response().Header().Set("X-Role", apiKey.Role.String())
		c.Response().Header().Set("X-API-KEY", key)

		return c.NoContent(http.StatusOK)
	}

	token, ok := c.Get(middleware.DefaultJWTConfig.ContextKey).(*jwt.Token)
	if !ok {
		return svc.ErrTypeAssertion
	}

	rawClaims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return svc.ErrTypeAssertion
	}

	// setHeader sets a reader to the HTTP response to be read in the subsequent request.
	setHeader := func(response gateway.Context, key string, value string) {
		response.Response().Header().Set(key, value)
	}

	// decodeMap parses the JWT claims into a struct.
	decodeMap := func(input *jwt.MapClaims, output any) error {
		config := &mapstructure.DecoderConfig{
			TagName:  "json",
			Metadata: nil,
			Result:   output,
		}

		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return err
		}

		return decoder.Decode(input)
	}

	switch (*rawClaims)["claims"] {
	case AuthRequestUserToken:
		claims := new(models.UserAuthClaims)
		if err := decodeMap(rawClaims, claims); err != nil {
			return err
		}

		// The TenantID is optional as the user may not be part of any namespace.
		if claims.Tenant != "" {
			// The rawClaims contain only the tenant ID of the namespace and not the user's role. This is because the role is a
			// dynamic attribute, and a JWT token must be stateless (the role can change, but the token cannot). For this reason,
			// we need to retrieve the role every time this middleware is invoked (generally from the cache; see the [method]
			// signature for more info).
			if err := h.service.FillClaimsRole(c.Ctx(), claims); err != nil {
				return err
			}
		}

		args := c.QueryParam("args")
		if args != "skip" && claims.Tenant != "" {
			// This forces any no cached token to be invalid, even if it not not expired.
			if ok, err := h.service.AuthIsCacheToken(c.Ctx(), claims.Tenant, claims.ID); err != nil || !ok {
				return svc.NewErrAuthUnathorized(err)
			}
		}

		c.Response().Header().Set("X-ID", claims.ID)
		c.Response().Header().Set("X-Username", claims.Username)
		c.Response().Header().Set("X-Tenant-ID", claims.Tenant)
		c.Response().Header().Set("X-Role", claims.Role.String())

		return c.NoContent(http.StatusOK)
	case AuthRequestDeviceToken:
		var claims models.DeviceAuthClaims

		if err := decodeMap(rawClaims, &claims); err != nil {
			return err
		}

		// Extract device UID from JWT and set it into the header.
		setHeader(c, client.DeviceUIDHeader, claims.UID)
		setHeader(c, "X-Tenant-ID", claims.Tenant)

		return c.NoContent(http.StatusOK)
	default:

		return svc.NewErrAuthUnathorized(nil)
	}
}

func (h *Handler) AuthDevice(c gateway.Context) error {
	var req requests.DeviceAuth
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	ip := c.Request().Header.Get("X-Real-IP")
	res, err := h.service.AuthDevice(c.Ctx(), req, ip)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) AuthUser(c gateway.Context) error {
	req := new(requests.UserAuth)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, lockout, mfaToken, err := h.service.AuthUser(c.Ctx(), req, c.RealIP())
	c.Response().Header().Set("X-Account-Lockout", strconv.FormatInt(lockout, 10))
	c.Response().Header().Set("X-MFA-Token", mfaToken)

	if lockout > 0 {
		return c.NoContent(http.StatusTooManyRequests)
	}

	if mfaToken != "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	if err != nil {
		switch {
		case errors.Is(err, svc.ErrUserNotFound):
			return errs.NewErrUnauthorized(err)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) CreateUserToken(c gateway.Context) error {
	req := new(requests.CreateUserToken)

	if err := c.Bind(req); err != nil {
		return err
	}

	// If the tenant id is not in the parameters, we get it from the header.
	if req.TenantID == "" {
		req.TenantID = c.Request().Header.Get("X-Tenant-ID")
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.CreateUserToken(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) AuthPublicKey(c gateway.Context) error {
	var req requests.PublicKeyAuth
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.AuthPublicKey(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, ok := c.Get("ctx").(*gateway.Context)
		if !ok {
			return svc.ErrTypeAssertion
		}

		apiKey := c.Request().Header.Get("X-API-KEY")
		if apiKey == "" {
			jwt := middleware.JWTWithConfig(middleware.JWTConfig{ //nolint:staticcheck
				Claims:        &jwt.MapClaims{},
				SigningKey:    ctx.Service().(svc.Service).PublicKey(),
				SigningMethod: "RS256",
			})

			return jwt(next)(c)
		}

		return next(c)
	}
}
