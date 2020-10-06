package routes

import (
	"net/http"

	"crypto/sha256"
	"encoding/hex"
	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/user"
)

const (
	UpdateUserURL   = "/user"
	UserSecurityURL = "/user/security"
)

func UpdateUser(c apicontext.Context) error {
	var req struct {
		Username        string `json:"username"`
		Email           string `json:"email"`
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json: "newPassword"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}
	if req.CurrentPassword != "" {
		sum := sha256.Sum256([]byte(req.CurrentPassword))
		sumByte := sum[:]
		req.CurrentPassword = hex.EncodeToString(sumByte)
	}
	if req.NewPassword != "" {
		sum := sha256.Sum256([]byte(req.NewPassword))
		sumByte := sum[:]
		req.NewPassword = hex.EncodeToString(sumByte)
	}

	svc := user.NewService(c.Store())

	if invalidFields, err := svc.UpdateDataUser(c.Ctx(), req.Username, req.Email, req.CurrentPassword, req.NewPassword, tenant); err != nil {
		switch {
		case err == user.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case err == user.ErrConflict:
			return c.JSON(http.StatusConflict, invalidFields)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, nil)
}

func UpdateUserSecurity(c apicontext.Context) error {
	var req struct {
		SessionRecord bool `json:"sessionRecord"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	svc := user.NewService(c.Store())

	err := svc.UpdateDataUserSecurity(c.Ctx(), req.SessionRecord, tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}

func GetUserSecurity(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	svc := user.NewService(c.Store())

	status, err := svc.GetDataUserSecurity(c.Ctx(), tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}
