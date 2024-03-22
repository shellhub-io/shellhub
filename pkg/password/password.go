package password

type Password interface {
	Hash(pwd string) (string, error)      // Hash takes a string and returns its hash.
	Compare(pwd string, hash string) bool // Compare takes a plaintext password and a hashed password, then checks if they match.
}

var Backend Password

func init() {
	Backend = &backend{}
}

// Hash takes a string and returns its hash.
func Hash(pwd string) (string, error) {
	return Backend.Hash(pwd)
}

// Compare reports whether a plain text matches with hash.
//
// For compatibility purposes, it can compare using both SHA256 and bcrypt algorithms.
// Hashes starting with "$" are assumed to be a bcrypt hash; otherwise, they are treated as
// SHA256 hashes.
func Compare(pwd string, hash string) bool {
	return Backend.Compare(pwd, hash)
}
