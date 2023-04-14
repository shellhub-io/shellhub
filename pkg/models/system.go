package models

type SystemInfo struct {
	Version   string               `json:"version"`
	Endpoints *SystemInfoEndpoints `json:"endpoints"`
}

type SystemInfoEndpoints struct {
	API string `json:"api"`
	SSH string `json:"ssh"`
}
