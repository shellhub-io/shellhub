package routes

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	ParamPublicKeyFingerprint = "fingerprint"
)

func (h *Handler) createPublicKey() *Route {
	return &Route{
		endpoint:              "/sshkeys/public-keys",
		method:                MethodPost,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.PublicKeyCreate
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
				req.TenantID = tenant
			}

			var res *responses.PublicKeyCreate
			err := guard.EvaluatePermission(c.Role(), guard.Actions.PublicKey.Create, func() error {
				var err error
				res, err = h.service.CreatePublicKey(c.Ctx(), req, tenant)

				return err
			})
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, res)
		},
	}
}

func (h *Handler) evaluatePublicKey() *Route {
	return &Route{
		endpoint:              "/public-keys/evaluate/:fingerprint/:username",
		method:                MethodPost,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var device models.Device
			if err := c.Bind(&device); err != nil {
				return c.JSON(http.StatusForbidden, err)
			}

			pubKey, err := h.service.GetPublicKey(c.Ctx(), c.Param(ParamPublicKeyFingerprint), device.TenantID)
			if err != nil {
				return c.JSON(http.StatusForbidden, err)
			}

			usernameOk, err := h.service.EvaluateKeyUsername(c.Ctx(), pubKey, c.Param(ParamUserName))
			if err != nil {
				return err
			}

			filterOk, err := h.service.EvaluateKeyFilter(c.Ctx(), pubKey, device)
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, usernameOk && filterOk)
		},
	}
}

func (h *Handler) getPublicKey() *Route {
	return &Route{
		endpoint:              "/sshkeys/public-keys/:fingerprint/:tenant",
		method:                MethodGet,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.PublicKeyGet
			err := c.Bind(&req)
			if err != nil {
				return err
			}

			err = c.Validate(&req)
			if err != nil {
				return err
			}

			pubKey, err := h.service.GetPublicKey(c.Ctx(), req.Fingerprint, req.Tenant)
			if err != nil {
				if err == store.ErrNoDocuments {
					return c.NoContent(http.StatusNotFound)
				}

				return err
			}

			return c.JSON(http.StatusOK, pubKey)
		},
	}
}

func (h *Handler) listPublicKeys() *Route {
	return &Route{
		endpoint:              "/sshkeys/public-keys",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			paginator := query.NewPaginator()
			if err := c.Bind(paginator); err != nil {
				return err
			}

			// TODO: normalize is not required when request is privileged
			paginator.Normalize()

			list, count, err := h.service.ListPublicKeys(c.Ctx(), *paginator)
			if err != nil {
				return err
			}

			c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

			return c.JSON(http.StatusOK, list)
		},
	}
}

func (h *Handler) updatePublicKey() *Route {
	return &Route{
		endpoint:              "/sshkeys/public-keys/:fingerprint",
		method:                MethodPut,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.PublicKeyUpdate
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			var key *models.PublicKey
			err := guard.EvaluatePermission(c.Role(), guard.Actions.PublicKey.Edit, func() error {
				var err error
				key, err = h.service.UpdatePublicKey(c.Ctx(), req.Fingerprint, tenant, req)

				return err
			})
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, key)
		},
	}
}

func (h *Handler) deletePublicKey() *Route {
	return &Route{
		endpoint:              "/sshkeys/public-keys/:fingerprint",
		method:                MethodDelete,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.PublicKeyDelete
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.PublicKey.Remove, func() error {
				err := h.service.DeletePublicKey(c.Ctx(), req.Fingerprint, tenant)

				return err
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) addPublicKeyTag() *Route {
	return &Route{
		endpoint:              "/sshkeys/public-keys/:fingerprint/tags",
		method:                MethodPost,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.PublicKeyTagAdd
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.PublicKey.AddTag, func() error {
				return h.service.AddPublicKeyTag(c.Ctx(), tenant, req.Fingerprint, req.Tag)
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) updatePublicKeyTags() *Route {
	return &Route{
		endpoint:              "/sshkeys/public-keys/:fingerprint/tags",
		method:                MethodPut,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.PublicKeyTagsUpdate
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.PublicKey.UpdateTag, func() error {
				return h.service.UpdatePublicKeyTags(c.Ctx(), tenant, req.Fingerprint, req.Tags)
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) removePublicKeyTag() *Route {
	return &Route{
		endpoint:              "/sshkeys/public-keys/:fingerprint/tags/:tag",
		method:                MethodDelete,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.PublicKeyTagRemove
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.PublicKey.RemoveTag, func() error {
				return h.service.RemovePublicKeyTag(c.Ctx(), tenant, req.Fingerprint, req.Tag)
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) createPrivateKey() *Route {
	return &Route{
		endpoint:              "/sshkeys/private-keys",
		method:                MethodPost,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			privKey, err := h.service.CreatePrivateKey(c.Ctx())
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, privKey)
		},
	}
}
