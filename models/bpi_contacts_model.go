package models

type BpiContactContent struct {
	BasedId          uint   `json:"based_id"`
	BranchId         uint   `json:"branch_id"`
	Number           string `json:"number"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Preferences      string `json:"preferences"`
	Position         uint   `json:"position"`
	IsDefaultContact bool   `json:"is_default_contact"`
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
