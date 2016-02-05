package subscription

import (
	"strings"
	"time"

	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/subscription"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// checkoutSuccessHandler is a handler for checkout success event which creates the subscriptions
func checkoutSuccessHandler(event string, eventData map[string]interface{}) bool {

	if !subscriptionEnabled {
		return true
	}

	var currentCheckout checkout.InterfaceCheckout
	if eventItem, present := eventData["checkout"]; present {
		if typedItem, ok := eventItem.(checkout.InterfaceCheckout); ok {
			currentCheckout = typedItem
		}
	}

	// means current order is placed by subscription handler
	if currentCheckout == nil || !currentCheckout.IsSubscription() || currentCheckout.GetInfo("subscription_id") != nil {
		return true
	}

	// allows subscription only for registered
	//	if currentCheckout.GetVisitor() == nil {
	//		return true
	//	}

	var checkoutOrder order.InterfaceOrder
	if eventItem, present := eventData["order"]; present {
		if typedItem, ok := eventItem.(order.InterfaceOrder); ok {
			checkoutOrder = typedItem
		}
	}

	if checkoutOrder != nil {
		go subscriptionCreate(currentCheckout, checkoutOrder)
	}

	return true
}

// subscriptionCreate is invoked via a go routine to create subscription based on finished checkout
func subscriptionCreate(currentCheckout checkout.InterfaceCheckout, checkoutOrder order.InterfaceOrder) error {

	currentCart := currentCheckout.GetCart()
	if currentCart == nil {
		return env.ErrorNew(ConstErrorModule, env.ConstErrorLevelActor, "ae108000-68ff-419f-b443-2df1554dd377", "No cart")
	}

	subscriptionItems := make(map[int]int)
	for _, cartItem := range currentCart.GetItems() {
		itemOptions := cartItem.GetOptions()
		if optionValue, present := itemOptions[subscription.ConstSubscriptionOptionName]; present {
			subscriptionItems[cartItem.GetIdx()] = subscription.GetSubscriptionPeriodValue(utils.InterfaceToString(optionValue))
		}
	}

	if len(subscriptionItems) == 0 {
		return nil
	}

	subscriptionInstance, err := subscription.GetSubscriptionModel()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	visitorCreditCard := retrieveCreditCard(currentCheckout, checkoutOrder)
	if visitorCreditCard == nil || visitorCreditCard.GetToken() == "" {
		return env.ErrorNew(ConstErrorModule, env.ConstErrorLevelActor, "333d3396-fddc-4aff-a3fe-083e50a2e1a6", "Credit card can't be obtained")
	}

	if err := validateCheckoutToSubscribe(currentCheckout); err != nil {
		return env.ErrorDispatch(err)
	}

	if err = subscriptionInstance.SetCreditCard(visitorCreditCard); err != nil {
		return env.ErrorDispatch(err)
	}

	visitor := currentCheckout.GetVisitor()
	if visitor != nil {
		subscriptionInstance.Set("visitor_id", visitor.GetID())
		subscriptionInstance.Set("customer_email", visitor.GetEmail())
		subscriptionInstance.Set("customer_name", visitor.GetFullName())
	} else {
		subscriptionInstance.Set("customer_email", currentCheckout.GetInfo("customer_email"))
		subscriptionInstance.Set("customer_name", currentCheckout.GetInfo("customer_name"))
	}

	subscriptionInstance.SetShippingAddress(currentCheckout.GetShippingAddress())
	subscriptionInstance.SetBillingAddress(currentCheckout.GetBillingAddress())

	shippingMethod := currentCheckout.GetShippingMethod()
	var shippingRate checkout.StructShippingRate

	if checkoutShippingRate := currentCheckout.GetShippingRate(); checkoutShippingRate != nil {
		shippingRate.Code = checkoutShippingRate.Code
		shippingRate.Name = checkoutShippingRate.Name
		shippingRate.Price = checkoutShippingRate.Price
	}

	// obtaining values of shipping method and rate from order if they weren't provided in checkout
	if shippingMethod == nil || shippingRate.Code == "" {

		shippingParts := strings.Split(checkoutOrder.GetShippingMethod(), "/")
		orderShippingMethod := checkout.GetShippingMethodByCode(shippingParts[0])

		for _, orderShippingRate := range orderShippingMethod.GetRates(currentCheckout) {
			if shippingParts[1] == orderShippingRate.Code {
				shippingRate = checkout.StructShippingRate{
					Name:  orderShippingRate.Name,
					Code:  orderShippingRate.Code,
					Price: orderShippingRate.Price,
				}
				shippingMethod = orderShippingMethod

				break
			}
		}
	}

	subscriptionInstance.SetShippingMethod(shippingMethod)
	subscriptionInstance.SetShippingRate(checkout.StructShippingRate{
		Name:  shippingRate.Name,
		Code:  shippingRate.Code,
		Price: shippingRate.Price,
	})

	subscriptionInstance.SetStatus(subscription.ConstSubscriptionStatusConfirmed)
	subscriptionInstance.Set("order_id", checkoutOrder.GetID())

	subscriptionTime := time.Now().Truncate(time.Hour)

	// create unique subscriptions for every subscription product
	for _, cartItem := range currentCart.GetItems() {
		if subscriptionPeriodValue, present := subscriptionItems[cartItem.GetIdx()]; present && subscriptionPeriodValue != 0 {

			if err = subscriptionInstance.SetActionDate(subscriptionTime); err != nil {
				env.LogError(err)
				continue
			}

			if err = subscriptionInstance.SetPeriod(subscriptionPeriodValue); err != nil {
				env.LogError(err)
				continue
			}

			if err = subscriptionInstance.UpdateActionDate(); err != nil {
				env.LogError(err)
				continue
			}

			var items []subscription.StructSubscriptionItem

			// populate the subscription object
			subscriptionItem := subscription.StructSubscriptionItem{
				Name:      "",
				ProductID: cartItem.GetProductID(),
				Qty:       cartItem.GetQty(),
				Options:   cartItem.GetOptions(),
			}

			if product := cartItem.GetProduct(); product != nil {
				product.ApplyOptions(subscriptionItem.Options)
				subscriptionItem.Name = product.GetName()
				subscriptionItem.Sku = product.GetSku()
				subscriptionItem.Price = product.GetPrice()
			}

			items = append(items, subscriptionItem)

			subscriptionInstance.Set("items", items)
			subscriptionInstance.SetID("")

			if err = subscriptionInstance.Save(); err != nil {
				env.LogError(err)
				continue
			}
		}
	}

	return nil
}

// getOptionsExtend is a handler for product get options event which extend available product options
// TODO: create some defined object for options (should explain keys)
func getOptionsExtend(event string, eventData map[string]interface{}) bool {

	if !subscriptionEnabled {
		return true
	}

	if value, present := eventData["options"]; present {
		options := utils.InterfaceToMap(value)

		// removing subscription option for products that are not in the list
		if len(subscriptionProducts) > 0 {
			if productID, present := eventData["id"]; !present || !utils.IsInListStr(utils.InterfaceToString(productID), subscriptionProducts) {
				delete(options, subscription.ConstSubscriptionOptionName)
				return true
			}
		}

		storedOptions := map[string]interface{}{
			"type":     "select",
			"required": true,
			"order":    1,
			"label":    "Subscription",
			"options": map[string]interface{}{
				"Just Once": map[string]interface{}{"order": 1, "label": "Just Once"},
				"30 days":   map[string]interface{}{"order": 2, "label": "30 days"},
				"60 days":   map[string]interface{}{"order": 3, "label": "60 days"},
				"90 days":   map[string]interface{}{"order": 4, "label": "90 days"},
				"120 days":  map[string]interface{}{"order": 5, "label": "120 days"},
			},
		}

		// when we are using getOptions for product after they was applied there add field Value
		if subscriptionOption, present := options[subscription.ConstSubscriptionOptionName]; present {
			subscriptionOptionMap := utils.InterfaceToMap(subscriptionOption)
			if appliedValue, present := subscriptionOptionMap["value"]; present {
				storedOptions["value"] = appliedValue
			}
		}

		options[subscription.ConstSubscriptionOptionName] = storedOptions
	}
	return true
}
