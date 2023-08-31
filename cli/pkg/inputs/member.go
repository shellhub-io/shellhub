package inputs

// MemberAdd is a struct for handling input when adding a member.
type MemberAdd struct {
	Username  string `validate:"required,username"`
	Namespace string
	Role      string
}

// MemberRemove is a struct for handling input when removing a member.
type MemberRemove struct {
	Username  string `validate:"required,username"`
	Namespace string
}
