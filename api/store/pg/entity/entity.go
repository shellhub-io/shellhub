package entity

func Entities() []any {
	return []any{
		(*APIKey)(nil),
		(*Membership)(nil),
		(*Namespace)(nil),
		(*User)(nil),
	}
}
