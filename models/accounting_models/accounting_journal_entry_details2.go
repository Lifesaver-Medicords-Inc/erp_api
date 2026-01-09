package accounting_models

import "github.com/pierceperado/smpc/models"

type JournalEntryDetails2Content struct {
	JournalEntryId uint    `json:"journal_entry_id"`
	InsertedDate   string  `json:"inserted_date"`
	PostingDate    string  `json:"posting_date"`
	AccountTitle   string  `json:"account_title"`
	PostingRef     string  `json:"posting_ref"`
	PostingRefId   uint    `json:"posting_ref_id"`
	Origin         string  `json:"origin"`
	OriginNo       uint    `json:"origin_no"`
	Debit          float64 `json:"debit"`
	Credit         float64 `json:"credit"`
	LineMemo       string  `json:"line_memo"`
	CreatedBy      string  `json:"created_by"`
}

type JournalEntryDetails2 struct {
	ID uint `gorm:"primarykey" json:"id"`
	JournalEntryDetails2Content
}

func (JournalEntryDetails2) TableName() string {
	return "tbl_accounting_journal_entry_details2"
}

type JournalEntryDetails2At struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	JournalEntryDetails2Content
	models.At
}

func (JournalEntryDetails2At) TableName() string {
	return "z_tbl_accounting_journal_entry_details2_at"
}
