package models

// BillingStatus represents the status of a subscription on the payment gateway.
//
// https://stripe.com/docs/api/subscriptions/object#subscription_object-status
// https://stripe.com/docs/billing/subscriptions/overview#subscription-lifecycle
type BillingStatus string

// IsActive returns true if the subscription grants full service.
//
// `past_due` counts as active because Stripe is still retrying the payment, and
// `to_cancel_at_end_of_period` still has a paid period left to run.
func (s BillingStatus) IsActive() bool {
	return s == BillingStatusActive || s == BillingStatusPastDue || s == BillingStatusTrialing || s == BillingStatusToCancelAtEndOfPeriod
}

// Represents the possible statuses of a subscription.
//
// There is no "no subscription" status: a namespace that never completed checkout carries a
// Billing with a nil Subscription instead.
const (
	// BillingStatusActive represents active status without any issues.
	BillingStatusActive BillingStatus = "active"
	// BillingStatusTrialing represents active status without any issues, but the subscription is in trial period.
	BillingStatusTrialing BillingStatus = "trialing"
	// BillingStatusIncomplete represents incomplete status.
	// If the initial payment attempt fails, the status of the subscription becomes incomplete.
	// If payment fails because of a card error, such as a decline, the status of the PaymentIntent is
	// requires_card and the subscription is incomplete.
	BillingStatusIncomplete BillingStatus = "incomplete"
	// BillingStatusIncompleteExpired represents incomplete_expired status.
	// If the first invoice is not paid within 23 hours, the status of the subscription becomes incomplete_expired.
	BillingStatusIncompleteExpired BillingStatus = "incomplete_expired"
	// BillingStatusPastDue represents past_due status.
	// The subscription’s status remains active as long as automatic payments succeed. If automatic payment fails, the
	// subscription updates to past_due and Stripe attempts to recover payment based on your retry rules. If payment
	// recovery fails, you can set the subscription status to canceled, unpaid, or leave it past_due.
	BillingStatusPastDue BillingStatus = "past_due"
	// BillingStatusCanceled represents canceled status.
	BillingStatusCanceled BillingStatus = "canceled"
	// BillingStatusUnpaid represents unpaid status.
	// If the retry attempts are exhausted, the status of the subscription becomes unpaid, depending on your subscriptions settings.
	BillingStatusUnpaid BillingStatus = "unpaid"
	// BillingStatusPaused represents paused status.
	BillingStatusPaused BillingStatus = "paused"
	// BillingStatusToCancelAtEndOfPeriod represents to_cancel_at_end_of_period status.
	// BillingStatusToCancelAtEndOfPeriod is not a Stripe status, but a custom status used by this package to indicate that the subscription is set to cancel at the end of the period.
	BillingStatusToCancelAtEndOfPeriod BillingStatus = "to_cancel_at_end_of_period"
)

// BillingSubscription is the namespace's subscription on the payment gateway.
type BillingSubscription struct {
	// ID is the ID of the subscription on the payment gateway.
	ID string `json:"id"`
	// Status is the current status of the subscription.
	Status BillingStatus `json:"status"`
	// CurrentPeriodEnd is the end of the current period.
	CurrentPeriodEnd int64 `json:"current_period_end"`
}

// Billing contains information about the ShellHub's subscription.
type Billing struct {
	// CustomerID is the ID of the customer on the payment gateway.
	CustomerID string `json:"customer_id"`
	// Subscription is the namespace's subscription, or nil when the namespace has a customer but
	// never completed checkout. A nil Subscription is not a failed subscription: the namespace
	// keeps the free tier, exactly as a namespace with no Billing at all does.
	Subscription *BillingSubscription `json:"subscription,omitempty"`
	// CreatedAt is the time at which this billing was created.
	// It follows the RFC 3339 format, and it is empty on records written before the store began
	// to maintain it.
	CreatedAt string `json:"created_at,omitempty"`
	// UpdatedAt is the time at which this billing was last updated.
	// It follows the RFC 3339 format, and it is empty on records written before the store began
	// to maintain it.
	UpdatedAt string `json:"updated_at,omitempty"`
}

// NewBilling returns a billing record for a namespace that has a customer on the payment gateway
// but no subscription yet. The store stamps CreatedAt and UpdatedAt when it writes the record.
func NewBilling(customerID string) *Billing {
	return &Billing{ //nolint:exhaustruct
		CustomerID: customerID,
	}
}

// Clone returns a deep copy, so that a caller can build the next billing state without touching
// the one the namespace still holds.
func (b *Billing) Clone() *Billing {
	if b == nil {
		return nil
	}

	clone := *b

	if b.Subscription != nil {
		subscription := *b.Subscription
		clone.Subscription = &subscription
	}

	return &clone
}

func (b *Billing) IsNil() bool {
	return b == nil
}

// IsActive indicates whether the namespace has a subscription that grants full service.
func (b *Billing) IsActive() bool {
	return b.HasSubscription() && b.Subscription.Status.IsActive()
}

func (b *Billing) HasCustomer() bool {
	return b != nil && b.CustomerID != ""
}

func (b *Billing) HasSubscription() bool {
	return b != nil && b.Subscription != nil
}

func (b *Billing) SetCustomer(id string) {
	b.CustomerID = id
}

// SetSubscription attaches a subscription to the billing, replacing any subscription already
// there.
func (b *Billing) SetSubscription(id string, status BillingStatus, currentPeriodEnd int64) {
	b.Subscription = &BillingSubscription{
		ID:               id,
		Status:           status,
		CurrentPeriodEnd: currentPeriodEnd,
	}
}

// SetSubscriptionStatus updates the status of the subscription, and does nothing when the billing
// has no subscription.
func (b *Billing) SetSubscriptionStatus(status BillingStatus) {
	if !b.HasSubscription() {
		return
	}

	b.Subscription.Status = status
}

// ClearSubscription detaches the subscription, returning the namespace to the free tier.
func (b *Billing) ClearSubscription() {
	b.Subscription = nil
}

// BillingEvaluation contains information about the billing evaluation of acceptance and connection.
// It is used to evaluate if a device can be accepted or a connection SSH can be created. Its idea is simplify the
// check the state of the namespace when related to billing.
type BillingEvaluation struct {
	// CanAccept indicates if the namespace can accept a new device.
	CanAccept bool `json:"can_accept"`
	// CanConnect indicates if the namespace can create a new connection SSH.
	CanConnect bool `json:"can_connect"`
	// Blocked names what denies the operation, and is empty when nothing does. The two reasons
	// need different answers from the user, so a caller must report them as different errors.
	Blocked BillingBlockReason `json:"blocked,omitempty"`
}

// BillingBlockReason names why a billing evaluation denies acceptance or connection.
type BillingBlockReason string

const (
	// BillingBlockedQuota means the namespace uses every device its plan allows.
	BillingBlockedQuota BillingBlockReason = "quota"
	// BillingBlockedSubscription means the namespace's subscription denies the operation,
	// whatever the device count is.
	BillingBlockedSubscription BillingBlockReason = "subscription"
)
