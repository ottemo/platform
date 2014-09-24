package visitor

import (
	"errors"
	"github.com/ottemo/foundation/app/actors/visitor/address"
	"github.com/ottemo/foundation/db"
)

// returns current product id
func (it *DefaultVisitor) GetId() string {
	return it.id
}

// sets current product id
func (it *DefaultVisitor) SetId(NewId string) error {
	it.id = NewId
	return nil
}

// loads visitor information from DB
func (it *DefaultVisitor) Load(Id string) error {

	collection, err := db.GetCollection(COLLECTION_NAME_VISITOR)
	if err != nil {
		return err
	}

	values, err := collection.LoadById(Id)
	if err != nil {
		return err
	}

	err = it.FromHashMap(values)
	if err != nil {
		return err
	}

	return nil
}

// removes current visitor from DB
func (it *DefaultVisitor) Delete() error {

	collection, err := db.GetCollection(COLLECTION_NAME_VISITOR)
	if err != nil {
		return err
	}
	addressCollection, err := db.GetCollection(address.COLLECTION_NAME_VISITOR_ADDRESS)
	if err != nil {
		return err
	}

	addressCollection.AddFilter("visitor_id", "=", it.GetId())
	if _, err := addressCollection.Delete(); err != nil {
		return err
	}

	err = collection.DeleteById(it.GetId())
	if err != nil {
		return err
	}

	return nil
}

// stores current visitor to DB
func (it *DefaultVisitor) Save() error {

	collection, err := db.GetCollection(COLLECTION_NAME_VISITOR)
	if err != nil {
		return err
	}

	if it.GetId() == "" {
		collection.AddFilter("email", "=", it.GetEmail())
		n, err := collection.Count()
		if err != nil {
			return err
		}
		if n > 0 {
			return errors.New("email already exists")
		}
	}

	storableValues := it.ToHashMap()

	delete(storableValues, "billing_address")
	delete(storableValues, "shipping_address")

	/*if it.Password == "" {
		return errors.New("password can't be blank")
	}*/

	storableValues["facebook_id"] = it.FacebookId
	storableValues["google_id"] = it.GoogleId
	storableValues["password"] = it.Password
	storableValues["validate"] = it.ValidateKey

	// shipping address save
	if it.ShippingAddress != nil {
		err := it.ShippingAddress.Save()
		if err != nil {
			return err
		}

		storableValues["shipping_address_id"] = it.ShippingAddress.GetId()
	}

	// billing address save
	if it.BillingAddress != nil {
		err := it.BillingAddress.Save()
		if err != nil {
			return err
		}

		storableValues["billing_address_id"] = it.BillingAddress.GetId()
	}

	// saving visitor
	if newId, err := collection.Save(storableValues); err == nil {
		it.Set("_id", newId)
	} else {
		return err
	}

	return nil
}
