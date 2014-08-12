package checkout

import (
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/cart"
)



// sets shipping address for checkout
func (it *DefaultCheckout) SetShippingAddress(address visitor.I_VisitorAddress) error {
	it.ShippingAddress = address
	return nil
}



// returns checkout shipping address
func (it *DefaultCheckout) GetShippingAddress() visitor.I_VisitorAddress {
	return it.ShippingAddress
}



// sets billing address for checkout
func (it *DefaultCheckout) SetBillingAddress(address visitor.I_VisitorAddress) error {
	it.BillingAddress = address
	return nil
}



// returns checkout billing address
func (it *DefaultCheckout) GetBillingAddress() visitor.I_VisitorAddress {
	return it.BillingAddress
}



// sets payment method for checkout
func (it *DefaultCheckout) SetPaymentMethod(paymentMethod checkout.I_PaymentMethod) error {
	it.PaymentMethod = paymentMethod
	return nil
}




// returns checkout payment method
func (it *DefaultCheckout) GetPaymentMethod() checkout.I_PaymentMethod {
	return it.PaymentMethod
}



// sets payment method for checkout
func (it *DefaultCheckout) SetShippingMethod(shippingMethod checkout.I_ShippingMehod) error {
	it.ShippingMethod = shippingMethod
	return nil
}



// returns checkout shipping rate
func (it *DefaultCheckout) GetShippingRate() *checkout.T_ShippingRate {
	return it.ShippingRate
}



// sets shipping rate for checkout
func (it *DefaultCheckout) SetShippingRate(shippingRate checkout.T_ShippingRate) error {
	it.ShippingRate = &shippingRate
	return nil
}



// return checkout shipping method
func (it *DefaultCheckout) GetShippingMethod() checkout.I_ShippingMehod {
	return it.ShippingMethod
}



// sets cart for checkout
func (it *DefaultCheckout) SetCart(checkoutCart cart.I_Cart) error {
	it.Cart = checkoutCart
	return nil
}



// return checkout cart
func (it *DefaultCheckout) GetCart() cart.I_Cart {
	return it.Cart
}



// sets visitor for checkout
func (it *DefaultCheckout) SetVisitor(checkoutVisitor visitor.I_Visitor) error {
	it.Visitor = checkoutVisitor
	return nil
}



// return checkout visitor
func (it *DefaultCheckout) GetVisitor() visitor.I_Visitor {
	return it.Visitor
}
