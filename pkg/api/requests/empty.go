package requests

// Empty is the request of a route that takes no input. A handler is a function of its request, so
// a route with nothing to read still names the shape of what it reads.
type Empty struct{}
