package models

type Stats struct {
	RegisteredDevices int `json:"registered_devices"`
	OnlineDevices     int `json:"online_devices"`
	ActiveSessions    int `json:"active_sessions"`
	PendingDevices    int `json:"pending_devices"`
}
