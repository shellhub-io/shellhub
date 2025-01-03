package models

type SystemInfo struct {
	Endpoints *SystemInfoEndpoints `json:"endpoints"`
	Version   string               `json:"version"`
	Setup     bool                 `json:"setup"`
}

type SystemInfoEndpoints struct {
	API string `json:"api"`
	SSH string `json:"ssh"`
}

type System struct {
	Setup bool `json:"setup" bson:"setup"`
}
