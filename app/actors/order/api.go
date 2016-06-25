package order

import (
	"encoding/csv"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/actors/discount/giftcard"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	// Admin
	service.GET("orders/attributes", api.IsAdmin(APIListOrderAttributes))
	service.GET("orders", api.IsAdmin(APIListOrders))
	service.POST("orders/exportToCSV", api.IsAdmin(APIExportOrders))

	service.GET("order/:orderID", api.IsAdmin(APIGetOrder))
	service.PUT("order/:orderID", api.IsAdmin(APIUpdateOrder))
	service.DELETE("order/:orderID", api.IsAdmin(APIDeleteOrder))
	service.GET("order/:orderID/emailShipStatus", api.IsAdmin(APISendShipStatusEmail))
	service.GET("order/:orderID/emailOrderConfirmation", api.IsAdmin(APISendOrderConfirmationEmail))

	// Public
	service.GET("visit/orders", APIGetVisitorOrders)
	service.GET("visit/order/:orderID", APIGetVisitorOrder)

	return nil
}

// APIListOrderAttributes returns a list of purchase order attributes
func APIListOrderAttributes(context api.InterfaceApplicationContext) (interface{}, error) {

	orderModel, err := order.GetOrderModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return orderModel.GetAttributesInfo(), nil
}

// APIListOrders returns a list of existing purchase orders
//   - if "action" parameter is set to "count" result value will be just a number of list items
func APIListOrders(context api.InterfaceApplicationContext) (interface{}, error) {

	// taking orders collection model
	orderCollectionModel, err := order.GetOrderCollectionModel()
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	// filters handle
	models.ApplyFilters(context, orderCollectionModel.GetDBCollection())

	// checking for a "count" request
	if context.GetRequestArgument(api.ConstRESTActionParameter) == "count" {
		return orderCollectionModel.GetDBCollection().Count()
	}

	// limit parameter handle
	orderCollectionModel.ListLimit(models.GetListLimit(context))

	// extra parameter handle
	models.ApplyExtraAttributes(context, orderCollectionModel)

	return orderCollectionModel.List()
}

// APIGetOrder return specified purchase order information
//   - order id should be specified in "orderID" argument
func APIGetOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	// pull order id off context
	orderModel, err := getOrder(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// operation
	//----------
	result := orderModel.ToHashMap()
	if notes, present := utils.InterfaceToMap(result["shipping_info"])["notes"]; present {
		utils.InterfaceToMap(result["shipping_address"])["notes"] = notes
	}

	result["items"] = orderModel.GetItems()
	return result, nil
}

// APISendShipStatusEmail will send the visitor a shipping confirmation email
// - order id should be specified in "orderID" argument
func APISendShipStatusEmail(context api.InterfaceApplicationContext) (interface{}, error) {

	orderModel, err := getOrder(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	err = orderModel.SendShippingStatusUpdateEmail()
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	return "Shipping status email sent", nil
}

// APIUpdateOrder update existing purchase order
//   - order id should be specified in "orderID" argument
func APIUpdateOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	orderModel, err := getOrder(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// update the order data from request
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorDispatch(err)
	}

	for attribute, value := range requestData {
		orderModel.Set(attribute, value)
	}

	orderModel.Save()

	return orderModel.ToHashMap(), nil
}

// APIDeleteOrder deletes existing purchase order
//   - order id should be specified in "orderID" argument
func APIDeleteOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	orderID, err := getOrderID(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// grab order ID off request
	orderModel, err := order.GetOrderModelAndSetID(orderID)
	if err != nil {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorDispatch(err)
	}

	orderModel.Delete()
	return "Order deleted: " + orderID, nil
}

// APIGetVisitorOrder returns current visitor order details for specified order
//   - orderID should be specified in arguments
func APIGetVisitorOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	orderModel, err := getOrder(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// allow anonymous visitors through if the session id matches
	if utils.InterfaceToString(orderModel.Get("session_id")) != context.GetSession().GetID() {
		// force anonymous visitors to log in if their session id does not match the one on the order
		visitorID := visitor.GetCurrentVisitorID(context)
		if visitorID == "" {
			return "No Visitor ID found, unable to process order request. Please log in first.", nil
		} else if utils.InterfaceToString(orderModel.Get("visitor_id")) != visitorID {
			context.SetResponseStatusBadRequest()
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "c5ca1fdb-7008-4a1c-a168-9df544df9825", "There is a mis-match between the current Visitor ID and the Visitor ID on the order.")
		}
	}

	result := orderModel.ToHashMap()
	result["items"] = orderModel.GetItems()

	return result, nil
}

