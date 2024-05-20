package hash

type Hasher interface {
	Do(plain string) (string, error)            // Hash takes a string and returns its hash.
	CompareWith(plain string, hash string) bool // Compare takes a plaintext password and a hashed password, then checks if they match.
}

var Backend Hasher = &backend{}

// Do takes a string and returns its hash.
func Do(pwd string) (string, error) {
	return Backend.Do(pwd)
}

// CompareWith reports whether a plain text matches with hash.
//
// For compatibility purposes, it can compare using both SHA256 and bcrypt algorithms.
// Hashes starting with "$" are assumed to be a bcrypt hash; otherwise, they are treated as
// SHA256 hashes.
func CompareWith(plain string, hash string) bool {
	return Backend.CompareWith(plain, hash)
}
