package gotaxcloud_test

import (
	"testing"
	"time"

	"github.com/ottemo/foundation/app/actors/tax/taxcloud/gotaxcloud"
)

func TestAuthorized(t *testing.T) {
	var unique = getUniqueStr()
	var cartID = "AUTHORIZE-CART-UUID" + unique
	var customerID = "AUTHORIZE-CUSTOMER-UUID" + unique
	var orderID = "AUTHORIZE-ORDER-UUID" + unique

	var err error

	_, err = doCorrectLookup(cartID, customerID)
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}

	_, err = doAuthorized(cartID, customerID, orderID)
	if err != nil {
		t.Fatalf("unexpected error '%s'", err)
	}
}

func doAuthorized(cartID, customerID, orderID string) (*gotaxcloud.ResponseBase, error) {
	return testGateway.Authorized(gotaxcloud.AuthorizedParams{
		CartID:         cartID,
		CustomerID:     customerID,
		OrderID:        orderID,
		DateAuthorized: time.Now(),
	})
}