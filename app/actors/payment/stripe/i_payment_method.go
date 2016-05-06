package stripe

import (
	"fmt"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/card"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/customer"
	"strings"
)

func (it *Payment) GetCode() string {
	return ConstPaymentCode
}

func (it *Payment) GetInternalName() string {
	return ConstPaymentName
}

func (it *Payment) GetName() string {
	return it.ConfigNameInCheckout()
}

func (it *Payment) GetType() string {
	return checkout.ConstPaymentTypeCreditCard
}

func (it *Payment) IsAllowed(checkoutInstance checkout.InterfaceCheckout) bool {
	return it.ConfigIsEnabled()
}

func (it *Payment) IsTokenable(checkoutInstance checkout.InterfaceCheckout) bool {
	isTokenable := true
	fmt.Println("stripe - isTokenable called", isTokenable)
	return isTokenable
}

func (it *Payment) Authorize(orderInstance order.InterfaceOrder, paymentInfo map[string]interface{}) (interface{}, error) {
	// Set our api key, applies to any http calls
	stripe.Key = it.ConfigAPIKey()

	// Check if we are just supposed to create a Customer (aka a token)
	action := paymentInfo[checkout.ConstPaymentActionTypeKey]
	isCreateToken := utils.InterfaceToString(action) == checkout.ConstPaymentActionTypeCreateToken
	if isCreateToken {
		// NOTE: `orderInstance = nil` when creating a token

		// 1. Get our customer token
		extra := utils.InterfaceToMap(paymentInfo["extra"])
		visitorID := utils.InterfaceToString(extra["visitor_id"])
		stripeCID := getStripeCustomerToken(visitorID)
		if stripeCID == "" {

			// 2. We don't have a stripe id on file, make a new customer
			c, err := createCustomer(paymentInfo)
			if err != nil {
				return nil, env.ErrorDispatch(err)
			}

			stripeCID = c.ID
		}

		// 3. Create a card
		ca, err := createCard(stripeCID, paymentInfo)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		// This response looks like our normal authorize response
		// but this map is translated into other keys to store a token
		result := map[string]interface{}{
			"transactionID":      ca.ID,                        // token_id
			"creditCardLastFour": ca.LastFour,                  // number
			"creditCardType":     getCCBrand(string(ca.Brand)), // type
			"creditCardExp":      formatCardExp(ca),            // expiration_date
			"customerID":         stripeCID,                    // customer_id
		}

		return result, nil
	}

	// We are not creating a token, so we are charging against a token
	cardID := ""
	stripeCID := ""
	if ccInfo, present := paymentInfo["cc"]; present {
		if creditCard, ok := ccInfo.(visitor.InterfaceVisitorCard); ok && creditCard != nil {
			cardID = creditCard.GetToken()
			stripeCID = creditCard.GetCustomerID()
			fmt.Println("cardID: ", cardID, "stripeCID: ", stripeCID)
		}
	}
	if cardID == "" || stripeCID == "" {
		err := env.ErrorNew(ConstErrorModule, 1, "02128bc6-83d6-4c12-ae90-900a94adb3ad", "Stripe Authorize called without valid tokens")
		return nil, env.ErrorDispatch(err)
	}

	// Assemble charge - https://stripe.com/docs/api/go#create_charge
	chParams := stripe.ChargeParams{
		Currency: "usd",
		Amount:   uint64(orderInstance.GetGrandTotal() * 100), // Amount is in cents
		Customer: stripeCID,                                   // Mandatory
		Email:    utils.InterfaceToString(orderInstance.Get("customer_email")),
	}
	chParams.SetSource(cardID)
	ch, err := charge.New(&chParams)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// env.LogEvent(env.LogFields{"api_response": ch}, "charge")

	// Assemble the response
	orderPaymentInfo := map[string]interface{}{
		"transactionID":     ch.ID,
		"creditCardNumbers": ch.Source.Card.LastFour,
		"creditCardExp":     formatCardExp(*ch.Source.Card),
		"creditCardType":    getCCBrand(string(ch.Source.Card.Brand)),
	}

	return orderPaymentInfo, nil
}

