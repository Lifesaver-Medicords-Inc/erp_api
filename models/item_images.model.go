package models

type ItemImageContent struct {
	BasedId uint   `json:"based_id"`
	Img1    string `json:"img1"`
	Img2    string `json:"img2"`
	Img3    string `json:"img3"`
	Img4    string `json:"img4"`
	Img5    string `json:"img5"`
	Img6    string `json:"img6"`
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