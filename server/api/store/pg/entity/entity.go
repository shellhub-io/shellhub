package entity

func Entities() []any {
	return []any{
		(*DeviceTag)(nil),
		(*PublicKeyTag)(nil),
		(*AccessPolicyTag)(nil),

		(*AccessPolicy)(nil),
		(*APIKey)(nil),
		(*Device)(nil),
		(*Membership)(nil),
		(*Namespace)(nil),
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
