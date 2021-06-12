package user

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"gopkg.in/go-playground/validator.v9"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflict")
	ErrBadRequest   = errors.New("bad request")
)

type Service interface {
	UpdateDataUser(ctx context.Context, data *models.User, id string) ([]InvalidField, error)
	UpdatePasswordUser(ctx context.Context, currentPassword, newPassword, id string) error
}

type service struct {
	store store.Store
}

type InvalidField struct {
	Name  string
	Kind  string
	Param string
	Extra string
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) UpdateDataUser(ctx context.Context, data *models.User, id string) ([]InvalidField, error) {
	var invalidFields []InvalidField

	if _, _, err := s.store.UserGetByID(ctx, id, false); err != nil {
		return invalidFields, err
	}

	if err := validator.New().Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			invalidFields = append(invalidFields, InvalidField{strings.ToLower(err.StructField()), "invalid", err.Tag(), err.Param()})
		}

		return invalidFields, ErrBadRequest
	}

	var checkUsername, checkEmail bool

	if user, err := s.store.UserGetByUsername(ctx, data.Username); err == nil && user.ID != id {
		checkUsername = true
		invalidFields = append(invalidFields, InvalidField{"username", "conflict", "", ""})
	}

	if user, err := s.store.UserGetByEmail(ctx, data.Email); err == nil && user.ID != id {
		checkEmail = true
		invalidFields = append(invalidFields, InvalidField{"email", "conflict", "", ""})
	}

	if checkUsername || checkEmail {
		return invalidFields, ErrConflict
	}

	return invalidFields, s.store.UserUpdateData(ctx, data, id)
}

func (s *service) UpdatePasswordUser(ctx context.Context, currentPassword, newPassword, id string) error {
	user, _, err := s.store.UserGetByID(ctx, id, false)
	if err != nil {
		return err
	}

	if user.Password == currentPassword {
		return s.store.UserUpdatePassword(ctx, newPassword, id)
	}

	return ErrUnauthorized
}
