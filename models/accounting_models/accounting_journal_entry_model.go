package accounting_models

import "github.com/pierceperado/smpc/models"

type JournalEntryContent struct {
	JournalName        string `json:"journal_name"`
	JournalDescription string `json:"journal_description"`
	DocNo              int    `json:"doc_no"`
	Period             string `json:"period"`
	PeriodFrom         string `json:"period_from"`
	PeriodTo           string `json:"period_to"`
	Currency           string `json:"currency"`
	CreatedBy          string `json:"created_by"`
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

type JournalEntryCurrent struct {
	ID                 uint   `json:"id"`
	JournalName        string `json:"journal_name"`
	JournalDescription string `json:"journal_description"`
	DocNo              string `json:"doc_no"`
	Period             string `json:"period"`
	PeriodFrom         string `json:"period_from"`
	PeriodTo           string `json:"period_to"`
	Currency           string `json:"currency"`
	CreatedBy          string `json:"created_by"`
}

func (JournalEntryCurrent) TableName() string {
	return "tbl_accounting_journal_entry"
}

type JournalEntryDetailsContent struct {
	JournalEntryId uint    `json:"journal_entry_id"`
	InsertedDate   string  `json:"inserted_date"`
	PostingDate    string  `json:"posting_date"`
	AccountTitle   string  `json:"account_title"`
	PostingRef     string  `json:"posting_ref"`
	PostingRefId   uint    `json:"posting_ref_id"`
	Origin         string  `json:"origin"`
	OriginId       uint    `json:"origin_id"`
	Debit          float64 `json:"debit"`
	Credit         float64 `json:"credit"`
	LineMemo       string  `json:"line_memo"`
	CreatedBy      string  `json:"created_by"`
}

type JournalEntryDetails struct {
	ID uint `gorm:"primarykey" json:"id"`
	JournalEntryDetailsContent
}

func (JournalEntryDetails) TableName() string {
	return "tbl_accounting_journal_entry_details"
}

type JournalEntryDetailsAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	JournalEntryDetailsContent
	models.At
}

func (JournalEntryDetailsAt) TableName() string {
	return "z_tbl_accounting_journal_entry_details_at"
}

type JournalEntryBody struct {
	JournalEntry        JournalEntry          `json:"journal_entry"`
	JournalEntryDetails []JournalEntryDetails `json:"journal_entry_details"`
}

type JournalEntryGet struct {
	JournalEntry        []JournalEntry        `json:"journal_entry"`
	JournalEntryDetails []JournalEntryDetails `json:"journal_entry_details"`
}
