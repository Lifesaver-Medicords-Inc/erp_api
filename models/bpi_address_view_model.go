package models

type BpiAddressViewContent struct {
	AddressBasedId  uint   `json:"address_based_id"`
	AddressBranchId uint   `json:"address_branch_id"`
	Location        string `json:"location"`
}

type BpiAddressView struct {
	AddressIds uint `gorm:"primarykey" json:"address_ids"`
	BpiAddressViewContent
}

func (BpiAddressView) TableName() string {
	return "vw_get_bpi_adddress"
}
