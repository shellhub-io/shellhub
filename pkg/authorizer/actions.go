package authorizer

// actions is a struct to act like an Enum and facilitate to indicate the action used in the service.
type actions struct {
	Device    deviceActions
	Session   sessionActions
	Firewall  firewallActions
	PublicKey publicKeyActions
	Namespace namespaceActions
}

type deviceActions struct {
	Accept, Reject, Remove, Connect, Rename int
}

type sessionActions struct {
	Play, Close, Remove, Details int
}

type firewallActions struct {
	Create, Edit, Remove int
}

type publicKeyActions struct {
	Create, Edit, Remove int
}

type namespaceActions struct {
	Create, Rename, AddMember, RemoveMember, EditMember, EnableSessionRecord, Delete int
}

// Actions has all available and allowed actions.
// You should use it to get the code's action.
var Actions = actions{
	Device: deviceActions{
		Accept:  DeviceAccept,
		Reject:  DeviceReject,
		Remove:  DeviceRemove,
		Connect: DeviceConnect,
		Rename:  DeviceRename,
	},
	Session: sessionActions{
		Play:    SessionPlay,
		Close:   SessionClose,
		Remove:  SessionRemove,
		Details: SessionDetails,
	},
	Firewall: firewallActions{
		Create: FirewallCreate,
		Edit:   FirewallEdit,
		Remove: FirewallRemove,
	},
	PublicKey: publicKeyActions{
		Create: PublicKeyCreate,
		Edit:   PublicKeyEdit,
		Remove: PublicKeyRemove,
	},
	Namespace: namespaceActions{
		Create:              NamespaceCreate,
		Rename:              NamespaceRename,
		AddMember:           NamespaceAddMember,
		RemoveMember:        NamespaceRemoveMember,
		EditMember:          NamespaceEditMember,
		EnableSessionRecord: NamespaceEnableSessionRecord,
		Delete:              NamespaceDelete,
	},
}
