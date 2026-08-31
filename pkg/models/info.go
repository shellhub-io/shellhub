package models

// Info is what an unauthenticated client reads to learn the server's version and where to reach
// it, so it is the one payload that must stay stable across upgrades.
type Info struct {
	Version   string    `json:"version"`
	Endpoints Endpoints `json:"endpoints"`
}

// Endpoints are the addresses a client connects back on, as the server sees itself from outside —
// behind a reverse proxy they come from configuration, not from the request.
type Endpoints struct {
	API string `json:"api"`
	SSH string `json:"ssh"`
}
