package entity

// Entities lists every model bun must know about, so that relations and test fixtures resolve.
// A new table is invisible to both until it is added here.
func Entities() []any {
	return []any{
		(*DeviceTag)(nil),
		(*PublicKeyTag)(nil),
		(*AccessPolicyTag)(nil),

		(*AccessPolicy)(nil),
		(*APIKey)(nil),
		(*Device)(nil),
		(*InstanceAPIKey)(nil),
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
