package payment_receipt_services

import (
	// "errors"

	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/journal_entry_services2"
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

	// Fetch debit and credit COAs
	var coaDEBIT, coaCREDIT accounting_models.ChartOfAccounts
	if err := tx.First(&coaDEBIT, 40029).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching debit chart of account")
	}
	if err := tx.First(&coaCREDIT, 40030).Error; err != nil {
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

	// Auto-insert debit and credit
	for _, e := range []accounting_models.JournalEntryDetails2{
		createJournalEntry(coaCREDIT, body.PaymentReceipt.TransactionAmount, true),
		createJournalEntry(coaDEBIT, body.PaymentReceipt.TransactionAmount, false),
	} {
		if err := jeService.AutoInsertJournalEntry(&e, body.PaymentReceipt.DocDate, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
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
