package accounting_models

import "github.com/pierceperado/smpc/models"

type JournalEntryDetailsContent struct {
	JournalEntryId uint    `json:"journal_entry_id"`
	AccountID      uint    `json:"account_id"`
	Debit          float64 `json:"debit"`
	DebitBased     float64 `json:"debit_based"`
	Credit         float64 `json:"credit"`
	CreditBased    float64 `json:"credit_based"`
	Remarks        string  `json:"remarks"`
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
