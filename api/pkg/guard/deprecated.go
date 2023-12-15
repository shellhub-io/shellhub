package guard

import "github.com/shellhub-io/shellhub/pkg/models"

// Deprecated, use namespace.FindMember() instead.
//
// CheckMember checks if a models.User's ID is a models.Namespace's member. A models.User is a member if its ID is in
// the models.Namespace's members list.
func CheckMember(namespace *models.Namespace, id string) (*models.Member, bool) {
	return namespace.FindMember(id)
}
