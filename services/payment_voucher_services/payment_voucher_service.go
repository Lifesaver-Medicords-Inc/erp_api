package payment_voucher_services

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

type PaymentVoucherService struct{}

func NewPaymentVoucherService() *PaymentVoucherService {
	return &PaymentVoucherService{}
}

func (s *PaymentVoucherService) GetPaymentVoucher(conditions map[string]interface{}) (interface{}, int, error) {
	var response accounting_models.PaymentVoucherGet

	if err := services.DbGet(&response.PaymentVoucher, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting payment voucher")
	}

	if err := services.DbGet(&response.PaymentVoucherDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting payment voucher details")
	}

	return response, fiber.StatusOK, nil
}

func (s *PaymentVoucherService) GetSupplierAPVoucher(conditions map[string]interface{}) (interface{}, int, error) {
	var apParent []accounting_models.APVoucherPaymentView

	// Get AP Voucher (Parent)
	if err := services.DbRaw(&apParent, "sp_GetAPVoucherPayment", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting ap voucher data")
	}

	if len(apParent) == 0 {
		return nil, fiber.StatusNotFound, errors.New("no ap voucher found")
	}

	var allChildren []accounting_models.APVoucherPaymentDetailsView

	for _, ap := range apParent {
		var poChild []accounting_models.APVoucherPaymentDetailsView

		childConditions := map[string]interface{}{
			"VoucherId": ap.APVoucherId,
		}

		if err := services.DbRaw(&poChild, "sp_GetAPVoucherPaymentDetails", childConditions); err != nil {
			return nil, fiber.StatusInternalServerError, errors.New("failed getting ap voucher details data")
		}

		// Append children to one slice
		allChildren = append(allChildren, poChild...)
	}

	// Response structure
	response := struct {
		APVoucherView        []accounting_models.APVoucherPaymentView        `json:"ap_voucher_view"`
		APVoucherDetailsView []accounting_models.APVoucherPaymentDetailsView `json:"ap_voucher_details_view"`
	}{
		APVoucherView:        apParent,
		APVoucherDetailsView: allChildren,
	}

	return response, fiber.StatusOK, nil
}