// APIGetVisitorOrders returns list of orders related to current visitor
//   - visitorID is required, visitor must be logged in
func APIGetVisitorOrders(context api.InterfaceApplicationContext) (interface{}, error) {

	// list operation
	visitorID := visitor.GetCurrentVisitorID(context)
	if visitorID == "" {
		return "No Visitor ID found, unable to process request.  Please log in first.", nil
	}

	orderCollection, err := order.GetOrderCollectionModel()
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	err = orderCollection.ListFilterAdd("visitor_id", "=", visitorID)
	if err != nil {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorDispatch(err)
	}

	// We only return orders that are in these two states
	statusFilter := [2]string{order.ConstOrderStatusProcessed, order.ConstOrderStatusCompleted}
	orderCollection.GetDBCollection().AddFilter("status", "in", statusFilter)

	descending := true
	orderCollection.GetDBCollection().AddSort("created_at", descending)

	// filters handle
	models.ApplyFilters(context, orderCollection.GetDBCollection())

	// extra parameter handle
	models.ApplyExtraAttributes(context, orderCollection)

	result, err := orderCollection.List()

	return result, env.ErrorDispatch(err)
}

// APISendOrderConfirmationEmail will send out an order confirmation email to the visitor specficied in the orderID
//   - orderID must be passed as a request argument
func APISendOrderConfirmationEmail(context api.InterfaceApplicationContext) (interface{}, error) {

	// load the email template
	email := utils.InterfaceToString(env.ConfigGetValue(checkout.ConstConfigPathConfirmationEmail))

	// load the order
	orderModel, err := getOrder(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// set visitor
	visitor := make(map[string]interface{})
	visitor["first_name"] = orderModel.Get("customer_name")
	visitor["email"] = orderModel.Get("customer_email")

	order := hydrateOrder(orderModel)

	// set store url
	customInfo := make(map[string]interface{})
	customInfo["base_storefront_url"] = utils.InterfaceToString(env.ConfigGetValue(app.ConstConfigPathStorefrontURL))

	// build the email
	confirmationEmail, err := utils.TextTemplate(email, map[string]interface{}{
		"Order":   order,
		"Visitor": visitor,
		"Info":    customInfo,
	})
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return "failure", env.ErrorDispatch(err)
	}

	// send the email
	emailAddress := utils.InterfaceToString(visitor["email"])
	err = app.SendMail(emailAddress, "Order confirmation", confirmationEmail)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return "failure", env.ErrorDispatch(err)
	}

	return "Order confirmation email sent", nil
}

// getOrder will load the order from the database using the orderID from the application context
func getOrder(context api.InterfaceApplicationContext) (order.InterfaceOrder, error) {

	orderID, err := getOrderID(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	orderModel, err := order.LoadOrderByID(orderID)
	if err != nil {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorDispatch(err)
	}

	return orderModel, nil
}

func getOrderID(context api.InterfaceApplicationContext) (string, error) {

	// load orderID
	orderID := context.GetRequestArgument("orderID")
	if orderID == "" {
		context.SetResponseStatusBadRequest()
		return orderID, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "fc3011c7-e58c-4433-b9b0-881a7ba005cf", "No order id found on request, orderID should be specified")
	}

	return orderID, nil
}

func hydrateOrder(orderModel order.InterfaceOrder) map[string]interface{} {

	var orderItems []map[string]interface{}

	order := orderModel.ToHashMap()
	timeZone := utils.InterfaceToString(env.ConfigGetValue(app.ConstConfigPathStoreTimeZone))
	giftCardSku := utils.InterfaceToString(env.ConfigGetValue(giftcard.ConstConfigPathGiftCardSKU))

	// convert date of order creation to store time zone
	if date, present := order["created_at"]; present {
		convertedDate, _ := utils.MakeTZTime(utils.InterfaceToTime(date), timeZone)
		if !utils.IsZeroTime(convertedDate) {
			order["created_at"] = convertedDate
		}
	}

	// load items into the order
	for _, item := range orderModel.GetItems() {
		options := make(map[string]interface{})

		for optionName, optionKeys := range item.GetOptions() {
			optionMap := utils.InterfaceToMap(optionKeys)

			if strings.Contains(item.GetSku(), giftCardSku) {

				// if we have a giftcard date, localize the date
				options[optionName] = ifGiftcardSetLocalDate(item.GetSku(), timeZone, optionMap)
			} else {

				// if not a date, just set the option value
				options[optionName] = optionMap["value"]
			}
		}

		orderItems = append(orderItems, map[string]interface{}{
			"name":    item.GetName(),
			"options": options,
			"sku":     item.GetSku(),
			"qty":     item.GetQty(),
			"price":   item.GetPrice()})
	}

	order["items"] = orderItems
	order["payment_method_title"] = orderModel.GetPaymentMethod()
	order["shipping_method_title"] = orderModel.GetShippingMethod()

	return order
}

func ifGiftcardSetLocalDate(optionName, timeZone string, optionMap map[string]interface{}) string {

	// make sure we are looking for a date
	if utils.IsAmongStr(optionName, "Date", "Delivery Date", "send_date", "Send Date", "date") {

		// localize the date
		giftcardDeliveryDate, _ := utils.MakeTZTime(utils.InterfaceToTime(optionMap["value"]), timeZone)

		// format the date if not zero
		if !utils.IsZeroTime(giftcardDeliveryDate) {
			// TODO: Should be "Monday Jan 2 15:04 (MST)" but we have a bug
			return giftcardDeliveryDate.Format("Monday Jan 2 15:04")
		}
	}

	return utils.InterfaceToString(optionMap["value"])
}

