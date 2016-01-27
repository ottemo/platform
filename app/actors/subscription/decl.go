// Package subscription implements base subscription functionality
package subscription

import (
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"time"
)

// Package global constants
const (
	ConstErrorModule = "subscription"
	ConstErrorLevel  = env.ConstErrorLevelActor

	ConstConfigPathSubscription        = "general.subscription"
	ConstConfigPathSubscriptionEnabled = "general.subscription.enabled"

	ConstConfigPathSubscriptionEmailSubject  = "general.subscription.emailSubject"
	ConstConfigPathSubscriptionEmailTemplate = "general.subscription.emailTemplate"

	ConstConfigPathSubscriptionStockEmailSubject  = "general.subscription.emailStockSubject"
	ConstConfigPathSubscriptionStockEmailTemplate = "general.subscription.emailStockTemplate"

	ConstCollectionNameSubscription = "subscription"
	ConstSubscriptionLogStorage     = "subscription.log"

	ConstSubscriptionStatusSuspended = "suspended"
	ConstSubscriptionStatusConfirmed = "confirmed"
	ConstSubscriptionStatusCanceled  = "canceled"

	ConstTimeDay = time.Hour * 24
)

var (
	subscriptionEnabled = false
	optionName          = "Subscription"
	optionValues        = map[string]int{
		"Every 5 days": 5, "Every 30 days": 30, "Every 60 days": 60, "Every 90 days": 90, "Every 120 days": 120,
		"":0, "10": 10, "30": 30, "60": 60, "90": 90, "hour": -1, "2hours": -2, "day": 1,
		"Just Once": 0, "30 days": 30, "60 days": 60, "90 days": 90, "120 days": 120,
	}
)

// DefaultSubscription struct to hold subscription information
type DefaultSubscription struct {
	id string

	VisitorID string
	CartID    string
	OrderID   string

	CustomerEmail string
	CustomerName  string

	Status     string
	State      string
	ActionDate time.Time
	Period     int

	ShippingAddress map[string]interface{}
	BillingAddress  map[string]interface{}

	ShippingMethodCode string

	ShippingRate checkout.StructShippingRate

	// should be stored credit card info with payment method in it
	PaymentInstrument map[string]interface{}

	LastSubmit time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// DefaultSubscriptionCollection is a default implementer of InterfaceSubscriptionCollection
type DefaultSubscriptionCollection struct {
	listCollection     db.InterfaceDBCollection
	listExtraAtributes []string
}
