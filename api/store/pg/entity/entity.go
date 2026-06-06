package entity

func Entities() []any {
	return []any{
		// Register intermediary models first for many-to-many relationships
		(*DeviceTag)(nil),
		(*PublicKeyTag)(nil),

		(*APIKey)(nil),
		(*Device)(nil),
		(*DeviceSettings)(nil),
		(*Membership)(nil),
		(*Namespace)(nil),
		(*NamespaceSettings)(nil),
		(*PrivateKey)(nil),
		(*PublicKey)(nil),
		(*Session)(nil),
		(*ActiveSession)(nil),
		(*SessionEvent)(nil),
		(*System)(nil),
		(*Tag)(nil),
		(*User)(nil),
	}
}
