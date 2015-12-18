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

	ConstConfigPathSubscriptionEmailSubject     = "general.subscription.emailSubject"
	ConstConfigPathSubscriptionEmailTemplate    = "general.subscription.emailTemplate"
	ConstConfigPathSubscriptionConfirmationLink = "general.subscription.confirmationLink"

	ConstConfigPathSubscriptionSubmitEmailSubject  = "general.subscription.emailSubmitSubject"
	ConstConfigPathSubscriptionSubmitEmailTemplate = "general.subscription.emailSubmitTemplate"
	ConstConfigPathSubscriptionSubmitEmailLink     = "general.subscription.SubmitLink"

	ConstCollectionNameSubscription = "subscription"

	ConstSubscriptionStatusSuspended = "suspended"
	ConstSubscriptionStatusConfirmed = "confirmed"
	ConstSubscriptionStatusCanceled  = "canceled"

	ConstSubscriptionActionSubmit = "submit"
	ConstSubscriptionActionUpdate = "update"
	ConstSubscriptionActionCreate = "create"

	ConstTimeDay = time.Hour * 24

	ConstCreationDaysDelay = 33
)

var (
	allowedSubscriptionPeriods = []int{10, 30, 55, 60}
	nextCreationDate           time.Time
)

// DefaultSubscription struct to hold subscription information
type DefaultSubscription struct {
	id string

	VisitorID string
	CartID    string
	OrderID   string

	Email string
	Name  string

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
