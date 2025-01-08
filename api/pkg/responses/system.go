package responses

type SystemInfo struct {
	Version        string                    `json:"version"`
	Endpoints      *SystemEndpointsInfo      `json:"endpoints"`
	Setup          bool                      `json:"setup"`
	Authentication *SystemAuthenticationInfo `json:"authentication"`
}

type SystemAuthenticationInfo struct {
	Local bool `json:"local"`
	SAML  bool `json:"saml"`
}

type SystemEndpointsInfo struct {
	API string `json:"api"`
	SSH string `json:"ssh"`
}
