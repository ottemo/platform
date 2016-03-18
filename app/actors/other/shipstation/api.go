package shipstation

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

func setupAPI() error {
	service := api.GetRestService()

	service.GET("shipstation", isEnabled(basicAuth(listOrders)))
	// service.POST("shipstation", updateShipmentStatus)

	return nil
}

func isEnabled(next api.FuncAPIHandler) api.FuncAPIHandler {
	return func(context api.InterfaceApplicationContext) (interface{}, error) {
		isEnabled := utils.InterfaceToBool(env.ConfigGetValue(ConstConfigPathShipstationEnabled))

		if !isEnabled {
			// TODO: update status?
			// return "not enabled", nil
			return next(context) //TODO: REMOVE
		}

		return next(context)
	}
}

func basicAuth(next api.FuncAPIHandler) api.FuncAPIHandler {
	return func(context api.InterfaceApplicationContext) (interface{}, error) {

		authHash := utils.InterfaceToString(context.GetRequestSetting("Authorization"))
		username := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathShipstationUsername))
		password := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathShipstationPassword))

		isAuthed := func(authHash string, username string, password string) bool {
			// authHash := "Basic jalsdfjaklsdfjalksdjf"
			hashParts := strings.SplitN(authHash, " ", 2)
			if len(hashParts) != 2 {
				return false
			}

			decodedHash, err := base64.StdEncoding.DecodeString(hashParts[1])
			if err != nil {
				return false
			}

			userPass := strings.SplitN(string(decodedHash), ":", 2)
			if len(userPass) != 2 {
				return false
			}

			return userPass[0] == username && userPass[1] == password
		}

		if !isAuthed(authHash, username, password) {
			// TODO: update status?
			return next(context) //TODO: REMOVE
			// return "not authed", nil
		}

		return next(context)
	}
}

// Your page should return data for any order that was modified between
// the start and end date, regardless of the order’s status.
func listOrders(context api.InterfaceApplicationContext) (interface{}, error) {
	context.SetResponseContentType("text/xml")

	// Our utils.InterfaceToTime doesn't handle this format well `01/23/2012 17:28`
	const parseDateFormat = "01/02/2006 15:04"

	// action := context.GetRequestArgument("action") // only expecting "export"
	// page := context.GetRequestArgument("page")
	startArg := context.GetRequestArgument("start_date")
	endArg := context.GetRequestArgument("end_date")
	startDate, _ := time.Parse(parseDateFormat, startArg)
	endDate, _ := time.Parse(parseDateFormat, endArg)

	// Get the orders
	orderQuery := getOrders(startDate, endDate)

	// Get the order items
	var orderIds []string
	for _, orderResult := range orderQuery {
		orderIds = append(orderIds, orderResult.GetID())
	}
	oiResults := getOrderItems(orderIds)

	// Assemble our response
	response := &Orders{}
	for _, orderResult := range orderQuery {
		responseOrder := buildItem(orderResult, oiResults)
		response.Orders = append(response.Orders, responseOrder)
	}

	return response, nil
}

func getOrders(startDate time.Time, endDate time.Time) []order.InterfaceOrder {
	oModel, _ := order.GetOrderCollectionModel()
	oModel.GetDBCollection().AddFilter("updated_at", ">=", startDate)
	oModel.GetDBCollection().AddFilter("updated_at", "<", endDate)
	result := oModel.ListOrders()

	return result
}

func getOrderItems(orderIds []string) []map[string]interface{} {
	oiModel, _ := order.GetOrderItemCollectionModel()
	oiDB := oiModel.GetDBCollection()
	oiDB.AddFilter("order_id", "in", orderIds)
	oiResults, _ := oiDB.Load()

	return oiResults
}

func buildItem(oItem order.InterfaceOrder, allOrderItems []map[string]interface{}) Order {
	const outputDateFormat = "01/02/2006 15:04"

	// Base Order Details
	createdAt := utils.InterfaceToTime(oItem.Get("created_at"))
	updatedAt := utils.InterfaceToTime(oItem.Get("updated_at"))

	orderDetails := Order{
		OrderId:        oItem.GetID(),
		OrderNumber:    oItem.GetID(),
		OrderDate:      createdAt.Format(outputDateFormat),
		OrderStatus:    oItem.GetStatus(),
		LastModified:   updatedAt.Format(outputDateFormat),
		OrderTotal:     oItem.GetSubtotal(),       // TODO: DOUBLE CHECK THIS IS THE RIGHT ONE, AND FORMAT?
		ShippingAmount: oItem.GetShippingAmount(), // TODO: FORMAT?
	}

	// Customer Details
	orderDetails.Customer.CustomerCode = utils.InterfaceToString(oItem.Get("customer_email"))

	oBillAddress := oItem.GetBillingAddress()
	orderDetails.Customer.BillingAddress = BillingAddress{
		Name: oBillAddress.GetFirstName() + " " + oBillAddress.GetLastName(),
	}

	oShipAddress := oItem.GetShippingAddress()
	orderDetails.Customer.ShippingAddress = ShippingAddress{
		Name:     oShipAddress.GetFirstName() + " " + oShipAddress.GetLastName(),
		Address1: oShipAddress.GetAddressLine1(),
		City:     oShipAddress.GetCity(),
		State:    oShipAddress.GetState(),
		Country:  oShipAddress.GetCountry(),
	}

	// Order Items
	for _, oiItem := range allOrderItems {
		isThisOrder := oiItem["order_id"] == oItem.GetID()
		if !isThisOrder {
			continue
		}

		orderItem := OrderItem{
			Sku:       utils.InterfaceToString(oiItem["sku"]),
			Name:      utils.InterfaceToString(oiItem["name"]),
			Quantity:  utils.InterfaceToInt(oiItem["qty"]),
			UnitPrice: utils.InterfaceToFloat64(oiItem["price"]), // TODO: FORMAT?
		}

		orderDetails.Items = append(orderDetails.Items, orderItem)
	}

	return orderDetails
}
