package models

import "time"

// BillingStatus represents the status of a subscription.
//
// https://stripe.com/docs/api/subscriptions/object#subscription_object-status
// https://stripe.com/docs/billing/subscriptions/overview#subscription-lifecycle
type BillingStatus string

// IsActive returns true if the subscription is active.
// It is active if its status is `active`, `past_due`, `trailing` or `to_cancel_at_end_of_period`.
func (s BillingStatus) IsActive() bool {
	return s == BillingStatusActive || s == BillingStatusPastDue || s == BillingStatusTrialing || s == BillingStatusToCancelAtEndOfPeriod
}

// Represents the possible statuses of a subscription.
const (
	// BillingStatusInactive represents inactive status.
	BillingStatusInactive BillingStatus = "inactive"
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

// Billing contains information about the ShellHub's subscription.
type Billing struct {
	// Active indicates if the subscription is active.
	// IT IS THE SOURCE OF TRUTH THAT DEFINES WHETHER A SUBSCRIPTION IS ACTIVE OR NOT and change due to the status of
	// the subscription.
	//
	// A subscription is active if its status is `active`, `trailing`, `past_due` or `to_cancel_at_end_of_period`.
	// `past_due` is a temporary status that occurs when a payment to renew the subscription fails, but the subscription
	// has not been canceled yet.
	// `to_cancel_at_end_of_period` is a custom status used by this package to indicate that the subscription is set to
	// cancel at the end of the period.
	// A subscription is not active if its status is `incomplete`, `incomplete_expired`, `canceled`, `unpaid` or `paused`.
	// TODO: evaluate if `paused` should be considered active.
	Active bool `json:"active" bson:"active"`
	// Status is the current status of the subscription.
	Status BillingStatus `json:"status" bson:"status"`
	// Customer is the ID of the customer the subscription belongs to.
	// Customer string `json:"customer" bson:"customer"`
	CustomerID string `json:"customer_id" bson:"customer_id"`
	// SubscriptionID is the ID of the subscription.
	SubscriptionID string `json:"subscription_id" bson:"subscription_id"`
	// CurrentPeriodEnd is the end of the current period.
	CurrentPeriodEnd int64 `json:"current_period_end" bson:"current_period_end"`
	// CreatedAt is the time at which this billing was created.
	// It must follow the RFC 3339 format.
	CreatedAt string `json:"created_at" bson:"created_at"`
	// UpdatedAt is the time at which this billing was last updated.
	// It must follow the RFC 3339 format.
	UpdatedAt string `json:"updated_at" bson:"updated_at"`
}

func NewBilling(status BillingStatus, customer, subscription string, currentPeridoEnd int64) *Billing {
	return &Billing{
		Active:           status.IsActive(),
		Status:           status,
		CustomerID:       customer,
		SubscriptionID:   subscription,
		CurrentPeriodEnd: currentPeridoEnd,
		CreatedAt:        time.Now().Format(time.RFC3339),
		UpdatedAt:        time.Now().Format(time.RFC3339),
	}
}

func (b *Billing) IsNil() bool {
	return b == nil
}

// IsActive indicates if the subscription is active.
func (b *Billing) IsActive() bool {
	return b != nil && b.Active
}

func (b *Billing) HasCutomer() bool {
	return b != nil && b.CustomerID != ""
}

func (b *Billing) HasSubscription() bool {
	return b != nil && b.SubscriptionID != ""
}

func (b *Billing) HasCurrentPeriodEnd() bool {
	return b != nil && b.CurrentPeriodEnd != 0
}

// UpdateBillingStatus updates the status of the billing.
func (b *Billing) UpdateBillingStatus(status BillingStatus) {
	b.Active = status.IsActive()
	b.Status = status
}

func (b *Billing) SetCustomer(id string) {
	b.CustomerID = id
}

func (b *Billing) SetSubscription(id string, status BillingStatus) {
	b.Active = status.IsActive()
	b.Status = status
	b.SubscriptionID = id
}

func (b *Billing) SetCurrentPeriodEnd(end int64) {
	b.CurrentPeriodEnd = end
}

// BillingEvaluation contains information about the billing evaluation of acceptance and connection.
// It is used to evaluate if a device can be accepted or a connection SSH can be created. Its idea is simplify the
// check the state of the namespace when related to billing.
type BillingEvaluation struct {
	// CanAccept indicates if the namespace can accept a new device.
	CanAccept bool `json:"can_accept"`
	// CanConnect indicates if the namespace can create a new connection SSH.
	CanConnect bool `json:"can_connect"`
}
