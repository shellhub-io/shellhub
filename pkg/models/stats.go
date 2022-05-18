package models

type Stats struct {
	// Number of registered devices.
	RegisteredDevices int `json:"registered_devices"`
	// Number of online devices.
	OnlineDevices     int `json:"online_devices"`
	// Number of active sessions.
	ActiveSessions    int `json:"active_sessions"`
	// Number of pending devices.
	PendingDevices    int `json:"pending_devices"`
	// Number of rejected devices.
	RejectedDevices   int `json:"rejected_devices"`
}
