package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

type UserService interface {
	// UpdateUser updates the user's data, such as email and username. Since some attributes must be unique per user,
	// it returns a list of duplicated unique values and an error if any.
	//
	// FIX:
	// When `req.RecoveryEmail` is equal to `user.Email` or `req.Email`, return a bad request status
	// with an error object like `{"error": "recovery_email must be different from email"}` instead of setting
	// conflicts to `["email", "recovery_email"]`.
	UpdateUser(ctx context.Context, req *requests.UpdateUser) (conflicts []string, err error)

	UpdatePasswordUser(ctx context.Context, id string, currentPassword, newPassword string) error
}

func (s *service) UpdateUser(ctx context.Context, req *requests.UpdateUser) ([]string, error) {
	return nil, nil
}

// UpdatePasswordUser updates a user's password.
//
// Deprecated, use [Service.UpdateUser] instead.
func (s *service) UpdatePasswordUser(ctx context.Context, id, currentPassword, newPassword string) error {
	return nil
}
