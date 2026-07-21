package requests

// ServiceAccountList is the request data for listing a namespace's service accounts.
type ServiceAccountList struct {
	TenantID string `json:"-"`
}

// ServiceAccountCreate is the request data for creating a service account: a display
// name plus the OpenSSH public key to enroll as its first identity.
type ServiceAccountCreate struct {
	TenantID string `json:"-"`
	Name     string `json:"name" validate:"required"`
	// Data is the OpenSSH public key to enroll for the service account.
	Data string `json:"data" validate:"required"`
}

// ServiceAccountIDParam represents a service account id as a path param.
type ServiceAccountIDParam struct {
	ID string `param:"id" validate:"required"`
}

// ServiceAccountDelete is the request data for deleting a service account. Removing
// the account cascades to its membership and every SSH identity it holds.
type ServiceAccountDelete struct {
	ServiceAccountIDParam
	TenantID string `json:"-"`
}
