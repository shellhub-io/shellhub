package models

// Stats is the dashboard's counter set for one namespace. The device counts partition by status,
// so registered, pending and rejected do not overlap and do not sum to the namespace's limit.
type Stats struct {
	RegisteredDevices int64 `json:"registered_devices"`
	OnlineDevices     int64 `json:"online_devices"`
	ActiveSessions    int64 `json:"active_sessions"`
	PendingDevices    int64 `json:"pending_devices"`
	RejectedDevices   int64 `json:"rejected_devices"`
}
