// Package giftcard creates and manage giftCards
package giftcard

import (
	"github.com/ottemo/foundation/env"
)

// Package global constants
const (
	ConstSessionKeyAppliedGiftCardCodes = "applied_giftcard_codes"
	ConstCollectionNameGiftCard         = "gift_card"

	ConstConfigPathGiftEmailTemplate = "general.discounts.giftCard_email"
	ConstConfigPathGiftEmailSubject  = "general.discounts.giftCard_email_subject"
	ConstConfigPathGiftCardSKU       = "general.discounts.giftCard_SKU_code"

	ConstConfigPathGiftCardApplyPriority = "general.discounts.giftCard_apply_priority"

	ConstErrorModule = "giftcard"
	ConstErrorLevel  = env.ConstErrorLevelActor

	ConstGiftCardStatusNew      = "new"
	ConstGiftCardStatusApplied  = "applied"
	ConstGiftCardStatusUsed     = "used"
	ConstGiftCardStatusOverUsed = "over-used"
	ConstGiftCardStatusRefilled = "refilled"
	ConstGiftCardStatusCanceled = "canceled"
)

// DefaultGiftcard is a default implementer of InterfaceDiscount
type DefaultGiftcard struct{}
