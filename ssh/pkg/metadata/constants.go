// Package metadata provides a way to store and retrieve data from an SSH context.
package metadata

const (
	authentication = "authentication" // authentication is the key to store and restore the authentication method.
	password       = "password"       // Password is the key to store and restore the password from the context.
	fingerprint    = "public_key"     // fingerprint is the key to store and restore the public key from the context.
	api            = "api"            // api is the key to store and restore an instance of internal api client.
	lookup         = "lookup"         // lookup is the key to store and restore the lookup from the context.
	request        = "request_type"   // request is the key to store and restore the request type from the context.
	device         = "device"         // device is the key to store and restore the device from the context.
	sshid          = "sshid"          // sshid is the key to store and restore the sshid from the context.
	agent          = "agent"          // agent is the key to store and restore the agent from the context.
	established    = "established"    // established is the key to store and restore the established state from the context.

	// tag is the key to store and restore the tag from the context.
	//
	// tag is the device name or the sshid.
	tag = "tag"
)

type AuthMethod int

const (
	AuthMethodInvalid AuthMethod = 0 // AuthMethodInvalid represents an invalid authentication method.
	AuthMethodPasswd  AuthMethod = 1 // PasswordAuthenticationMethod represents the password authentication method.
	AuthMethodPubKey  AuthMethod = 2 // PublicKeyAuthenticationMethod represents the public key authentication method.
)
