package routes

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mitchellh/mapstructure"
	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/authsvc"
	api "github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	AuthRequestURL   = "/auth"
	AuthDeviceURL    = "/devices/auth"
	AuthDeviceURLV2  = "/auth/device"
	AuthUserURL      = "/login"
	AuthUserURLV2    = "/auth/user"
	AuthUserTokenURL = "/auth/token/:tenant"
	AuthPublicKeyURL = "/auth/ssh"
)

func AuthRequest(c apicontext.Context) error {
	token := c.Get("user").(*jwt.Token)
	rawClaims := token.Claims.(*jwt.MapClaims)

	switch claims := (*rawClaims)["claims"]; claims {
	case "user":
		var claims models.UserAuthClaims

		if err := DecodeMap(rawClaims, &claims); err != nil {
			return err
		}

		// Extract tenant and username from JWT
		c.Response().Header().Set("X-Tenant-ID", claims.Tenant)
		c.Response().Header().Set("X-Username", claims.Username)

		return nil
	case "device":
		var claims models.DeviceAuthClaims

		if err := DecodeMap(rawClaims, &claims); err != nil {
			return err
		}

		// Extract device UID from JWT
		c.Response().Header().Set(api.DeviceUIDHeader, claims.UID)

		return nil
	}

	return echo.ErrUnauthorized
}

func AuthDevice(c apicontext.Context) error {
	var req models.DeviceAuthRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	svc := authsvc.NewService(c.Store(), nil, nil)

	res, err := svc.AuthDevice(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func AuthUser(c apicontext.Context) error {
	var req models.UserAuthRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	svc := authsvc.NewService(c.Store(), nil, nil)

	res, err := svc.AuthUser(c.Ctx(), req)
	if err != nil {
		return echo.ErrUnauthorized
	}

	return c.JSON(http.StatusOK, res)
}

func AuthUserInfo(c apicontext.Context) error {
	username := c.Request().Header.Get("X-Username")

	user, err := c.Store().GetUserByUsername(c.Ctx(), username)
	if err != nil {
		return echo.ErrUnauthorized
	}

	return c.JSON(http.StatusOK, &models.UserAuthResponse{
		Token:  c.Request().Header.Get(echo.HeaderAuthorization),
		Name:   user.Name,
		User:   user.Username,
		Tenant: user.TenantID,
	})
}

func AuthGetToken(c apicontext.Context) error {
	svc := authsvc.NewService(c.Store(), nil, nil)
	res, err := svc.AuthGetToken(c.Ctx(), c.Param("tenant"))
	if err != nil {
		return echo.ErrUnauthorized
	}
	return c.JSON(http.StatusOK, res)
}

func AuthPublicKey(c apicontext.Context) error {
	var req models.PublicKeyAuthRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	svc := authsvc.NewService(c.Store(), nil, nil)

	res, err := svc.AuthPublicKey(c.Ctx(), &req)
	if err != nil {
		return echo.ErrUnauthorized
	}

	return c.JSON(http.StatusOK, res)
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Get("ctx").(*apicontext.Context)
		svc := authsvc.NewService(ctx.Store(), nil, nil)

		jwt := middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:        &jwt.MapClaims{},
			SigningKey:    svc.PublicKey(),
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
