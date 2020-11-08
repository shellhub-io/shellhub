package models

type Namespace struct {
	Name     string             `json:"name"  validate:"required,hostname_rfc1123"`
	Owner    string             `json:"owner"`
	TenantID string             `json:"tenant_id" bson:"tenant_id,omitempty"`
	Members  []string           `json:"members" bson:"members"`
	Settings *NamespaceSettings `json:"settings"`
}

type NamespaceSettings struct {
	SessionRecord bool `json:"session_record" bson:"session_record,omitempty"`
}
