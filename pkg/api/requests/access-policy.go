package requests

// AccessPolicyFilter selects the devices an access policy applies to. It is
// either a hostname regexp or a set of tags, never both, mirroring the
// public-key filter shape.
type AccessPolicyFilter struct {
	Hostname string   `json:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`
	Tags     []string `json:"tags,omitempty" validate:"required_without=Hostname"`
}

// AccessPolicySubject identifies who an access policy grants access to.
type AccessPolicySubject struct {
	Type  string `json:"type" validate:"required,oneof=user role all-members"`
	Value string `json:"value"`
}

// AccessPolicyIDParam represents an access policy id as a path param.
type AccessPolicyIDParam struct {
	ID string `param:"id" validate:"required"`
}

// AccessPolicyList is the structure to represent the request data for the list access policies endpoint.
type AccessPolicyList struct {
	TenantID string `json:"-"`
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
	Filter        AccessPolicyFilter  `json:"filter" validate:"required"`
	Logins        []string            `json:"logins" validate:"required,min=1,dive,required"`
	SourceIP      []string            `json:"source_ip" validate:"omitempty,dive,cidr|ip"`
	Effect        string              `json:"effect" validate:"omitempty,oneof=allow deny"`
	RequireStepUp bool                `json:"require_step_up" validate:""`
	TenantID      string              `json:"-"`
}

// AccessPolicyUpdate is the structure to represent the request data for the update access policy endpoint.
type AccessPolicyUpdate struct {
	AccessPolicyIDParam
	Name          string              `json:"name" validate:"required"`
	Subject       AccessPolicySubject `json:"subject" validate:"required"`
	Filter        AccessPolicyFilter  `json:"filter" validate:"required"`
	Logins        []string            `json:"logins" validate:"required,min=1,dive,required"`
	SourceIP      []string            `json:"source_ip" validate:"omitempty,dive,cidr|ip"`
	Effect        string              `json:"effect" validate:"omitempty,oneof=allow deny"`
	RequireStepUp bool                `json:"require_step_up" validate:""`
	TenantID      string              `json:"-"`
}

// AccessPolicyDelete is the structure to represent the request data for the delete access policy endpoint.
type AccessPolicyDelete struct {
	AccessPolicyIDParam
	TenantID string `json:"-"`
}
