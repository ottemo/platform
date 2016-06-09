package order

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models"
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

	service.GET("order/:orderID", api.IsAdmin(APIGetOrder))
	service.PUT("order/:orderID", api.IsAdmin(APIUpdateOrder))
	service.DELETE("order/:orderID", api.IsAdmin(APIDeleteOrder))
	service.GET("order/:orderID/sendStatusEmail", api.IsAdmin(APISendStatusEmail))

	// Public
	service.GET("visit/orders", APIGetVisitOrders)
	service.GET("visit/order/:orderID", APIGetVisitOrder)

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

	// check request context
	//---------------------
	orderID := context.GetRequestArgument("orderID")
	if orderID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "723ef443-f974-4455-9be0-a8af13916554", "order id should be specified")
	}

	orderModel, err := order.LoadOrderByID(orderID)
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

func APISendStatusEmail(context api.InterfaceApplicationContext) (interface{}, error) {
	orderID := context.GetRequestArgument("orderID")
	orderModel, err := order.LoadOrderByID(orderID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	err = orderModel.SendShippingStatusUpdateEmail()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "sent", nil
}

// APIUpdateOrder update existing purchase order
//   - order id should be specified in "orderID" argument
func APIUpdateOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	orderID := context.GetRequestArgument("orderID")
	if orderID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "20a08638-e9e6-428b-b70c-a418d7821e4b", "order id should be specified")
	}

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// operation
	//----------
	orderModel, err := order.LoadOrderByID(orderID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for attribute, value := range requestData {
		orderModel.Set(attribute, value)
	}

	orderModel.SetID(orderID)
	orderModel.Save()

	return orderModel.ToHashMap(), nil
}

// APIDeleteOrder deletes existing purchase order
//   - order id should be specified in "orderID" argument
func APIDeleteOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	orderID := context.GetRequestArgument("orderID")
	if orderID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "fc3011c7-e58c-4433-b9b0-881a7ba005cf", "order id should be specified")
	}

	// operation
	//----------
	orderModel, err := order.GetOrderModelAndSetID(orderID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	orderModel.Delete()

	return "ok", nil
}

// APIGetOrder returns current visitor order details for specified order
//   - orderID should be specified in arguments
func APIGetVisitOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	orderModel, err := order.LoadOrderByID(context.GetRequestArgument("orderID"))
	if err != nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "cc719b01-c1e4-4b69-9c89-5735f5c0d339", "Unable to retrieve an order associated with OrderID: "+context.GetRequestArgument("orderID")+".")
	}

	// allow anonymous visitors through if the session id matches
	if utils.InterfaceToString(orderModel.Get("session_id")) != context.GetSession().GetID() {
		// force anonymous visitors to log in if their session id does not match the one on the order
		visitorID := visitor.GetCurrentVisitorID(context)
		if visitorID == "" {
			return "No Visitor ID found, unable to process order request. Please log in first.", nil
		} else if utils.InterfaceToString(orderModel.Get("visitor_id")) != visitorID {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "c5ca1fdb-7008-4a1c-a168-9df544df9825", "There is a mis-match between the current Visitor ID and the Visitor ID on the order.")
		}
	}

	result := orderModel.ToHashMap()
	result["items"] = orderModel.GetItems()

	return result, nil
}

// APIGetOrders returns list of orders related to current visitor
func APIGetVisitOrders(context api.InterfaceApplicationContext) (interface{}, error) {

	// list operation
	//---------------
	visitorID := visitor.GetCurrentVisitorID(context)
	if visitorID == "" {
		return "No Visitor ID found, unable to process request.  Please log in first.", nil
	}

	orderCollection, err := order.GetOrderCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = orderCollection.ListFilterAdd("visitor_id", "=", visitorID)
	if err != nil {
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
