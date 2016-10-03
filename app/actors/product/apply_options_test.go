package product_test

// This package provide additional product package tests
// To run it use command line
//
// $ go test github.com/ottemo/foundation/app/actors/product/
//
// or, if fmt.Println output required, use it with "-v" flag
//
// $ go test -v github.com/ottemo/foundation/app/actors/product/

import (
	"testing"

	"fmt"

	"github.com/ottemo/foundation/test"
	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/app/models/product"
)

func TestConfigurableProductApplyOptions(t *testing.T) {

	// starting application and getting product model
	err := test.StartAppInTestingMode()
	if err != nil {
		t.Error(err)
	}

	// options to apply
	options := map[string]interface{}{
		"color": "red",
		"field_option": "field_option value",
	}

	// create simple product
	simpleProductData, err := utils.DecodeJSONToStringKeyMap(`{
		"sku": "test-simple",
		"enabled": "true",
		"name": "Test Simple Product",
		"short_description": "something short",
		"description": "something long",
		"default_image": "",
		"price": 1.0,
		"weight": 0.4
	}`)
	if err != nil {
		t.Error(err)
	}

	simpleProductModel, err := product.GetProductModel()
	if err != nil || simpleProductModel == nil {
		t.Error(err)
	}

	// setting values for simple product
	err = simpleProductModel.FromHashMap(simpleProductData)
	if err != nil {
		t.Error(err)
	}

	// saving simple product
	err = simpleProductModel.Save()
	if err != nil {
		t.Error(err)
	}

	// making data for test object
	// options example {"color":{"code":"color","controls_inventory":true,"key":"color","label":"Color","options":{
	// "blue":{"key":"blue","label":"Blue","order":2,"price":"11","sku":"-blue"},
	// "red":{"key":"red","label":"Red","order":1,"price":"10","sku":"-red"}},
	// "order":1,"required":true,"type":"select"},
	// "field_option":{"code":"field_option","controls_inventory":false,"key":"field_option","label":"FieldOption",
	// "order":2,"price":"13","required":false,"sku":"-fo","type":"field"}}

	var productJson = `{
		"id": "123456789012345678901234",
		"sku": "test",
		"name": "Test Product",
		"short_description": "something short",
		"description": "something long",
		"default_image": "",
		"price": 1.1,
		"weight": 0.5,
		"test": "ok",
		"options" : {
			"field_option":{
				"code": "field_option", "controls_inventory": false, "key": "field_option",
				"label": "FieldOption", "order": 2, "price": "13", "required": false,
				"sku": "-fo", "type": "field"},
			"color" : {
				"code": "color", "controls_inventory": true, "key": "color", "label": "Color",
				"order": 1, "required": true, "type": "select",
				"options" : {
					"black": {"order": "3", "key": "black", "label": "Black", "price": 1.3, "sku": "-black"},
					"blue":  {"order": "1", "key": "blue",  "label": "Blue",  "price": 2.0, "sku": "-blue"},
					"red":   {"order": "2", "key": "red",   "label": "Red",   "price": 100, "sku": "-red",`+
					`"`+product.ConstOptionSimpleIDsName + `": ["` + simpleProductModel.GetID()+ `"]}
				}
			}
		}
	}`

	fmt.Println(utils.InterfaceToString("\n= Simple product original ID: " + simpleProductModel.GetID()))

	productData, err := utils.DecodeJSONToStringKeyMap([]byte(productJson))
	if err != nil {
		t.Error(err)
	}

	// populate configurable product model
	productModel, err := product.GetProductModel()
	if err != nil || productModel == nil {
		t.Error(err)
	}

	// setting values for configurable product
	err = productModel.FromHashMap(productData)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("\n= Original configurable product: " + utils.InterfaceToString(productModel))
	var configurableProductID = productModel.GetID()

	// apply options
	err = productModel.ApplyOptions(options)
	if err != nil {
		t.Error("error applying options")
	}

	fmt.Println("\n= Configurable with applied options: " + utils.InterfaceToString(productModel))

	// check "configurable" prodcut populated by simple product values
	var productHashMap = productModel.ToHashMap()
	for key, newValue := range productHashMap {
		originalValue, present := simpleProductData[key]
		var newValueStr = utils.InterfaceToString(newValue)
		if !present {
			fmt.Println("\n= New key [" + key + "] with value [" + newValueStr + "] found")
		}

		var originalValueStr = utils.InterfaceToString(originalValue)
		if present && originalValueStr != newValueStr {
			t.Error("Key [" + key + "] original value [" + originalValueStr + "] not equal to new value [" + newValueStr + "].")
		}
	}

	if productHashMap["_id"] != simpleProductModel.GetID() {
		t.Error("ID is not equal to simple product")
	}

	productOptionHashMap, ok := productHashMap["options"].(map[string]interface{})
	if !ok {
		t.Error("Options are wrong")
	}

	// configurable_id should be stored to reference it in reports
	if productOptionHashMap["configurable_id"] != configurableProductID {
		t.Error("ID of configurable product is not stored")
	}

	// other options should be stored too
	if _, present := productOptionHashMap["field_option"]; !present {
		t.Error("[field_option] absent in new product")
	}

	// deleting simple product
	err = simpleProductModel.Delete()
	if err != nil {
		t.Error(err)
	}

}
