package seo

import (
	"github.com/ottemo/commerce/app/models"
	"github.com/ottemo/commerce/app/models/seo"
)

// GetCollection returns collection of current instance type
func (it *DefaultSEOItem) GetCollection() models.InterfaceCollection {
	model, err := models.GetModel(ConstCollectionNameURLRewrites)
	if err != nil {
		return nil
	}
	if result, ok := model.(seo.InterfaceSEOCollection); ok {
		return result
	}

	return nil
}
