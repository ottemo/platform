package cart

import (
	"errors"

	"github.com/ottemo/foundation/api"

	"github.com/ottemo/foundation/app/models/cart"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/app/utils"
)

func setupAPI() error {

	var err error = nil

	err = api.GetRestService().RegisterAPI("cart", "GET", "info", restCartInfo)
	if err != nil {
		return err
	}
	err = api.GetRestService().RegisterAPI("cart", "POST", "add/:productId/:qty", restCartAdd)
	if err != nil {
		return err
	}
	err = api.GetRestService().RegisterAPI("cart", "PUT", "update/:itemIdx/:qty", restCartUpdate)
	if err != nil {
		return err
	}
	err = api.GetRestService().RegisterAPI("cart", "DELETE", "delete/:itemIdx", restCartDelete)
	if err != nil {
		return err
	}

	return nil
}

// returns cart for current session
func getCurrentCart(params *api.T_APIHandlerParams) (cart.I_Cart, error) {
	sessionCartId := params.Session.Get(cart.SESSION_KEY_CURRENT_CART)

	if sessionCartId != nil && sessionCartId != "" {

		currentCart, err := cart.LoadCartById(utils.InterfaceToString(sessionCartId))
		if err != nil {
			return nil, err
		}

		return currentCart, nil
	} else {

		visitorId := params.Session.Get(visitor.SESSION_KEY_VISITOR_ID)
		if visitorId != nil {
			currentCart, err := cart.GetCartForVisitor(utils.InterfaceToString(visitorId))
			if err != nil {
				return nil, err
			}

			params.Session.Set(cart.SESSION_KEY_CURRENT_CART, currentCart.GetId())

			return currentCart, nil
		} else {
			return nil, errors.New("you are not registered")
		}

	}

	return nil, errors.New("can't get cart for current session")
}

// WEB REST API function to get cart information
//   - parent categories and categorys will not be present in list
func restCartInfo(params *api.T_APIHandlerParams) (interface{}, error) {

	currentCart, err := getCurrentCart(params)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0)

	cartItems := currentCart.ListItems()
	for _, cartItem := range cartItems {

		item := make(map[string]interface{})

		item["_id"] = cartItem.GetId()
		item["idx"] = cartItem.GetIdx()
		item["qty"] = cartItem.GetQty()
		item["pid"] = cartItem.GetProductId()
		item["options"] = cartItem.GetOptions()

		if product := cartItem.GetProduct(); product != nil {
			productData := make(map[string]interface{})

			mediaPath, _ := product.GetMediaPath("image")

			productData["name"] = product.GetName()
			productData["sku"] = product.GetSku()
			productData["image"] = mediaPath + product.GetDefaultImage()
			productData["price"] = product.GetPrice()
			productData["weight"] = product.GetWeight()
			productData["size"] = product.GetSize()

			item["product"] = productData
		}

		items = append(items, item)
	}

	result := map[string]interface{}{
		"visitor_id": currentCart.GetVisitorId(),
		"cart_info":  currentCart.GetCartInfo(),
		"items":      items,
	}

	return result, nil
}

// WEB REST API for adding new item into cart
//   - "pid" (product id) should be specified
//   - "qty" and "options" are optional params
func restCartAdd(params *api.T_APIHandlerParams) (interface{}, error) {

	// check request params
	//---------------------
	reqData, err := api.GetRequestContentAsMap(params)
	if err != nil {
		return nil, err
	}

	var pid string = ""
	reqPid, present := params.RequestURLParams["productId"]
	pid = utils.InterfaceToString(reqPid)
	if !present || pid == "" {
		return nil, errors.New("pid should be specified")
	}

	var qty int = 1
	reqQty, present := params.RequestURLParams["qty"]
	if present {
		qty = utils.InterfaceToInt(reqQty)
	}

	var options map[string]interface{} = nil
	reqOptions, present := reqData["options"]
	if present {
		if tmpOptions, ok := reqOptions.(map[string]interface{}); ok {
			options = tmpOptions
		}
	}

	// operation
	//----------
	currentCart, err := getCurrentCart(params)
	if err != nil {
		return nil, err
	}

	currentCart.AddItem(pid, qty, options)
	currentCart.Save()

	return "ok", nil
}

// WEB REST API used to update cart item qty
//   - "itemIdx" and "qty" should be specified in request URI
func restCartUpdate(params *api.T_APIHandlerParams) (interface{}, error) {

	// check request params
	//---------------------
	if !utils.KeysInMapAndNotBlank(params.RequestURLParams, "itemIdx", "qty") {
		return nil, errors.New("itemIdx and qty should be specified")
	}

	itemIdx, err := utils.StrToInt(params.RequestURLParams["itemIdx"])
	if err != nil {
		return nil, err
	}

	qty, err := utils.StrToInt(params.RequestURLParams["qty"])
	if err != nil {
		return nil, err
	}
	if qty <= 0 {
		return nil, errors.New("qty should be greather then 0")
	}

	// operation
	//----------
	currentCart, err := getCurrentCart(params)
	if err != nil {
		return nil, err
	}

	found := false
	cartItems := currentCart.ListItems()

	for _, cartItem := range cartItems {
		if cartItem.GetIdx() == itemIdx {
			cartItem.SetQty( qty )
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("wrong itemIdx was specified")
	}

	currentCart.Save()

	return "ok", nil
}

// WEB REST API used to delete cart item from cart
//   - "itemIdx" should be specified in request URI
func restCartDelete(params *api.T_APIHandlerParams) (interface{}, error) {

	_, present := params.RequestURLParams["itemIdx"]
	if !present {
		return nil, errors.New("itemIdx should be specified")
	}

	itemIdx, err := utils.StrToInt(params.RequestURLParams["itemIdx"])
	if err != nil {
		return nil, err
	}

	// operation
	//----------
	currentCart, err := getCurrentCart(params)
	if err != nil {
		return nil, err
	}

	err = currentCart.RemoveItem(itemIdx)
	if err != nil {
		return nil, err
	}
	currentCart.Save()

	return "ok", nil
}
