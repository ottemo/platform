package stripe

import (
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	"strings"
)

// getCCBrand Standardize the cc brand
func getCCBrand(ccBrand string) string {
	switch ccBrand {
	case "Visa":
		return "VISA"

	case "American Express":
		return "AmericanExpress"

	case "MasterCard":
		return "MasterCard"

	case "Discover":
		return "Discover"

	case "JCB":
		return "JCB"

	case "Diners Club":
		return "DinersClub"

	case "Unknown":
		return "Unknown"
	}

	return ccBrand
}

// getStripeCustomerToken We attach customer tokens to card tokens in the visitor_token table
// - the customer token is sensitive data because you can make a charge with it alone
// - if you are going to make a charge against a card that is attached to a customer though,
//   you must attach the customer id
func getStripeCustomerToken(vid string) string {
	const customerTokenPrefix = "cus"

	if vid == "" {
		env.ErrorDispatch(env.ErrorNew(ConstErrorModule, 1, "2ecfa3ec-7cfc-4783-9060-8467ca63beae", "empty vid passed to look up customer token"))
		return ""
	}

	model, _ := visitor.GetVisitorCardCollectionModel()
	model.ListFilterAdd("visitor_id", "=", vid)
	model.ListFilterAdd("payment", "=", ConstPaymentCode)

	// 3rd party customer identifier, used by stripe
	err := model.ListAddExtraAttribute("customer_id")
	if err != nil {
		env.ErrorDispatch(err)
	}

	tokens, err := model.List()
	if err != nil {
		env.ErrorDispatch(err)
	}

	for _, t := range tokens {
		ts := utils.InterfaceToString(t.Extra["customer_id"])

		// Double check that this field is filled out
		if strings.HasPrefix(ts, customerTokenPrefix) {
			return ts
		}
	}

	return ""
}