// returns mmyy
func formatCardExp(c stripe.Card) string {
	ccExp := utils.InterfaceToString(c.Month)
	if c.Month < 10 {
		ccExp = "0" + ccExp
	}
	ccExp = ccExp + utils.InterfaceToString(c.Year)[:2]

	return ccExp
}

func createCustomer(paymentInfo map[string]interface{}) (stripe.Customer, error) {
	extra := utils.InterfaceToMap(paymentInfo["extra"])

	c, err := customer.New(&stripe.CustomerParams{
		Email: utils.InterfaceToString(extra["email"]),
		// TODO: coupons?
	})
	if err != nil {
		return stripe.Customer{}, err
	}

	env.LogEvent(env.LogFields{"api_response": c}, "customer") // TODO: COMMENT OUT

	return *c, nil
}

func createCard(stripeCID string, paymentInfo map[string]interface{}) (stripe.Card, error) {

	// Assemble card params
	ccInfo := utils.InterfaceToMap(paymentInfo["cc"])
	ccCVC := utils.InterfaceToString(ccInfo["cvc"])
	if ccCVC == "" {
		err := env.ErrorNew(ConstErrorModule, 1, "15edae76-1d3e-4e7a-a474-75ffb61d26cb", "CVC field was left empty")
		return stripe.Card{}, err
	}
	// ccName := orderInstance.GetBillingAddress().GetFirstName() + " " + orderInstance.GetBillingAddress().GetLastName()

	c, err := card.New(&stripe.CardParams{
		Customer: stripeCID,
		Number:   utils.InterfaceToString(ccInfo["number"]),
		Month:    utils.InterfaceToString(ccInfo["expire_month"]),
		Year:     utils.InterfaceToString(ccInfo["expire_year"]),
		CVC:      ccCVC, // Optional, highly recommended
		// Name:   ccName, // Optional
		// Address fields can be passed here as well to aid in fraud prevention
	})
	if err != nil {
		return stripe.Card{}, err
	}

	env.LogEvent(env.LogFields{"api_response": c}, "card")

	// dereference the pointer
	return *c, nil
}

func getStripeCustomerToken(vid string) string {
	const customerTokenPrefix = "cus"

	if vid == "" {
		env.ErrorDispatch(env.ErrorNew(ConstErrorModule, 1, "2ecfa3ec-7cfc-4783-9060-8467ca63beae", "empty vid passed to look up customer token"))
		return ""
	}

	tokens := visitor.LoadByVID(vid)
	env.LogEvent(env.LogFields{"token_list": tokens, "vid": vid}, "get customer token")
	for _, t := range tokens {
		ts := utils.InterfaceToString(t.Extra["customer_id"])

		// Double check that this field is filled out
		if strings.HasPrefix(ts, customerTokenPrefix) {
			return ts
		}
	}

	return ""
}

func (it *Payment) Capture(orderInstance order.InterfaceOrder, paymentInfo map[string]interface{}) (interface{}, error) {
	return nil, env.ErrorNew(ConstErrorModule, 1, "05199a06-7bd4-49b6-9fb0-0f1589a9cd74", "called but not implemented")
}

func (it *Payment) Refund(orderInstance order.InterfaceOrder, paymentInfo map[string]interface{}) (interface{}, error) {
	return nil, env.ErrorNew(ConstErrorModule, 1, "c8768719-80ab-453d-b52e-513dfb4aab22", "called but not implemented")
}

func (it *Payment) Void(orderInstance order.InterfaceOrder, paymentInfo map[string]interface{}) (interface{}, error) {
	return nil, env.ErrorNew(ConstErrorModule, 1, "4194a950-18fd-4b0d-96e6-e33e930f4320", "called but not implemented")
}
