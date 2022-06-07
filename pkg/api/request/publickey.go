package request

// FingerprintParam is a parameter that can be used to validate a fingerprint .
type FingerprintParam struct {
	// Fingerprint is the public key's fingerprint.
	Fingerprint string `param:"fingerprint" validate:"required"`
}

// PublicKeyGet is the structure for the request data at get public key endpoint.
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

// PublicKeyCreate is the structure for the request data at create public key endpoint.
type PublicKeyCreate struct {
	Data        []byte          `json:"data" validate:"required"`
	Filter      PublicKeyFilter `json:"filter" validate:"required"`
	Name        string          `json:"name" validate:"required"`
	Username    string          `json:"username" validate:"required,regexp"`
	TenantID    string          `json:"-"`
	Fingerprint string          `json:"-"`
}

// PublicKeyUpdate is the structure for the request data at update public key endpoint.
type PublicKeyUpdate struct {
	FingerprintParam
	// Name is the public key's name.
	Name string `json:"name" validate:"required"`
	// Username is the public key's username.
	Username string `json:"username" validate:"required,regexp"`
	// Filter is the public key's filter.
	Filter PublicKeyFilter `json:"filter" validate:"required"`
}

// PublicKeyDelete is the structure for the request data at delete public key endpoint.
type PublicKeyDelete struct {
	FingerprintParam
}

// PublicKeyTagAdd is the structure for the request data at add tag to public key endpoint.
type PublicKeyTagAdd struct {
	FingerprintParam
	TagParam
}

// PublicKeyTagRemove is the structure for the request data at remove tag from public key endpoint.
type PublicKeyTagRemove struct {
	FingerprintParam
	TagParam
}

// PublicKeyTagsUpdate is the structure for the request data at update tags from public key endpoint.
type PublicKeyTagsUpdate struct {
	FingerprintParam
	Tags []string `json:"tags" validate:"required,min=1,max=3,unique,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}
