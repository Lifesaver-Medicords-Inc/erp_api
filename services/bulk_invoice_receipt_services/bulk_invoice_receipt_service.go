package bulk_invoice_receipt_services

import (
	// "errors"

	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/journal_entry_services2"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type BulkInvoiceReceiptService struct{}

func NewBulkInvoiceReceiptService() *BulkInvoiceReceiptService {
	return &BulkInvoiceReceiptService{}
}

func (s *BulkInvoiceReceiptService) GetBulkInvoiceReceipt(conditions map[string]interface{}) (interface{}, int, error) {
	var response accounting_models.BulkInvoiceReceiptGet

	if err := services.DbGet(&response.BulkInvoiceReceipt, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting bulk invoice receipt")
	}

	if err := services.DbGet(&response.BulkInvoiceReceiptDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting bulk invoice receipt details")
	}

	return response, fiber.StatusOK, nil
}

func (s *BulkInvoiceReceiptService) CreateBulkInvoiceReceipt(body *accounting_models.BulkInvoiceReceiptBody, at models.At) (*accounting_models.BulkInvoiceReceiptBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback() // rollback unless committed

	// Insert main Bulk Invoice Receipt
	if err := services.DbInsert(tx, &body.BulkInvoiceReceipt); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating bulk invoice receipt")
	}

	// Generate and update DocNo
	body.BulkInvoiceReceipt.DocNo = utils.DocNoGenerator(body.BulkInvoiceReceipt.ID)
	if err := tx.Model(&body.BulkInvoiceReceipt).Update("doc_no", body.BulkInvoiceReceipt.DocNo).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating bulk invoice receipt doc")
	}

	// Insert Bulk Invoice Receipt Details
	if err := s.CreateBulkInvoiceReceiptDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Find matching journal entry
	var journals []accounting_models.JournalEntry2
	if err := tx.Find(&journals).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching journal entries")
	}

	docDate, err := time.Parse("01/02/2006", strings.ToLower(strings.TrimSpace(body.BulkInvoiceReceipt.DocDate)))
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("invalid bulk invoice receipt doc date format")
	}

	var matchedJournalID uint
	for _, j := range journals {
		parts := strings.Split(j.Period, " to ")
		if len(parts) != 2 {
			continue
		}
		startDate, err1 := time.Parse("02/01/2006 3:04:05 pm", strings.TrimSpace(parts[0]))
		endDate, err2 := time.Parse("02/01/2006 3:04:05 pm", strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			continue
		}
		if !docDate.Before(startDate) && !docDate.After(endDate) {
			matchedJournalID = j.ID
			break
		}
	}
	if matchedJournalID == 0 {
		return body, fiber.StatusNotFound, errors.New("no journal entry found for the bulk invoice period")
	}

	// Fetch debit and credit COAs
	var coaDEBIT, coaCREDIT accounting_models.ChartOfAccounts
	if err := tx.First(&coaDEBIT, 70036).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching debit chart of account")
	}
	if err := tx.First(&coaCREDIT, 50030).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching credit chart of account")
	}

	// Helper to create journal entry
	createJournalEntry := func(account accounting_models.ChartOfAccounts, amount float64, isCredit bool) accounting_models.JournalEntryDetails2 {
		entry := accounting_models.JournalEntryDetails2{
			JournalEntryDetails2Content: accounting_models.JournalEntryDetails2Content{
				Origin:         "Bulk Invoice Receipt",
				OriginId:       body.BulkInvoiceReceipt.ID,
				LineMemo:       "Auto entry - " + map[bool]string{true: "CREDIT", false: "DEBIT"}[isCredit],
				PostingDate:    body.BulkInvoiceReceipt.DocDate,
				CreatedBy:      body.BulkInvoiceReceipt.PreparedBy,
				AccountTitle:   account.Name,
				PostingRef:     account.Code,
				PostingRefId:   account.ID,
				JournalEntryId: matchedJournalID,
			},
		}
		if isCredit {
			entry.Credit = amount
		} else {
			entry.Debit = amount
		}
		return entry
	}

	jeService := journal_entry_services2.NewJournalEntryService2()

	// Auto-insert debit and credit
	for _, e := range []accounting_models.JournalEntryDetails2{
		createJournalEntry(coaCREDIT, body.BulkInvoiceReceipt.NetAmount, true),
		createJournalEntry(coaDEBIT, body.BulkInvoiceReceipt.NetAmount, false),
	} {
		if err := jeService.AutoInsertJournalEntry(&e, body.BulkInvoiceReceipt.DocDate, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	// Insert audit record
	atdata := accounting_models.BulkInvoiceReceiptAt{
		RefId:                     body.BulkInvoiceReceipt.ID,
		BulkInvoiceReceiptContent: body.BulkInvoiceReceipt.BulkInvoiceReceiptContent,
		At:                        at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating bulk invoice receipt at")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(accounting_models.InvoicePOView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(accounting_models.InvoicePODetailView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(accounting_models.InvoiceReceiptView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, fiber.StatusOK, nil
}

func (s *BulkInvoiceReceiptService) CreateBulkInvoiceReceiptDetails(tx *gorm.DB, body *accounting_models.BulkInvoiceReceiptBody, at models.At) error {
	for i := range body.BulkInvoiceReceiptDetails {
		detail := &body.BulkInvoiceReceiptDetails[i]
		detail.BulkInvoiceReceiptID = body.BulkInvoiceReceipt.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating bulk invoice receipt details")
		}

		// Audit trail for each detail
		atdataDetail := accounting_models.BulkInvoiceReceiptDetailsAt{
			RefId:                            detail.ID,
			BulkInvoiceReceiptDetailsContent: detail.BulkInvoiceReceiptDetailsContent,
			At:                               at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating bulk invoice receipt details at")
		}
	}
	return nil
}
