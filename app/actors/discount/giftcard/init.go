package giftcard

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
)

// init makes package self-initialization routine
func init() {
	instance := new(DefaultGiftcard)
	var _ checkout.InterfacePriceAdjustment = instance

	checkout.RegisterPriceAdjustment(instance)

	freeShipping := new(Shipping)
	var _ checkout.InterfaceShippingMethod = freeShipping
	checkout.RegisterShippingMethod(freeShipping)

	db.RegisterOnDatabaseStart(setupDB)
	env.RegisterOnConfigStart(setupConfig)
	api.RegisterOnRestServiceStart(setupAPI)

	app.OnAppStart(onAppStart)
}

// DB preparations for current model implementation
func setupDB() error {
	collection, err := db.GetCollection(ConstCollectionNameGiftCard)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	collection.AddColumn("code", db.ConstTypeID, true)
	collection.AddColumn("sku", db.TypeWPrecision(db.ConstTypeVarchar, 100), false)

	collection.AddColumn("amount", db.ConstTypeMoney, false)

	collection.AddColumn("order_id", db.ConstTypeID, true)
	collection.AddColumn("visitor_id", db.ConstTypeID, false)

	collection.AddColumn("status", db.TypeWPrecision(db.ConstTypeVarchar, 50), false)
	collection.AddColumn("orders_used", db.ConstTypeJSON, false)

	collection.AddColumn("name", db.TypeWPrecision(db.ConstTypeVarchar, 150), false)
	collection.AddColumn("message", db.TypeWPrecision(db.ConstTypeVarchar, 150), false)

	collection.AddColumn("recipient_mailbox", db.TypeWPrecision(db.ConstTypeVarchar, 100), false)
	collection.AddColumn("delivery_date", db.ConstTypeDatetime, true)
	collection.AddColumn("created_at", db.ConstTypeDatetime, true)

	return nil
}

// onAppStart makes module initialization on application startup
func onAppStart() error {

	env.EventRegisterListener("checkout.success", checkoutSuccessHandler)
	env.EventRegisterListener("order.proceed", orderProceedHandler)
	env.EventRegisterListener("order.rollback", orderRollbackHandler)

	if scheduler := env.GetScheduler(); scheduler != nil {
		scheduler.RegisterTask("sendGiftCards", SendTask)
		scheduler.ScheduleRepeat("0 8 * * *", "sendGiftCards", nil)
	}

	return nil
}
