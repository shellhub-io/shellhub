package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

func (s *service) UserCreate(username, password, email string) (*models.User, error) {
	ctx := context.Background()

	// returnDuplicatedField checks user's name and user's email already exist in the database.
	returnDuplicatedField := func(ctx context.Context, username, email string) error {
		list, _, err := s.store.UserList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil)
		if err != nil {
			return err
		}

		var errs [2]error
		for _, item := range list {
			if item.Username == username {
				errs[0] = ErrUserNameExists
			}

			if item.Email == email {
				errs[1] = ErrUserEmailExists
			}
		}

		switch errs {
		case [2]error{ErrUserNameExists, nil}:
			return ErrUserNameExists
		case [2]error{nil, ErrUserEmailExists}:
			return ErrUserEmailExists
		case [2]error{ErrUserNameExists, ErrUserEmailExists}:
			return ErrUserNameAndEmailExists
		}

		return nil
	}

	name := normalizeField(username)
	mail := normalizeField(email)

	userData := models.UserData{
		Name:     name,
		Email:    mail,
		Username: name,
	}

	_, err := validator.ValidateStruct(userData)
	if err != nil {
		return nil, ErrUserDataInvalid
	}

	if ok := validator.ValidateFieldPassword(password); !ok {
		return nil, ErrUserPasswordInvalid
	}

	userPass := models.UserPassword{
		Password: hashPassword(password),
	}

	user := &models.User{
		UserData:      userData,
		UserPassword:  userPass,
		Confirmed:     true,
		CreatedAt:     clock.Now(),
		MaxNamespaces: -1,
	}

	err = s.store.UserCreate(ctx, user)
	if err != nil {
		if err == store.ErrDuplicate {
			return nil, returnDuplicatedField(ctx, user.Username, user.Email)
		}

		return nil, ErrCreateNewUser
	}

	return user, nil
}

func (s *service) UserDelete(username string) error {
	ctx := context.Background()

	// Gets the user data.
	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}

	// Gets data about the namespaces what the user is either member or owner.
	detach, err := s.store.UserDetachInfo(ctx, user.ID)
	if err != nil {
		return ErrNamespaceNotFound
	}
	// Owned namespaces.
	owned := detach["owner"]
	// Joined namespaces.
	joined := detach["member"]

	// Delete all namespaces what the user is member.
	for _, o := range owned {
		if err := s.store.NamespaceDelete(ctx, o.TenantID); err != nil {
			return err
		}
	}

	// Remove user from all namespaces what it is a member.
	for _, m := range joined {
		if _, err := s.store.NamespaceRemoveMember(ctx, m.TenantID, user.ID); err != nil {
			return err
		}
	}

	// Delete the user.
	if err := s.store.UserDelete(ctx, user.ID); err != nil {
		return ErrFailedDeleteUser
	}

	return nil
}

func (s *service) UserUpdate(username, password string) error {
	ctx := context.Background()

	ok := validator.ValidateFieldPassword(password)
	if !ok {
		return ErrUserPasswordInvalid
	}

	passHash := hashPassword(password)

	user, err := s.store.UserGetByUsername(ctx, username)
	if err != nil {
		return ErrUserNotFound
	}

	if err := s.store.UserUpdatePassword(ctx, passHash, user.ID); err != nil {
		return ErrFailedUpdateUser
	}

	return nil
}
