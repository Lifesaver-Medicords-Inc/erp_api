package accounting_models

import "github.com/pierceperado/smpc/models"

type JournalEntry2Content struct {
	JournalName        string `json:"journal_name"`
	JournalDescription string `json:"journal_description"`
	DocNo              int    `json:"doc_no"`
	Period             string `json:"period"`
	PeriodFrom         string `json:"period_from"`
	PeriodTo           string `json:"period_to"`
	Currency           string `json:"currency"`
	CreatedBy          string `json:"created_by"`
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
	return "tbl_accounting_journal_entry2"
}

type JournalEntryDetails2Content struct {
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

type JournalEntryBody struct {
	JournalEntry        JournalEntry2          `json:"journal_entry"`
	JournalEntryDetails []JournalEntryDetails2 `json:"journal_entry_details"`
}

type JournalEntryGet struct {
	JournalEntry        []JournalEntry2        `json:"journal_entry"`
	JournalEntryDetails []JournalEntryDetails2 `json:"journal_entry_details"`
}
