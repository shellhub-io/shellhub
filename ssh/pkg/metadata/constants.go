// Package metadata provides a way to store and retrieve data from a context.
package metadata

const (
	// authentication is the key to store and restore the authentication method.
	authentication = "authentication"
	// Password is the key to store and restore the password from the context.
	password = "password"
	// fingerprint is the key to store and restore the public key from the context.
	fingerprint = "public_key"
	// api is the key to store and restore an instance of internal api client.
	api = "api"
	// tag is the key to store and restore the tag from the context.
	//
	// tag is the device name or the sshid.
	tag = "tag"
	// lookup is the key to store and restore the lookup from the context.
	lookup = "lookup"
	// request is the key to store and restore the request type from the context.
	request = "request_type"
	// device is the key to store and restore the device from the context.
	device = "device"
	// sshid is the key to store and restore the sshid from the context.
	sshid = "sshid"
)

const (
	// PasswordAuthenticationMethod represents the password authentication method.
	PasswordAuthenticationMethod = iota + 1
	// PublicKeyAuthenticationMethod represents the public key authentication method.
	PublicKeyAuthenticationMethod
)