// APIExportOrders returns a list of orders in CSV format
//    - returns orders specified in url parameters
//    - must include at least one order
func APIExportOrders(context api.InterfaceApplicationContext) (interface{}, error) {
	var itemCSVRecords [][]string
	var orders []string

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	// must have at least one order in request
	if !utils.KeysInMapAndNotBlank(requestData, "orders") {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "f2602f73-7cae-4525-8405-9e470681c20e", "Specifiy a minimum of one order in the orders parameter.")
	}

	// look up the order items collection
	orderItemsCollectionModel, err := order.GetOrderItemCollectionModel()
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}
	dbOrderItemsCollection := orderItemsCollectionModel.GetDBCollection()

	// look up the order collection
	orderCollectionModel, err := order.GetOrderCollectionModel()
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}
	dbOrderCollection := orderCollectionModel.GetDBCollection()
	dbOrderCollection.AddSort("created_at", false)

	// load orders based on order IDs passed
	orders = utils.InterfaceToStringArray(utils.InterfaceToArray(requestData["orders"]))
	if orders != nil && len(orders) > 0 && !utils.IsInListStr("all", orders) {
		dbOrderCollection.AddFilter("_id", "in", orders)
	}

	ordersRecords, err := dbOrderCollection.Load()
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	// error if no records returned from db
	if len(ordersRecords) == 0 {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "28eac91b-39ec-4034-b664-4004e940a6d1", "No orders were found.")
	}

	// build the field names for csv file
	for _, columnsHeaders := range orderFields {
		itemCSVRecords = append(itemCSVRecords, columnsHeaders)
	}

	// Webgility importer bombs out on some characters even if they are properly
	// escaped in a csv: . , \ / ( )
	regexCleaner, _ := regexp.Compile(`[.,\/\\()]`)

	for _, orderRecord := range ordersRecords {

		dbOrderItemsCollection.ClearFilters()
		dbOrderItemsCollection.AddFilter("order_id", "=", orderRecord["_id"])

		orderItemsRecords, err := dbOrderItemsCollection.Load()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		orderRecord := utils.InterfaceToMap(orderRecord)

		// convert discount to an absolute value to fix quickbooks bug
		discount := utils.InterfaceToFloat64(orderRecord["discount"])
		discount = math.Abs(discount)
		orderRecord["discount"] = discount
		shippingAddress := utils.InterfaceToMap(orderRecord["shipping_address"])
		billingAddress := utils.InterfaceToMap(orderRecord["billing_address"])

		for orderItemIndex, orderItem := range orderItemsRecords {

			// process an order and convert it to csv format
			for _, inputValues := range dataSet {
				var rowData []string

				// build the csv row from the order data set
				for _, value := range inputValues {
					cellValue := ""

					switch typedValue := value.(type) {
					case string:
						cellValue = typedValue
						switch {
						case strings.Index(cellValue, "$") == 0:
							cellValue = utils.InterfaceToString(orderRecord[strings.Replace(cellValue, "$", "", 1)])
							break

						case strings.HasPrefix(cellValue, "item."):
							cellValue = utils.InterfaceToString(orderItem[strings.Replace(cellValue, "item.", "", 1)])
							break

						case strings.HasPrefix(cellValue, "shipping."):
							addressKey := strings.Replace(cellValue, "shipping.", "", 1)
							cellValue = utils.InterfaceToString(shippingAddress[addressKey])
							cellValue = regexCleaner.ReplaceAllString(cellValue, "")
							break

						case strings.HasPrefix(cellValue, "billing."):
							addressKey := strings.Replace(cellValue, "billing.", "", 1)
							cellValue = utils.InterfaceToString(billingAddress[addressKey])
							cellValue = regexCleaner.ReplaceAllString(cellValue, "")
							break
						}
						break

					case func(record map[string]interface{}) string:

						cellValue = typedValue(orderRecord)
						break

					case func(int, map[string]interface{}) string:
						cellValue = typedValue(orderItemIndex, orderItem)
						break

					}
					// append cellValue to row
					rowData = append(rowData, cellValue)
				}
				// add row to table
				itemCSVRecords = append(itemCSVRecords, rowData)
			}
		}
	}

	// preparing csv writer
	csvWriter := csv.NewWriter(context.GetResponseWriter())

	exportFilename := "orders_export_" + time.Now().Format(time.RFC3339) + ".csv"

	context.SetResponseContentType("text/csv")
	context.SetResponseSetting("Content-disposition", "attachment;filename="+exportFilename)

	for _, csvRecord := range itemCSVRecords {
		csvWriter.Write(csvRecord)
	}
	csvWriter.Flush()

	return "", nil
}
