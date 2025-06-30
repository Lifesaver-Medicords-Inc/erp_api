package sales_invoices_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/journal_entry_services.go"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type Body struct {
	accounting_models.SalesInvoice
	SalesInvoiceDetails []accounting_models.SalesInvoiceDetail `json:"sales_invoice_details"`
}

func GetSalesInvoices(conditions map[string]interface{}) (interface{}, int, error) {

	return "", 0, nil
}

func GetSalesInvoiceDocNo(conditions map[string]interface{}) (interface{}, int, error) {

	var response int
	if err := initializers.DB.Table("tbl_accounting_sales_invoice").
		Select("MAX(id)").
		Scan(&response).Error; err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get sales invoice doc_no")
	}

	return response, 0, nil
}

func CreateSalesInvoice(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body
	// var JournalRecords struct {
	// 	accounting_models.JournalEntry
	// 	Details []accounting_models.JournalEntryDetails
	// }

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.SalesInvoice); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to insert sales invoice header")
	}

	at, ok := c.Locals("at").(models.At)
	userAt := utils.GetAtData(c, models.At{})
	at.AtUserId = userAt.AtUserId
	if !ok {
		at = models.At{}
	}

	parentChildAt := accounting_models.SalesInvoiceAt{RefId: body.ID, SalesInvoiceContent: body.SalesInvoiceContent, At: at}
	if err := services.DbInsert(tx, &parentChildAt); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating parentat")
	}

	if err := CreateSalesInvoiceDetails(tx, &body.SalesInvoiceDetails, body.ID, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	selectedJournalData := journal_entry_services.FilterJournalRecords{
		DocDate:          body.DocDate,
		PostingDate:      body.DocDate,
		Origin:           "SALES INVOICE",
		OriginNo:         body.ID,
		DocNo:            body.DocNo,
		JournalId:        body.JournalId,
		TotalAmountDue:   body.TotalAmountDue,
		TotalAmountDueFc: body.TotalAmountDueFc,
		AmountDue:        body.AmountDue,
		AmountDueFc:      body.AmountDueFc,
		AddVat:           body.AddVat,
		AddVatFc:         body.AddVatFc,
		TaxName:          body.TaxCode,
	}
	fmt.Println(selectedJournalData)

	if err := journal_entry_services.CreateAutoEntry(tx, selectedJournalData, body.ID); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// JournalRecords.DocDate = body.DocDate
	// JournalRecords.PostingDate = body.DocDate // this is posting date
	// JournalRecords.Origin = "SALES INVOICE"
	// JournalRecords.OriginNo = body.ID
	// JournalRecords.DocNo = body.DocNo
	// JournalRecords.TranNo = body.DocSoNo

	// if err := services.DbInsert(tx, JournalRecords.JournalEntry); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	// journalDebit := journal_entry_services.GenerateAutoEntry("DEBIT", body.TotalAmountDue, body.TotalAmountDueFc, 1)
	// journalCredit := journal_entry_services.GenerateAutoEntry("CREDIT", body.AmountDue, body.AmountDueFc, body.JournalId)

	// if body.JournalName == "VAT" {
	// 	journalVatCredit := journal_entry_services.GenerateAutoEntry("CREDIT", body.AddVat, body.AddVatFc, body.JournalId)
	// 	JournalRecords.Details = append(JournalRecords.Details, journalVatCredit)
	// }
	// JournalRecords.Details = append(JournalRecords.Details, journalDebit, journalCredit)

	// if err := journal_entry_services.CreateEntries(tx, JournalRecords.Details, body.ID); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }

	return body, 0, nil
}
