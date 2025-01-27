package models

type BpiContactContent struct {
	BasedId         uint   `json:"based_id"`
	TransactionType string `json:"transaction_type"`
	Name            string `json:"name"`
	Email           string `json:"Email"`
	Position        int    `json:"Position"`
}

type BpiContacts struct {
	ID uint `gorm:"primarykey" json:"id"`
	BpiContactContent
}

func (BpiContacts) TableName() string {
	return "tbl_bpi_contacts"
}

type BpiContactsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	BpiContactContent
	At
}

func (BpiContactsAt) TableName() string {
	return "z_tbl_bpi_contacts_at"
}
