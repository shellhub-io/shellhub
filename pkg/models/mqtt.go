package models

const (
	MqttClientConnectedEventType    = "client_connected"
	MqttClientDisconnectedEventType = "client_disconnected"
)

type MqttEvent struct {
	Action string `json:"action"`

	MqttClientEvent
}

type MqttClientEvent struct {
	ClientID string `json:"client_id"`
	Username string `json:"username"`
}

type MqttAuthQuery struct {
	Username string `query:"username"`
	Password string `query:"password"`
	IPAddr   string `query:"ipaddr"`
}

type MqttACLQuery struct {
	Access   string `query:"access"`
	Username string `query:"username"`
	Topic    string `query:"topic"`
	IPAddr   string `query:"ipaddr"`
}
