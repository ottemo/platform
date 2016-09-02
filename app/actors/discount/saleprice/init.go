package saleprice

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models/discount/saleprice"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/db"
)

// init makes package self-initialization routine
func init() {
	salePriceInstance := new(DefaultSalePrice)
	var _ saleprice.InterfaceSalePrice = salePriceInstance
	models.RegisterModel(saleprice.ConstModelNameSalePrice, salePriceInstance)

	salePriceCollectionInstance := new(DefaultSalePriceCollection)
	var _ saleprice.InterfaceSalePriceCollection = salePriceCollectionInstance
	models.RegisterModel(saleprice.ConstModelNameSalePriceCollection, salePriceCollectionInstance)

	db.RegisterOnDatabaseStart(salePriceInstance.setupDB)
	api.RegisterOnRestServiceStart(setupAPI)
}

