package request

// FingerprintParam is a structure to represent and validate a public key fingerprint as path param.
type FingerprintParam struct {
	Fingerprint string `param:"fingerprint" validate:"required"`
}

// PublicKeyGet is the structure to represent the request data for get public key endpoint.
type PublicKeyGet struct {
	FingerprintParam
	TenantParam
}

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
	Data        []byte          `json:"data" validate:"required"`
	Filter      PublicKeyFilter `json:"filter" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Username    string          `json:"username" validate:"required,regexp"`
	TenantID    string          `json:"-"`
	Fingerprint string          `json:"-"`
}

// PublicKeyUpdate is the structure to represent the request data for update public key endpoint.
type PublicKeyUpdate struct {
	FingerprintParam
	// Name is the public key's name.
	Name string `json:"name" validate:"required"`
	// Username is the public key's username.
	Username string `json:"username" validate:"required,regexp"`
	// Filter is the public key's filter.
	Filter PublicKeyFilter `json:"filter" validate:"required"`
}

// PublicKeyDelete is the structure to represent the request data for delete public key endpoint.
type PublicKeyDelete struct {
	FingerprintParam
}

// PublicKeyTagAdd is the structure to represent the request data for add tag to public key endpoint.
type PublicKeyTagAdd struct {
	FingerprintParam
	TagParam
}

// PublicKeyTagRemove is the structure to represent the request data for remove tag from public key endpoint.
type PublicKeyTagRemove struct {
	FingerprintParam
	TagParam
}

// PublicKeyTagsUpdate is the structure to represent the request data for update tags from public key endpoint.
type PublicKeyTagsUpdate struct {
	FingerprintParam
	Tags []string `json:"tags" validate:"required,min=1,max=3,unique,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// PublicKeyAuth is the structure to represent the request data for public key auth endpoint.
type PublicKeyAuth struct {
	Fingerprint string `json:"fingerprint" validate:"required"`
	Data        string `json:"data" validate:"required"`
}
