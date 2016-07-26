package stripesubscription

import (
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"time"
)

// Package global constants
const (
	ConstConfigPathPlans                  = "payment.stripe.plans"
	ConstErrorModule                      = "stripesubscription"
	ConstErrorLevel                       = env.ConstErrorLevelActor
	ConstCollectionNameStripeSubscription = "stripe_subscription"
)

// DefaultStripeSubscription struct to hold subscription information and represent
// default implementer of InterfaceStripeSubscription
type DefaultStripeSubscription struct {
	id string

	VisitorID       string
	CustomerName    string
	CustomerEmail   string
	BillingAddress  map[string]interface{}
	ShippingAddress map[string]interface{}

	StripeSubscriptionID string
	StripeCustomerID     string
	StripeCoupon         string
	LastPaymentInfo      map[string]interface{}
	NextPaymentAt        time.Time

	Price float64

	CreatedAt time.Time
	UpdatedAt time.Time

	Description string
	Info        map[string]interface{}
	Status      string
}

// DefaultStripeSubscriptionCollection is a default implementer of InterfaceStripeSubscriptionCollection
type DefaultStripeSubscriptionCollection struct {
	listCollection      db.InterfaceDBCollection
	listExtraAttributes []string
}
