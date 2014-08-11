package order

import (
	"errors"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/db"
)

// module entry point before app start
func init() {
	instance := new(DefaultOrder)


	models.RegisterModel("Order", instance)

	db.RegisterOnDatabaseStart(instance.setupDB)

	// api.RegisterOnRestServiceStart(setupAPI)
}

// DB preparations for current model implementation
func (it *DefaultOrder) setupDB() error {

	if dbEngine := db.GetDBEngine(); dbEngine != nil {
		collection, err := dbEngine.GetCollection(ORDER_COLLECTION_NAME)
		if err != nil {
			return err
		}

		collection.AddColumn("increment_id", "varchar(50)", true)
		collection.AddColumn("status", "varchar(50)", true)

		collection.AddColumn("visitor_id", "id", true)
		collection.AddColumn("cart_id", "id", true)

		collection.AddColumn("customer_email", "varchar(100)", true)
		collection.AddColumn("customer_name", "varchar(100)", false)

		collection.AddColumn("payment_method", "varchar(100)", false)
		collection.AddColumn("shipping_method", "varchar(100)", false)

		collection.AddColumn("subtotal", "decimal(10,2)", false)
		collection.AddColumn("discount", "decimal(10,2)", false)
		collection.AddColumn("tax_amount", "decimal(10,2)", false)
		collection.AddColumn("shipping_amount", "decimal(10,2)", false)
		collection.AddColumn("grand_total", "decimal(10,2)", false)

		collection.AddColumn("created_at", "datetime", false)
		collection.AddColumn("updaed_at", "datetime", false)


		collection, err = dbEngine.GetCollection(ORDER_ITEMS_COLLECTION_NAME)
		if err != nil {
			return err
		}

		collection.AddColumn("idx", "int", false)

		collection.AddColumn("order_id", "id", true)
		collection.AddColumn("product_id", "id", true)

		collection.AddColumn("qty", "int", false)

		collection.AddColumn("name", "varchar(150)", false)
		collection.AddColumn("sku", "varchar(100)", false)
		collection.AddColumn("short_description", "varchar(255)", false)

		collection.AddColumn("options", "text", false)

		collection.AddColumn("price", "decimal(10,2)", false)
		collection.AddColumn("weight", "decimal(10,2)", false)
		collection.AddColumn("size", "decimal(10,2)", false)

	} else {
		return errors.New("Can't get database engine")
	}

	return nil
}
