package models

// APIKeyParam is the structure to represent the request data for delete a APIKey.
type APIKey struct {
	ID        string `json:"id" bson:"_id" validate:"required"`
	UserID    string `json:"user_id" bson:"user_id" validate:"required"`
	TenantID  string `json:"tenant_id" bson:"tenant_id" validate:"required"`
	Name      string `json:"name" bson:"name" validate:"required"`
	ExpiresIn int64  `json:"expires_in" bson:"expires_in" validate:"required"`
}
