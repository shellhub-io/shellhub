package requests

type RegisterUser struct {
	Name     string `json:"name" validate:"required,name"`
	Username string `json:"username" validate:"required,username"`
	// Email is required for open self-registration, but omitted on the invite flow:
	// there the email is derived from the invitation the Sig resolves to (the invitee
	// can't retarget it), so it's optional when a Sig is present.
	Email          string `json:"email" validate:"required_without=Sig,omitempty,email"`
	Password       string `json:"password" validate:"required,password"`
	EmailMarketing bool   `json:"email_marketing"`

	// Sig is the invitation code from the accept-invite link. When present, it proves
	// the email was already validated (the invitee clicked the link) and identifies the
	// invitation. It's a human-readable pairing code, not a UUID.
	Sig string `json:"sig" validate:"omitempty"`
}
