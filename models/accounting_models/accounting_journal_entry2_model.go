package accounting_models

import "github.com/pierceperado/smpc/models"

type JournalEntry2Content struct {
	JournalName        string `json:"journal_name"`
	JournalDescription string `json:"journal_description"`
	DocNo              string `json:"doc_no"`
	Period             string `json:"period"`
	Currency           string `json:"currency"`
	Origin             string `json:"origin"`
	OriginNo           uint   `json:"origin_no"`
	CreatedById        uint   `json:"created_by_id"`
}
type JournalEntry2 struct {
	ID uint `gorm:"primarykey" json:"id"`
	JournalEntry2Content
}

func (JournalEntry2) TableName() string {
	return "tbl_accounting_journal_entry2"
}

type JournalEntry2At struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	JournalEntry2Content
	models.At
}

func (JournalEntry2At) TableName() string {
	return "z_tbl_accounting_journal_entry2_at"
}
