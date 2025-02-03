package models

type BpiViewContactsContent struct {
	ContactsBasedId uint   `json:"contacts_based_id"`
	Number          string `json:"number"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Preferences     string `json:"preferences"`
	Position        uint   `json:"position"`
}

type BpiViewContacts struct {
	ContactsID uint `gorm:"primarykey" json:"contacts_id"`
	BpiViewContactsContent
}

func (BpiViewContacts) TableName() string {
	return "GetBpiContacts"
}
