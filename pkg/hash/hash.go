package hash

// Hasher hashes passwords and checks them. CompareWith must compare in constant time, so a wrong
// password cannot be found byte by byte from timing.
type Hasher interface {
	Do(plain string) (string, error)            // Hash takes a string and returns its hash.
	CompareWith(plain string, hash string) bool // Compare takes a plaintext password and a hashed password, then checks if they match.
}

// Backend is the implementation the package-level helpers use. Tests swap it with Set, which puts
// the previous one back when the test ends; production leaves it alone.
var Backend Hasher = &backend{}

// Set serves Do and CompareWith from h until t and its subtests finish, then puts back the hasher
// it replaced. Calls nest: an inner Set gives back the outer one when its test ends. The hasher is
// process-global, so a test that calls Set must not run in parallel with one that hashes.
func Set(t interface{ Cleanup(func()) }, h Hasher) {
	previous := Backend
	t.Cleanup(func() { Backend = previous })
	Backend = h
}

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
