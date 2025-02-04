package models

type BpiAddressViewContent struct {
	AddressBasedId uint   `json:"address_based_id"`
	Location       string `json:"location"`
}

type BpiAddressView struct {
	AddressID uint `gorm:"primarykey" json:"address_id"`
	BpiAddressViewContent
}

func (BpiAddressView) TableName() string {
	return "GetBpiAddress"
}
