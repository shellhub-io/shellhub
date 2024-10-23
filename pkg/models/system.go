package models

type SystemInfo struct {
	Version   string               `json:"version"`
	Endpoints *SystemInfoEndpoints `json:"endpoints"`
	Setup     bool                 `json:"setup"`
}

type SystemInfoEndpoints struct {
	API string `json:"api"`
	SSH string `json:"ssh"`
}

type System struct {
	Setup bool `json:"setup" bson:"setup"`
}
