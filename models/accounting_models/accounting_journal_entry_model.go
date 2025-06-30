package accounting_models

import "github.com/pierceperado/smpc/models"

type JournalEntryContent struct {
	TranNo      string `json:"trans_no"`
	DocNo       string `json:"doc_no"`
	PostingDate string `json:"posting_date"`
	DueDate     string `json:"due_date"`
	DocDate     string `json:"doc_date"`
	Remarks     string `json:"remarks"`
	Origin      string `json:"origin"`
	OriginNo    uint   `json:"origin_no"`
	CreatedById uint   `json:"created_by_id"`
}
type JournalEntry struct {
	ID uint `gorm:"primarykey" json:"id"`
	JournalEntryContent
}

func (JournalEntry) TableName() string {
	return "tbl_accounting_journal_entry"
}

type JournalEntryAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	JournalEntryContent
	models.At
}

func (JournalEntryAt) TableName() string {
	return "z_tbl_accounting_journal_entry_at"
}
