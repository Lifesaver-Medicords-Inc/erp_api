package payment_receipt_services

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
	"github.com/pierceperado/smpc/services/overpayment_services"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type PaymentReceiptService struct{}

func NewPaymentReceiptService() *PaymentReceiptService {
	return &PaymentReceiptService{}
}

func (s *PaymentReceiptService) GetPaymentReceipt(conditions map[string]interface{}) (interface{}, int, error) {
	var response accounting_models.PaymentReceiptGet

	if err := services.DbGet(&response.PaymentReceipt, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting payment receipt")
	}

	if err := services.DbGet(&response.PaymentReceiptDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting payment receipt details")
	}

	return response, fiber.StatusOK, nil
}

func (s *PaymentReceiptService) GetCustomerSalesInvoice(conditions map[string]interface{}) (interface{}, int, error) {
	var response []accounting_models.SalesInvoiceReceiptView

	// Get Sales Invoice (Parent)
	if err := services.DbRaw(&response, "sp_GetSIPaymentReceipt", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting customer sales invoice data")
	}

	if len(response) == 0 {
		return nil, fiber.StatusNotFound, errors.New("no sales invoice found")
	}

	return response, fiber.StatusOK, nil
}

func (s *PaymentReceiptService) CreatePaymentReceipt(body *accounting_models.PaymentReceiptBody, at models.At) (*accounting_models.PaymentReceiptBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback() // rollback unless committed

	// Insert main Payment Receipt
	if err := services.DbInsert(tx, &body.PaymentReceipt); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating payment receipt")
	}

	// Generate and update DocNo
	body.PaymentReceipt.DocNo = utils.DocNoGenerator(body.PaymentReceipt.ID)
	if err := tx.Model(&body.PaymentReceipt).Update("doc_no", body.PaymentReceipt.DocNo).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating payment receipt doc")
	}

	// Insert Payment Receipt Details
	if err := s.CreatePaymentReceiptDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Find matching journal entry
	var journals []accounting_models.JournalEntry2
	if err := tx.Find(&journals).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching journal entries")
	}

	docDate, err := time.Parse("01/02/2006", strings.ToLower(strings.TrimSpace(body.PaymentReceipt.DocDate)))
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("invalid payment receipt doc date format")
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
		return body, fiber.StatusNotFound, errors.New("no journal entry found for the payment receipt period")
	}

	debitCOAId := s.getDebitCOAId(body)

	// Fetch debit and credit COAs
	var coaDEBIT, coaCREDIT accounting_models.ChartOfAccounts

	if err := tx.First(&coaDEBIT, debitCOAId).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching debit chart of account")
	}
	if err := tx.First(&coaCREDIT, 70037).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching credit chart of account")
	}

	// Helper to create journal entry
	createJournalEntry := func(account accounting_models.ChartOfAccounts, amount float64, isCredit bool) accounting_models.JournalEntryDetails2 {
		entry := accounting_models.JournalEntryDetails2{
			JournalEntryDetails2Content: accounting_models.JournalEntryDetails2Content{
				Origin:         "Payment Receipt",
				OriginId:       body.PaymentReceipt.ID,
				LineMemo:       "Auto entry - " + map[bool]string{true: "CREDIT", false: "DEBIT"}[isCredit],
				PostingDate:    body.PaymentReceipt.DocDate,
				CreatedBy:      body.PaymentReceipt.PreparedBy,
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

	// Compute total applied from details
	var totalApplied float64
	for _, d := range body.PaymentReceiptDetails {
		totalApplied += d.AmountApplied
	}

	// Auto-insert debit and credit
	for _, e := range []accounting_models.JournalEntryDetails2{
		createJournalEntry(coaCREDIT, body.PaymentReceipt.TransactionAmount, true),
		createJournalEntry(coaDEBIT, totalApplied, false),
	} {
		if err := jeService.AutoInsertJournalEntry(&e, body.PaymentReceipt.DocDate, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//Insert difference debit if needed
	if err := s.InsertDifferenceDebit(tx, body, matchedJournalID, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Insert audit record
	atdata := accounting_models.PaymentReceiptAt{
		RefId:                 body.PaymentReceipt.ID,
		PaymentReceiptContent: body.PaymentReceipt.PaymentReceiptContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating payment receipt at")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(accounting_models.SalesInvoiceReceiptView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	setup_services.InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *PaymentReceiptService) CreatePaymentReceiptDetails(tx *gorm.DB, body *accounting_models.PaymentReceiptBody, at models.At) error {
	for i := range body.PaymentReceiptDetails {
		detail := &body.PaymentReceiptDetails[i]
		detail.PaymentReceiptID = body.PaymentReceipt.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating payment receipt details")
		}

		// Audit trail for each detail
		atdataDetail := accounting_models.PaymentReceiptDetailsAt{
			RefId:                        detail.ID,
			PaymentReceiptDetailsContent: detail.PaymentReceiptDetailsContent,
			At:                           at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating payment receipt details at")
		}
	}
	return nil
}

func (s *PaymentReceiptService) InsertDifferenceDebit(tx *gorm.DB, body *accounting_models.PaymentReceiptBody, matchedJournalID uint, at models.At) error {

	//Compute total applied from details
	var totalApplied float64
	for _, d := range body.PaymentReceiptDetails {
		totalApplied += d.AmountApplied
	}

	transactionAmount := body.PaymentReceipt.TransactionAmount

	//If transaction amount is greater than applied
	if transactionAmount > totalApplied {

		difference := transactionAmount - totalApplied

		//Fetch COA 60030
		var coaDifference accounting_models.ChartOfAccounts
		if err := tx.First(&coaDifference, 60030).Error; err != nil {
			return errors.New("failed fetching difference chart of account (60030)")
		}

		//Create DEBIT journal entry
		entry := accounting_models.JournalEntryDetails2{
			JournalEntryDetails2Content: accounting_models.JournalEntryDetails2Content{
				Origin:         "Payment Receipt",
				OriginId:       body.PaymentReceipt.ID,
				LineMemo:       "Auto entry - DEBIT",
				PostingDate:    body.PaymentReceipt.DocDate,
				CreatedBy:      body.PaymentReceipt.PreparedBy,
				AccountTitle:   coaDifference.Name,
				PostingRef:     coaDifference.Code,
				PostingRefId:   coaDifference.ID,
				JournalEntryId: matchedJournalID,
				Debit:          difference,
			},
		}

		//Create overpayment record
		overpayment := accounting_models.BpiOverpayment{
			BpiOverpaymentContent: accounting_models.BpiOverpaymentContent{
				BpiId:             body.PaymentReceipt.CustomerId,
				OverpaymentAmount: difference,
				ReferenceDate:     body.PaymentReceipt.DocDate,
				ReferenceDocType:  "Payment Receipt",
				ReferenceDocId:    body.PaymentReceipt.ID,
			},
		}

		overService := overpayment_services.NewOverpaymentService()

		if err := overService.CreateBpiOverpayment(tx, &overpayment, at); err != nil {
			return err
		}

		jeService := journal_entry_services2.NewJournalEntryService2()

		if err := jeService.AutoInsertJournalEntry(&entry, body.PaymentReceipt.DocDate, at); err != nil {
			return err
		}
	}

	return nil
}

func (s *PaymentReceiptService) getDebitCOAId(body *accounting_models.PaymentReceiptBody) uint {

	// If CashAmount is greater than 0 (and not zero)
	if body.PaymentReceipt.CashAmount > 0 {
		return 70034
	}

	// If zero or empty (default float64 is 0)
	return 70035
}
