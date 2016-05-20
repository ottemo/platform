package flatweight

import (
	"github.com/ottemo/foundation/env"
)

const (
	ConstShippingCode = "flat_weight"
	ConstShippingName = "Flat Weight"

	ConstConfigPathGroup   = "shipping.flat_weight"
	ConstConfigPathEnabled = "shipping.flat_weight.enabled"
	ConstConfigPathRates   = "shipping.flat_weight.rates"

	ConstErrorModule = "shipping/flatweight"
	ConstErrorLevel  = env.ConstErrorLevelActor
)

// Package global vars
var (
	rates Rates
)

// ShippingMethod is a implementer of InterfaceShippingMethod for a "Flat Weight" shipping method
type ShippingMethod struct{}

type Rates []Rate

type Rate struct {
	Title      string
	Code       string
	Price      float64
	WeightFrom float64
	WeightTo   float64
}
