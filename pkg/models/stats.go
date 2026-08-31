package models

// Stats is the dashboard's counter set for one namespace. The device counts partition by status,
// so registered, pending and rejected do not overlap and do not sum to the namespace's limit.
type Stats struct {
	RegisteredDevices int `json:"registered_devices"`
	OnlineDevices     int `json:"online_devices"`
	ActiveSessions    int `json:"active_sessions"`
	PendingDevices    int `json:"pending_devices"`
	RejectedDevices   int `json:"rejected_devices"`
}
