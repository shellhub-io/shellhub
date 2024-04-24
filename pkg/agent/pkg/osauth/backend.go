package osauth

import (
	"os"
)

type backend struct{}

func (b *backend) AuthUser(username, password string) bool {
	file, err := os.Open(DefaultShadowFilename)
	if err != nil {
		return false
	}

	return AuthUserFromShadow(username, password, file)
}

func (b *backend) LookupUser(username string) (*User, error) {
	file, err := os.Open(DefaultPasswdFilename)
	if err != nil {
		return nil, err
	}

	return LookupUserFromPasswd(username, file)
}
