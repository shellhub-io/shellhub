package guard

type Action int

// AllActions is a struct to act like an Enum and facilitate to indicate the action used in the service.
type AllActions struct {
	Device    DeviceActions
	Session   SessionActions
	Firewall  FirewallActions
	PublicKey PublicKeyActions
	Namespace NamespaceActions
	Billing   BillingActions
}

type DeviceActions struct {
	Accept, Reject, Update, Remove, Connect, Rename, CreateTag, UpdateTag, RemoveTag, RenameTag, DeleteTag int
}

type SessionActions struct {
	Play, Close, Remove, Details int
}

type FirewallActions struct {
	Create, Edit, Remove, AddTag, UpdateTag, RemoveTag int
}

type PublicKeyActions struct {
	Create, Edit, Remove, AddTag, RemoveTag, UpdateTag int
}

type NamespaceActions struct {
	Rename, AddMember, RemoveMember, EditMember, EnableSessionRecord, Delete int
}

type BillingActions struct {
	CreateCustomer, ChooseDevices, AddPaymentMethod, UpdatePaymentMethod, RemovePaymentMethod, CancelSubscription, CreateSubscription, GetSubscription int
}

// Actions has all available and allowed actions.
// You should use it to get the code's action.
var Actions = AllActions{
	Device: DeviceActions{
		Accept:    DeviceAccept,
		Reject:    DeviceReject,
		Update:    DeviceUpdate,
		Remove:    DeviceRemove,
		Connect:   DeviceConnect,
		Rename:    DeviceRename,
		CreateTag: DeviceCreateTag,
		UpdateTag: DeviceUpdateTag,
		RemoveTag: DeviceRemoveTag,
		RenameTag: DeviceRenameTag,
		DeleteTag: DeviceDeleteTag,
	},
	Session: SessionActions{
		Play:    SessionPlay,
		Close:   SessionClose,
		Remove:  SessionRemove,
		Details: SessionDetails,
	},
	Firewall: FirewallActions{
		Create: FirewallCreate,
		Edit:   FirewallEdit,
		Remove: FirewallRemove,
	},
	PublicKey: PublicKeyActions{
		Create:    PublicKeyCreate,
		Edit:      PublicKeyEdit,
		Remove:    PublicKeyRemove,
		AddTag:    PublicKeyAddTag,
		RemoveTag: PublicKeyRemoveTag,
		UpdateTag: PublicKeyUpdateTag,
	},
	Namespace: NamespaceActions{
		Rename:              NamespaceRename,
		AddMember:           NamespaceAddMember,
		RemoveMember:        NamespaceRemoveMember,
		EditMember:          NamespaceEditMember,
		EnableSessionRecord: NamespaceEnableSessionRecord,
		Delete:              NamespaceDelete,
	},
	Billing: BillingActions{
		CreateCustomer:      BillingCreateCustomer,
		ChooseDevices:       BillingChooseDevices,
		AddPaymentMethod:    BillingAddPaymentMethod,
		UpdatePaymentMethod: BillingUpdatePaymentMethod,
		RemovePaymentMethod: BillingRemovePaymentMethod,
		CancelSubscription:  BillingCancelSubscription,
		CreateSubscription:  BillingCreateSubscription,
		GetSubscription:     BillingGetSubscription,
	},
}
