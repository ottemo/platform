package stripesubscription

import (
	"github.com/ottemo/foundation/app/models"
	"time"
)

//InterfaceStripeSubscription represents interface to access business layer implementation of purchase subscription object
type InterfaceStripeSubscription interface {
	GetVisitorID() string
	GetCustomerEmail() string
	GetPeriodEnd() time.Time
	GetStripeSubscriptionID() string

	models.InterfaceModel
	models.InterfaceObject
	models.InterfaceStorable
	models.InterfaceListable
}

//InterfaceStripeSubscriptionCollection represents interface to access business layer implementation of purchase subscription collection
type InterfaceStripeSubscriptionCollection interface {
	ListSubscriptions() []InterfaceStripeSubscription

	models.InterfaceCollection
}