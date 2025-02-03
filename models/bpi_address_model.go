package models

type BpiAddressContent struct {
	BasedId  uint   `json:"based_id"`
	Location string `json:"location"`
}

type BpiAddress struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiAddressContent
}

func (BpiAddress) TableName() string {
	return "tbl_bpi_address"
}

type BpiAddressAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiAddressContent
	At
}

func (BpiAddressAt) TableName() string {
	return "z_tbl_bpi_address_at"
}
