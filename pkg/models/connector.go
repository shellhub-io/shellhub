package models

type ConnectorStatus struct {
	// State of connection.
	State string `json:"state" bson:"state"`
	// Message contains a message what caused the current [State].
	Message string `json:"message" bson:"message"`
}

type ConnectorTLS struct {
	// CA is the Certificate Authority used to generate the [Cert] for the server and the client.
	CA string `json:"ca" bson:"ca" validate:"required,certPEM"`
	// Cert is generated from [CA] certificate and used by the client to authorize the connection to the Container Engine.
	Cert string `json:"cert" bson:"cert" validate:"required,certPEM"`
	// Key is the private key for the certificate on [Cert] field.
	Key string `json:"key" bson:"key" validate:"required,privateKeyPEM"`
}

// ConnectorData contains the mutable data for each Connector.
type ConnectorData struct {
	// Enable indicates if the Connection's connection is enable.
	Enable *bool `json:"enable" bson:"enable,omitempty"`
	// Secure indicates if the Connector use HTTPS for authentication.
	Secure *bool `json:"secure" bson:"secure,omitempty"`
	// Address is the address to the Container Engine.
	Address *string `json:"address" bson:"address,omitempty" validate:"required,hostname_rfc1123"`
	// Port is the port to Container Engine.
	Port *uint `json:"port" bson:"port,omitempty" validate:"required,min=1,max=65535"`
	// TLS stores the configuration for authenticate using TLS on the Container Engine.
	TLS *ConnectorTLS `json:"tls,omitempty" bson:"tls,omitempty"  validate:"required_if=Secure true"`
}

type Connector struct {
	// UID is the unique identifier of Connector.
	UID string `json:"uid" bson:"uid"`
	// TenantID indicate which namespace this connector is related.
	TenantID string `json:"tenant_id" bson:"tenant_id"`
	// Status shows the connection status for the connector.
	Status        ConnectorStatus `json:"status" bson:"-"`
	ConnectorData `bson:",inline"`
}
