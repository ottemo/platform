package category

import (
	"errors"
	"strings"

	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/category"
	"github.com/ottemo/foundation/app/models/product"
)

//---------------------------------
// IMPLEMENTATION SPECIFIC METHODS
//---------------------------------

// updates path attribute of model
func (it *DefaultCategory) updatePath() {
	if it.GetId() == "" {
		it.Path = ""
	} else if it.Parent != nil {
		parentPath, ok := it.Parent.Get("path").(string)
		if ok {
			it.Path = parentPath + "/" + it.GetId()
		}
	} else {
		it.Path = "/" + it.GetId()
	}
}

//--------------------------
// INTERFACE IMPLEMENTATION
//--------------------------

func (it *DefaultCategory) Get(attribute string) interface{} {
	switch strings.ToLower(attribute) {
	case "_id", "id":
		return it.id

	case "name":
		return it.Name

	case "path":
		if it.Path == "" {
			it.updatePath()
		}
		return it.Path

	case "parent_id":
		if it.Parent != nil {
			return it.Parent.GetId()
		} else {
			return ""
		}

	case "parent":
		return it.Parent

	case "products":
		result := make([]map[string]interface{}, 0)

		for _, categoryProduct := range it.GetProducts() {
			result = append(result, categoryProduct.ToHashMap())
		}

		return result
	}

	return nil
}

func (it *DefaultCategory) Set(attribute string, value interface{}) error {
	attribute = strings.ToLower(attribute)

	switch attribute {
	case "_id", "id":
		it.id = value.(string)

	case "name":
		it.Name = value.(string)

	case "parent_id":
		if value, ok := value.(string); ok {
			value = strings.TrimSpace(value)
			if value != "" {
				model, err := models.GetModel("Category")
				if err != nil {
					return err
				}
				categoryModel, ok := model.(category.I_Category)
				if !ok {
					return errors.New("unsupported category model " + model.GetImplementationName())
				}

				err = categoryModel.Load(value)
				if err != nil {
					return err
				}

				selfId := it.GetId()
				if selfId != "" {
					parentPath, ok := categoryModel.Get("path").(string)
					if categoryModel.GetId() != selfId && ok && !strings.Contains(parentPath, selfId) {
						it.Parent = categoryModel
					} else {
						return errors.New("category can't have sub-category or itself as parent")
					}
				} else {
					it.Parent = categoryModel
				}
			} else {
				it.Parent = nil
			}
		} else {
			return errors.New("unsupported id specified")
		}
		it.updatePath()

	case "parent":
		switch value := value.(type) {
		case category.I_Category:
			it.Parent = value
		case string:
			it.Set("parent_id", value)
		default:
			errors.New("unsupported 'parent' value")
		}
		// path should be changed as well
		it.updatePath()

	case "products":
		switch typedValue := value.(type) {

		case []interface{}:
			for _, listItem := range typedValue {
				productId, ok := listItem.(string)
				if ok {
					productModel, err := product.LoadProductById(productId)
					if err != nil {
						return err
					}

					it.ProductIds = append(it.ProductIds, productModel.GetId())
				}
			}

		case []product.I_Product:
			it.ProductIds = make([]string, 0)
			for _, productItem := range typedValue {
				it.ProductIds = append(it.ProductIds, productItem.GetId())
			}

		default:
			return errors.New("unsupported 'products' value")
		}
	}
	return nil
}

func (it *DefaultCategory) FromHashMap(input map[string]interface{}) error {

	for attribute, value := range input {
		if err := it.Set(attribute, value); err != nil {
			return err
		}
	}

	return nil
}

func (it *DefaultCategory) ToHashMap() map[string]interface{} {

	result := make(map[string]interface{})

	result["_id"] = it.id

	result["parent_id"] = it.Get("parent_id")
	result["name"] = it.Get("name")
	result["products"] = it.Get("products")
	result["path"] = it.Get("path")

	return result
}

func (it *DefaultCategory) GetAttributesInfo() []models.T_AttributeInfo {

	info := []models.T_AttributeInfo{
		models.T_AttributeInfo{
			Model:      "Category",
			Collection: "Category",
			Attribute:  "_id",
			Type:       "id",
			IsRequired: false,
			IsStatic:   true,
			Label:      "ID",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Category",
			Collection: "Category",
			Attribute:  "name",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Name",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Category",
			Collection: "Category",
			Attribute:  "parent_id",
			Type:       "id",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Parent",
			Group:      "General",
			Editors:    "category_selector",
			Options:    "",
			Default:    "",
		},
		models.T_AttributeInfo{
			Model:      "Category",
			Collection: "Category",
			Attribute:  "products",
			Type:       "id",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Products",
			Group:      "General",
			Editors:    "product_selector",
			Options:    "",
			Default:    "",
		},
	}

	return info
}
