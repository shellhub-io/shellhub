package models

type ConnectorTLS struct {
	// CA is the Certificate Authority used to generate the [Cert] for the server and the client.
	CA string `json:"ca" bson:"ca" validate:"required,certPEM"`
	// Cert is generated from [CA] certificate and used by the client to authorize the connection to the Docker Engine.
	Cert string `json:"cert" bson:"cert" validate:"required,certPEM"`
	// Key is the private key for the certificate on [Cert] field.
	Key string `json:"key" bson:"key" validate:"required,privateKeyPEM"`
}

type Connector struct {
	// UID is the unique identifier of Connector.
	UID string `json:"uid" bson:"uid"`
	// TenantID indicate which namespace this connector is related.
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	// Status shows the connection status for the connector.
	Status string `json:"status" bson:"-"`
	// Enable indicates if the Connection's connection is enable.
	Enable bool `json:"enable" bson:"enable"`
	// Secure indicates if the Connector use HTTPS for authentication.
	Secure bool `json:"secure" bson:"secure"`
	// Address is the address with the port for the Docker Engine.
	Address string `json:"address" bson:"address" validate:"required,hostname_port"`
	// TLS stores the configuration for authenticate using TLS on the Docker Engine.
	TLS *ConnectorTLS `json:"tls,omitempty" bson:"tls,omitempty"`
}

type ConnectorChanges struct {
	// Enable indicates if the Connection's connection is enable.
	Enable *bool `json:"enable"`
	// Secure indicates if the Connector use HTTPS for authentication.
	Secure *bool `json:"secure"`
	// Address is the address with the port for the Docker Engine.
	Address *string `json:"address" validate:"hostname_port"`
	// TLS stores the configuration for authenticate using TLS on the Docker Engine.
	TLS *ConnectorTLS `json:"tls"`
}
