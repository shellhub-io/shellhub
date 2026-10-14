package requests

// AccessPolicyFilter selects the devices an access policy applies to: every device when All is
// set, or the devices carrying any of Tags. More than three tags fails validation; the service
// refuses a filter that gives neither, or both.
type AccessPolicyFilter struct {
	All  bool     `json:"all"`
	Tags []string `json:"tags" validate:"max=3"`
}

// AccessPolicySubject identifies who an access policy grants access to.
type AccessPolicySubject struct {
	Type  string `json:"type" validate:"required,oneof=user role api-key all-members"`
	Value string `json:"value"`
}

// AccessPolicyIDParam represents an access policy id as a path param.
type AccessPolicyIDParam struct {
	ID string `param:"id" validate:"required"`
}

// AccessPolicyGet is the structure to represent the request data for the get access policy endpoint.
type AccessPolicyGet struct {
	AccessPolicyIDParam
	TenantID string `json:"-"`
}

// AccessPolicyCreate is the structure to represent the request data for the create access policy endpoint.
type AccessPolicyCreate struct {
	Name          string              `json:"name" validate:"required"`
	Subject       AccessPolicySubject `json:"subject" validate:"required"`
	Filter        AccessPolicyFilter  `json:"filter"`
	Logins        []string            `json:"logins" validate:"required,min=1,dive,required"`
	SourceIP      []string            `json:"source_ip" validate:"omitempty,dive,cidr|ip"`
	Action        string              `json:"action" validate:"omitempty,oneof=allow deny"`
	RequireReauth bool                `json:"require_reauth" validate:""`
	ReauthPeriod  *int                `json:"reauth_period" validate:"omitempty,gte=0"`
	TenantID      string              `json:"-"`
}

// AccessPolicyUpdate is the structure to represent the request data for the update access policy endpoint.
type AccessPolicyUpdate struct {
	AccessPolicyIDParam
	Name          string              `json:"name" validate:"required"`
	Subject       AccessPolicySubject `json:"subject" validate:"required"`
	Filter        AccessPolicyFilter  `json:"filter"`
	Logins        []string            `json:"logins" validate:"required,min=1,dive,required"`
	SourceIP      []string            `json:"source_ip" validate:"omitempty,dive,cidr|ip"`
	Action        string              `json:"action" validate:"omitempty,oneof=allow deny"`
	RequireReauth bool                `json:"require_reauth" validate:""`
	ReauthPeriod  *int                `json:"reauth_period" validate:"omitempty,gte=0"`
	TenantID      string              `json:"-"`
}

// AccessPolicyDelete is the structure to represent the request data for the delete access policy endpoint.
type AccessPolicyDelete struct {
	AccessPolicyIDParam
	TenantID string `json:"-"`
}
