package response

type PublicKeyFilter struct {
	Hostname string `json:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`
	// FIXME: add validation for tags when it has at least one item.
	//
	// If used `min=1` to do that validation, when tags is empty, its zero value, and only hostname is provided,
	// it throws a error even with `required_without` and `excluded_with`.
	Tags []string `json:"tags,omitempty" validate:"required_without=Hostname,excluded_with=Hostname,max=3,unique,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// PublicKeyCreate is the structure to represent the request data for create public key endpoint.
type PublicKeyCreate struct {
	Data        []byte          `json:"data"`
	Filter      PublicKeyFilter `json:"filter"`
	Name        string          `json:"name"`
	Username    string          `json:"username"`
	TenantID    string          `json:"tenant_id"`
	Fingerprint string          `json:"fingerprint"`
}
