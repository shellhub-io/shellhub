package routes

import (
	"net/http"

	jwt "github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mitchellh/mapstructure"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	svc "github.com/shellhub-io/shellhub/api/services"
	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/request"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	AuthRequestURL   = "/auth"
	AuthDeviceURL    = "/devices/auth"
	AuthDeviceURLV2  = "/auth/device"
	AuthUserURL      = "/login"
	AuthUserURLV2    = "/auth/user"
	AuthUserTokenURL = "/auth/token/:tenant" //nolint:gosec
	AuthPublicKeyURL = "/auth/ssh"
)

func (h *Handler) AuthRequest(c gateway.Context) error {
	token, ok := c.Get(middleware.DefaultJWTConfig.ContextKey).(*jwt.Token)
	if !ok {
		return svc.ErrTypeAssertion
	}

	rawClaims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return svc.ErrTypeAssertion
	}

	switch claims := (*rawClaims)["claims"]; claims {
	case "user":
		var claims models.UserAuthClaims

		if err := DecodeMap(rawClaims, &claims); err != nil {
			return err
		}

		args := c.QueryParam("args")
		if args != "skip" && claims.Tenant != "" {
			// This forces any no cached token to be invalid, even if it not not expired.
			if ok, err := h.service.AuthIsCacheToken(c.Ctx(), claims.Tenant, claims.ID); err != nil || !ok {
				return svc.NewErrAuthUnathorized(err)
			}
		}

		// Extract tenant and username from JWT
		c.Response().Header().Set("X-Tenant-ID", claims.Tenant)
		c.Response().Header().Set("X-Username", claims.Username)
		c.Response().Header().Set("X-ID", claims.ID)
		c.Response().Header().Set("X-Role", claims.Role)

		return c.NoContent(http.StatusOK)
	case "device":
		var claims models.DeviceAuthClaims

		if err := DecodeMap(rawClaims, &claims); err != nil {
			return err
		}

		// Extract device UID from JWT
		c.Response().Header().Set(client.DeviceUIDHeader, claims.UID)

		return c.NoContent(http.StatusOK)
	}

	return svc.NewErrAuthUnathorized(nil)
}

func (h *Handler) AuthDevice(c gateway.Context) error {
	var req request.DeviceAuth
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

	err = h.service.SetDevicePosition(c.Ctx(), models.UID(res.UID), ip)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) AuthUser(c gateway.Context) error {
	var req request.UserAuth
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.AuthUser(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) AuthUserInfo(c gateway.Context) error {
	username := c.Request().Header.Get("X-Username")
	tenant := c.Request().Header.Get("X-Tenant-ID")
	token := c.Request().Header.Get(echo.HeaderAuthorization)

	res, err := h.service.AuthUserInfo(c.Ctx(), username, tenant, token)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) AuthGetToken(c gateway.Context) error {
	var req request.AuthTokenGet
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.AuthGetToken(c.Ctx(), req.Tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) AuthSwapToken(c gateway.Context) error {
	var req request.AuthTokenSwap
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	var id string
	if v := c.ID(); v != nil {
		id = v.ID
	}

	res, err := h.service.AuthSwapToken(c.Ctx(), id, req.Tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) AuthPublicKey(c gateway.Context) error {
	var req request.PublicKeyAuth
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

		jwt := middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:        &jwt.MapClaims{},
			SigningKey:    ctx.Service().(svc.Service).PublicKey(),
			SigningMethod: "RS256",
		})

		return jwt(next)(c)
	}
}

func DecodeMap(input, output interface{}) error {
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
