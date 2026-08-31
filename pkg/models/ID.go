package models

// ID carries the authenticated user's id as the API's gateway reads it out of a request's
// context. It is a struct rather than a string so a missing id is a nil pointer, not "".
type ID struct {
	ID string
}
