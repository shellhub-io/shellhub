package models

import "time"

// ServiceAccount is a non-human principal for automated systems (CI, backups,
// config management): a service-typed user (see [UserTypeService]) plus a namespace
// membership that holds one or more SSH identities. It never signs in to the console
// and is not an API principal, existing only for the SSH identity scheme. It is
// authorized by the same Access Policies as human members; the human/service
// distinction lives in the user's type, not in the policy.
type ServiceAccount struct {
	// ID is the underlying service user's id.
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	// Identities are the SSH keys enrolled for this service account in the namespace.
	Identities []SSHIdentity `json:"identities"`
}
