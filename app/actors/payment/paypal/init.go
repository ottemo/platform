package paypal

import (
	"github.com/ottemo/commerce/api"
	"github.com/ottemo/commerce/env"

	"github.com/ottemo/commerce/app/models/checkout"
)

// init makes package self-initialization routine
func init() {
	if err := checkout.RegisterPaymentMethod(new(Express)); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "4f72b52a-9af1-4725-a1dc-6731774be323", err.Error())
	}
	if err := checkout.RegisterPaymentMethod(new(PayFlowAPI)); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "4abc8432-31e5-4ef7-bcdc-362b4b32fa8a", err.Error())
	}
	api.RegisterOnRestServiceStart(setupAPI)
	env.RegisterOnConfigStart(setupConfig)
}