func (s *PaymentVoucherService) CreatePaymentVoucher(body *accounting_models.PaymentVoucherBody, at models.At) (*accounting_models.PaymentVoucherBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback() // rollback unless committed

	nextDocNo, err := utils.NextDocNo(tx, new(accounting_models.PaymentVoucher), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	body.PaymentVoucher.DocNo = nextDocNo

	// Insert main Payment Voucher
	if err := services.DbInsert(tx, &body.PaymentVoucher); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating payment voucher")
	}

	// Insert Payment Voucher Details
	if err := s.CreatePaymentVoucherDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Find matching journal entry
	var journals []accounting_models.JournalEntry2
	if err := tx.Find(&journals).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching journal entries")
	}

	docDate, err := time.Parse("01/02/2006", strings.ToLower(strings.TrimSpace(body.PaymentVoucher.DocDate)))
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("invalid payment voucher doc date format")
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
		return body, fiber.StatusNotFound, errors.New("no journal entry found for the payment voucher period")
	}

	// Fetch debit and credit COAs
	var coaDEBIT, coaCREDIT accounting_models.ChartOfAccounts

	crebitCOAId := s.getCrebitCOAId(body)

	if err := tx.First(&coaDEBIT, 40030).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching debit chart of account")
	}
	if err := tx.First(&coaCREDIT, crebitCOAId).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching credit chart of account")
	}

	// Helper to create journal entry
	createJournalEntry := func(account accounting_models.ChartOfAccounts, amount float64, isCredit bool) accounting_models.JournalEntryDetails2 {
		entry := accounting_models.JournalEntryDetails2{
			JournalEntryDetails2Content: accounting_models.JournalEntryDetails2Content{
				Origin:         "Payment Voucher",
				OriginId:       body.PaymentVoucher.ID,
				LineMemo:       "Auto entry - " + map[bool]string{true: "CREDIT", false: "DEBIT"}[isCredit],
				PostingDate:    body.PaymentVoucher.DocDate,
				CreatedBy:      body.PaymentVoucher.PreparedBy,
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
	for _, d := range body.PaymentVoucherDetails {
		totalApplied += d.AmountApplied
	}

	// Auto-insert debit and credit
	for _, e := range []accounting_models.JournalEntryDetails2{
		createJournalEntry(coaCREDIT, body.PaymentVoucher.TransactionAmount, true),
		createJournalEntry(coaDEBIT, totalApplied, false),
	} {
		if err := jeService.AutoInsertJournalEntry(&e, body.PaymentVoucher.DocDate, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//Insert difference debit if needed
	if err := s.InsertDifferenceDebit(tx, body, matchedJournalID, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Insert audit record
	atdata := accounting_models.PaymentVoucherAt{
		RefId:                 body.PaymentVoucher.ID,
		PaymentVoucherContent: body.PaymentVoucher.PaymentVoucherContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating payment voucher at")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	if err := services.InvalidateCacheByModel(accounting_models.APVoucherPaymentView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(accounting_models.APVoucherPaymentDetailsView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	setup_services.InvalidateItemCaches()

	return body, fiber.StatusOK, nil
}

func (s *PaymentVoucherService) CreatePaymentVoucherDetails(tx *gorm.DB, body *accounting_models.PaymentVoucherBody, at models.At) error {
	for i := range body.PaymentVoucherDetails {
		detail := &body.PaymentVoucherDetails[i]
		detail.PaymentVoucherID = body.PaymentVoucher.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating payment voucher details")
		}

		// Audit trail for each detail
		atdataDetail := accounting_models.PaymentVoucherDetailsAt{
			RefId:                        detail.ID,
			PaymentVoucherDetailsContent: detail.PaymentVoucherDetailsContent,
			At:                           at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating payment voucher details at")
		}
	}
	return nil
}

func (s *PaymentVoucherService) InsertDifferenceDebit(tx *gorm.DB, body *accounting_models.PaymentVoucherBody, matchedJournalID uint, at models.At) error {

	//Compute total applied from details
	var totalApplied float64
	for _, d := range body.PaymentVoucherDetails {
		totalApplied += d.AmountApplied
	}

	transactionAmount := body.PaymentVoucher.TransactionAmount

	//If transaction amount is greater than applied
	if transactionAmount > totalApplied {

		difference := transactionAmount - totalApplied

		//Fetch COA 70038
		var coaDifference accounting_models.ChartOfAccounts
		if err := tx.First(&coaDifference, 70038).Error; err != nil {
			return errors.New("failed fetching difference chart of account (70038)")
		}

		//Create DEBIT journal entry
		entry := accounting_models.JournalEntryDetails2{
			JournalEntryDetails2Content: accounting_models.JournalEntryDetails2Content{
				Origin:         "Payment Voucher",
				OriginId:       body.PaymentVoucher.ID,
				LineMemo:       "Auto entry - DEBIT",
				PostingDate:    body.PaymentVoucher.DocDate,
				CreatedBy:      body.PaymentVoucher.PreparedBy,
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
				BpiId:                    body.PaymentVoucher.SupplierId,
				CompanyOverpaymentAmount: difference,
				ReferenceDate:            body.PaymentVoucher.DocDate,
				ReferenceDocType:         "Payment Voucher",
				ReferenceDocId:           body.PaymentVoucher.ID,
			},
		}

		overService := overpayment_services.NewOverpaymentService()

		if err := overService.CreateBpiOverpayment(tx, &overpayment, at); err != nil {
			return err
		}

		jeService := journal_entry_services2.NewJournalEntryService2()

		if err := jeService.AutoInsertJournalEntry(&entry, body.PaymentVoucher.DocDate, at); err != nil {
			return err
		}
	}

	return nil
}

func (s *PaymentVoucherService) getCrebitCOAId(body *accounting_models.PaymentVoucherBody) uint {

	// If CashAmount is greater than 0 (and not zero)
	if body.PaymentVoucher.CashAmount > 0 {
		return 70034
	}

	// If zero or empty (default float64 is 0)
	return 70035
}
