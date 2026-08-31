package models

// UID identifies a device or a session — the opaque, externally visible identifier a client
// addresses a resource by, not the store's primary key.
type UID string

// AuthClaims carries the kind of claims a JWT holds ("user", "device"), which decides how the
// rest of the token is read.
type AuthClaims struct {
	Claims string `json:"claims"`
}
