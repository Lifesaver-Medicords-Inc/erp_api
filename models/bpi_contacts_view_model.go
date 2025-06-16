package models

type BpiViewContactsContent struct {
	ContactsBasedId  uint   `json:"contacts_based_id"`
	BranchId         uint   `json:"branch_id"`
	Number           string `json:"number"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Preferences      string `json:"preferences"`
	Position         uint   `json:"position"`
	IsDefaultContact bool   `json:"is_default_contact"`
	ContactNotes     string `json:"contact_notes"`
}

type BpiContactView struct {
	ContactsID uint `gorm:"primarykey" json:"contacts_id"`
	BpiViewContactsContent
}

func (BpiContactView) TableName() string {
	return "vw_get_bpi_contacts"
}
