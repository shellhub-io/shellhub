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

	// lookup is the key to store and restore the lookup from the context.
	lookup = "lookup"

	// request is the key to store and restore the request type from the context.
	request = "request_type"

	// device is the key to store and restore the device from the context.
	device = "device"

	// sshid is the key to store and restore the sshid from the context.
	sshid = "sshid"

	// agent is the key to store and restore the agent from the context.
	agent = "agent"

	// established is the key to store and restore the established state from the context.
	established = "established"

	// tag is the key to store and restore the tag from the context.
	//
	// tag is the device name or the sshid.
	tag = "tag"
)

type AuthenticationMethod int

const (
	// InvalidAuthenticationMethod represents an invalid InvalidAuthenticationMethod
	InvalidAuthenticationMethod AuthenticationMethod = iota

	// PasswordAuthenticationMethod represents the password authentication method.
	PasswordAuthenticationMethod

	// PublicKeyAuthenticationMethod represents the public key authentication method.
	PublicKeyAuthenticationMethod
)
