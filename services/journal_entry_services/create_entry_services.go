package journal_entry_services

import (
	"errors"

	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type FilterJournalRecords struct {
	DocDate          string
	PostingDate      string
	Origin           string
	OriginNo         uint
	DocNo            string
	JournalId        uint
	TotalAmountDue   float64
	TotalAmountDueFc float64
	AmountDue        float64
	AmountDueFc      float64
	AddVat           float64
	AddVatFc         float64
	TaxName          string
	TaxId            uint
}

func CreateEntries(tx *gorm.DB, child []accounting_models.JournalEntryDetails, parentId uint) error {

	for _, v := range child {
		if err := CreateJournalEntry(tx, v, parentId); err != nil {
			return err
		}
	}
	return nil
}

func CreateJournalEntry(tx *gorm.DB, child accounting_models.JournalEntryDetails, parentId uint) error {
	child.JournalEntryDetailsContent.JournalEntryId = parentId

	if err := services.DbInsert(tx, &child); err != nil {
		return errors.New("failed to insert journal entry details")
	}

	return nil
}

func CreateAutoEntry(tx *gorm.DB, body FilterJournalRecords, parentId uint) error {

	var JournalRecords struct {
		accounting_models.JournalEntry
		Details []accounting_models.JournalEntryDetails
	}
	JournalRecords.DocDate = body.DocDate
	JournalRecords.PostingDate = body.DocDate // this is posting date
	JournalRecords.Origin = "SALES INVOICE"
	JournalRecords.OriginNo = parentId
	JournalRecords.DocNo = body.DocNo
	JournalRecords.TranNo = body.DocNo

	if err := services.DbInsert(tx, &JournalRecords.JournalEntry); err != nil {
		return errors.New("failed to insert auto entry journal header")
	}

	journalDebit := GenerateAutoEntry("DEBIT", body.TotalAmountDue, body.TotalAmountDueFc, 1)
	journalCredit := GenerateAutoEntry("CREDIT", body.AmountDue, body.AmountDueFc, body.JournalId)
	if body.TaxName == "VAT" {
		journalVatCredit := GenerateAutoEntry("CREDIT", body.AddVat, body.AddVatFc, body.TaxId)
		JournalRecords.Details = append(JournalRecords.Details, journalVatCredit)
	}
	JournalRecords.Details = append(JournalRecords.Details, journalDebit, journalCredit)

	if err := CreateEntries(tx, JournalRecords.Details, parentId); err != nil {
		return err
	}
	return nil
}

func GenerateAutoEntry(entryType string, amount float64, amountFc float64, accountID uint) accounting_models.JournalEntryDetails {
	var records accounting_models.JournalEntryDetails

	switch entryType {
	case "DEBIT":
		records = accounting_models.JournalEntryDetails{
			JournalEntryDetailsContent: accounting_models.JournalEntryDetailsContent{
				AccountID:   accountID,
				Debit:       amount,
				DebitBased:  amountFc,
				Credit:      0.00,
				CreditBased: 0.00,
				Remarks:     "Auto entry - DEBIT",
			},
		}
	case "CREDIT":
		records = accounting_models.JournalEntryDetails{
			JournalEntryDetailsContent: accounting_models.JournalEntryDetailsContent{
				AccountID:   accountID,
				Debit:       0.00,
				DebitBased:  0.00,
				Credit:      amount,
				CreditBased: amountFc,
				Remarks:     "Auto entry - CREDIT",
			},
		}
	}

	return records
}
