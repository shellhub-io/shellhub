package entity

import (
	"time"

	"github.com/uptrace/bun"
)

// EnrollmentCallbackRedemption records a redeemed deferred-decision callback token by its JWT id, so
// the token can only be spent once.
type EnrollmentCallbackRedemption struct {
	bun.BaseModel `bun:"table:enrollment_callback_redemptions"`

	JTI        string    `bun:"jti,pk"`
	RedeemedAt time.Time `bun:"redeemed_at"`
}
