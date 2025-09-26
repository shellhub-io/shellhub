package entity

func Entities() []any {
	return []any{
		(*APIKey)(nil),
		(*Device)(nil),
		(*Membership)(nil),
		(*Namespace)(nil),
		(*PrivateKey)(nil),
		(*PublicKey)(nil),
		(*Tag)(nil),
		(*DeviceTag)(nil),
		(*PublicKeyTag)(nil),
		(*User)(nil),
	}
}
