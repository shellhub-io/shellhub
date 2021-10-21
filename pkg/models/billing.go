package models

import (
	"time"
)

type Billing struct {
	SubscriptionID   string         `json:"subscription_id" bson:"subscription_id,omitempty"`
	CurrentPeriodEnd time.Time      `json:"current_period_end" bson:"current_period_end,omitempty"`
	PriceID          string         `json:"price_id" bson:"price_id,omitempty"`
	CustomerID       string         `json:"customer_id" bson:"customer_id,omitempty"`
	PaymentMethodID  string         `json:"payment_method_id" bson:"payment_method_id,omitempty"`
	PaymentFailed    *PaymentFailed `json:"payment_failed" bson:"payment_failed,omitempty"`
	State            string         `json:"state" bson:"state,omitempty"`
	Active           bool           `json:"active" bson:"active,omitempty"`
}

type PaymentFailed struct {
	Status  bool      `json:"status" bson:"status,omitempty"`
	Amount  float64   `json:"amount" bson:"amount,omitempty"`
	Date    time.Time `json:"date" bson:"date,omitempty"`
	Details string    `json:"details" bson:"details,omitempty"`
}

type UsageRecord struct {
	Timestamp int64      `json:"timestamp" bson:"timestamp,omitempty"`
	Inc       bool       `json:"inc" bson:"type,omitempty"`
	Namespace *Namespace `json:"namespace"`
	Device    *Device    `json:"device"`
}
