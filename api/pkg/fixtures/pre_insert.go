package fixtures

import "github.com/shellhub-io/mongotest"

func setupPreInsertFuncs() []mongotest.PreInsertFunc {
	fns := make([]mongotest.PreInsertFunc, 0)

	fns = append(fns, preInsertUsers()...)
	fns = append(fns, preInsertRecoveryTokens()...)
	fns = append(fns, preInsertAnnouncements()...)
	fns = append(fns, preInsertPublicKeys()...)
	fns = append(fns, preInsertPrivateKeys()...)
	fns = append(fns, preInsertLicenses()...)
	fns = append(fns, preInsertNamespaces()...)
	fns = append(fns, preInsertDevices()...)
	fns = append(fns, preInsertConnectedDevices()...)
	fns = append(fns, preInsertFirewallRules()...)
	fns = append(fns, preInsertSessions()...)
	fns = append(fns, preInsertActiveSessions()...)
	fns = append(fns, preInsertRecordedSessions()...)

	return fns
}

func preInsertUsers() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("users", "_id"),
		mongotest.SimpleConvertTime("users", "created_at"),
		mongotest.SimpleConvertTime("users", "last_login"),
	}
}

func preInsertRecoveryTokens() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("recovery_tokens", "_id"),
		mongotest.SimpleConvertTime("recovery_tokens", "created_at"),
	}
}

func preInsertAnnouncements() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("announcements", "_id"),
		mongotest.SimpleConvertTime("announcements", "date"),
	}
}

func preInsertPublicKeys() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("public_keys", "_id"),
		mongotest.SimpleConvertBytes("public_keys", "data"),
		mongotest.SimpleConvertTime("public_keys", "created_at"),
	}
}

func preInsertPrivateKeys() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("private_keys", "_id"),
		mongotest.SimpleConvertBytes("private_keys", "data"),
		mongotest.SimpleConvertTime("private_keys", "created_at"),
	}
}

func preInsertLicenses() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("licenses", "_id"),
		mongotest.SimpleConvertBytes("licenses", "rawdata"),
		mongotest.SimpleConvertTime("licenses", "created_at"),
	}
}

func preInsertNamespaces() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("namespaces", "_id"),
		mongotest.SimpleConvertTime("namespaces", "created_at"),
	}
}

func preInsertDevices() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("devices", "_id"),
		mongotest.SimpleConvertTime("devices", "created_at"),
		mongotest.SimpleConvertTime("devices", "last_seen"),
		mongotest.SimpleConvertTime("devices", "status_updated_at"),
	}
}

func preInsertConnectedDevices() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertTime("connected_devices", "last_seen"),
	}
}

func preInsertFirewallRules() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("firewall_rules", "_id"),
	}
}

func preInsertSessions() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("sessions", "_id"),
		mongotest.SimpleConvertTime("sessions", "started_at"),
		mongotest.SimpleConvertTime("sessions", "last_seen"),
	}
}

func preInsertActiveSessions() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("active_sessions", "_id"),
		mongotest.SimpleConvertTime("active_sessions", "last_seen"),
	}
}

func preInsertRecordedSessions() []mongotest.PreInsertFunc {
	return []mongotest.PreInsertFunc{
		mongotest.SimpleConvertObjID("recorded_sessions", "_id"),
		mongotest.SimpleConvertTime("recorded_sessions", "time"),
	}
}
