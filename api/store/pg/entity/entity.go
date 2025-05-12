package entity

func Entities() []any {
	return []any{
		(*User)(nil),
		(*Namespace)(nil),
		(*Membership)(nil),
		(*Device)(nil),
		(*APIKey)(nil),
		(*PublicKey)(nil),
		(*PrivateKey)(nil),
	}
}
