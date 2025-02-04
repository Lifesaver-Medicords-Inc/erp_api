package models

type ModelView struct {
	Model
	RelatedName  string `json:"related_name"`
	RelatedBrand string `json:"related_brand"`
}

func (ModelView) TableName() string {
	return "ItemModelList"
}
