package responses

// SystemInfo is what an unauthenticated client learns about the instance: its version, where
// to reach it, whether it still needs setting up, and how to authenticate.
type SystemInfo struct {
	Version        string                    `json:"version"`
	Endpoints      *SystemEndpointsInfo      `json:"endpoints"`
	Setup          bool                      `json:"setup"`
	Authentication *SystemAuthenticationInfo `json:"authentication"`
}

// SystemAuthenticationInfo tells a client which login methods the instance offers, so the UI
// can render only those.
type SystemAuthenticationInfo struct {
	Local bool `json:"local"`
}

// SystemEndpointsInfo is where the client should reach the API and SSH, which may differ from
// the address it asked, behind a proxy.
type SystemEndpointsInfo struct {
	API string `json:"api"`
	SSH string `json:"ssh"`
}
