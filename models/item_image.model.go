package models

type ItemImageContent struct {
	BasedId  uint   `json:"based_id"`
	Image    string `json:"image"`
	Filename string `json:"filename"`
}

type ItemImage struct {
	ID uint `gorm:"primarykey" json:"id"`
	ItemImageContent
}

func (ItemImage) TableName() string {
	return "tbl_setup_item_image"
}

type ItemImageAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	ItemImageContent
	At
}

func (ItemImageAt) TableName() string {
	return "z_tbl_setup_item_image_at"
}
