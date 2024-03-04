package metadata

const (
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

	// tag is the key to store and restore the tag from the context.
	//
	// tag is the device name or the sshid.
	tag = "tag"
)
