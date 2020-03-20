package models

type Info struct {
	Version   string    `json:"version"`
	Endpoints Endpoints `json:"endpoints"`
}

type Endpoints struct {
	API string `json:"api"`
	SSH string `json:"ssh"`
}
