package invoice_receipt_services

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
	"github.com/pierceperado/smpc/services/journal_entry_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type InvoiceReceiptService struct{}

func NewInvoiceReceiptService() *InvoiceReceiptService {
	return &InvoiceReceiptService{}
}

func (s *InvoiceReceiptService) GetInvoiceReceipt(conditions map[string]interface{}) (interface{}, int, error) {
	var response accounting_models.InvoiceReceiptGet

	if err := services.DbGet(&response.InvoiceReceipt, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting invoice receipt")
	}

	if err := services.DbGet(&response.InvoiceReceiptDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting invoice receipt details")
	}

	return response, fiber.StatusOK, nil
}

func (s *InvoiceReceiptService) GetSupplierPO(conditions map[string]interface{}) (interface{}, int, error) {
	var poParent []accounting_models.InvoicePOView

	// Get Purchase Order (Parent)
	if err := services.DbRaw(&poParent, "sp_GetPurchaseOrderInvoice", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting purchase order data")
	}

	if len(poParent) == 0 {
		return nil, fiber.StatusNotFound, errors.New("no purchase order found")
	}

	var allChildren []accounting_models.InvoicePODetailView

	for _, po := range poParent {
		var poChild []accounting_models.InvoicePODetailView

		childConditions := map[string]interface{}{
			"PurchaseId": po.ID,
		}

		if err := services.DbRaw(&poChild, "sp_GetPurchaseOrderDetailsInvoice", childConditions); err != nil {
			return nil, fiber.StatusInternalServerError, errors.New("failed getting purchase order details data")
		}

		// Append children to one slice
		allChildren = append(allChildren, poChild...)
	}

	// Response structure
	response := struct {
		PurchaseOrderView        []accounting_models.InvoicePOView       `json:"purchase_order_view"`
		PurchaseOrderDetailsView []accounting_models.InvoicePODetailView `json:"purchase_order_details_view"`
	}{
		PurchaseOrderView:        poParent,
		PurchaseOrderDetailsView: allChildren,
	}

	return response, fiber.StatusOK, nil
}

func (s *InvoiceReceiptService) GetTaxView(conditions map[string]interface{}) (interface{}, int, error) {
	var response []accounting_models.TaxView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting tax view")
	}

	return response, fiber.StatusOK, nil
}

func (s *InvoiceReceiptService) GetSupplierTradeView(conditions map[string]interface{}) (interface{}, int, error) {
	var response []accounting_models.SupplierTradeView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting supplier trade view")
	}

	return response, fiber.StatusOK, nil
}

func (s *InvoiceReceiptService) CreateInvoiceReceipt(body *accounting_models.InvoiceReceiptBody, at models.At) (*accounting_models.InvoiceReceiptBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback() // rollback unless committed

	nextDocNo, err := utils.NextDocNo(tx, new(accounting_models.InvoiceReceipt), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	body.InvoiceReceipt.DocNo = nextDocNo

	// Insert main Invoice Receipt
	if err := services.DbInsert(tx, &body.InvoiceReceipt); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating invoice receipt")
	}

	// Insert Invoice Receipt Details
	if err := s.CreateInvoiceReceiptDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Find matching journal entry
	var journals []accounting_models.JournalEntry
	if err := tx.Find(&journals).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching journal entries")
	}

	docDate, err := time.Parse("01/02/2006", strings.ToLower(strings.TrimSpace(body.InvoiceReceipt.DocDate)))
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("invalid invoice receipt doc date format")
	}

	var matchedJournalID uint
	for _, j := range journals {
		parts := strings.Split(j.Period, " to ")
		if len(parts) != 2 {
			continue
		}
		startDate, err1 := time.Parse("1/2/2006 3:04:05 PM", strings.TrimSpace(parts[0]))
		endDate, err2 := time.Parse("1/2/2006 3:04:05 PM", strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			continue
		}
		if !docDate.Before(startDate) && !docDate.After(endDate) {
			matchedJournalID = j.ID
			break
		}
	}
	if matchedJournalID == 0 {
		return body, fiber.StatusNotFound, errors.New("no journal entry found for the invoice period")
	}

	// Fetch debit and credit COAs
	var coaDEBIT, coaCREDIT accounting_models.ChartOfAccounts
	if err := tx.First(&coaDEBIT, 70036).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching debit chart of account")
	}
	if err := tx.First(&coaCREDIT, 70033).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching credit chart of account")
	}

	// Helper to create journal entry
	createJournalEntry := func(account accounting_models.ChartOfAccounts, amount float64, isCredit bool) accounting_models.JournalEntryDetails {
		entry := accounting_models.JournalEntryDetails{
			JournalEntryDetailsContent: accounting_models.JournalEntryDetailsContent{
				Origin:         "Invoice Receipt",
				OriginId:       body.InvoiceReceipt.ID,
				LineMemo:       "Auto entry - " + map[bool]string{true: "CREDIT", false: "DEBIT"}[isCredit],
				PostingDate:    body.InvoiceReceipt.DocDate,
				CreatedBy:      body.InvoiceReceipt.PreparedBy,
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

	jeService := journal_entry_services.NewJournalEntryService2()

	// Auto-insert debit and credit
	for _, e := range []accounting_models.JournalEntryDetails{
		createJournalEntry(coaCREDIT, body.InvoiceReceipt.NetAmount, true),
		createJournalEntry(coaDEBIT, body.InvoiceReceipt.NetAmount, false),
	} {
		if err := jeService.AutoInsertJournalEntry(&e, body.InvoiceReceipt.DocDate, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	// Insert audit record
	atdata := accounting_models.InvoiceReceiptAt{
		RefId:                 body.InvoiceReceipt.ID,
		InvoiceReceiptContent: body.InvoiceReceipt.InvoiceReceiptContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating invoice receipt at")
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

func (s *InvoiceReceiptService) CreateInvoiceReceiptDetails(tx *gorm.DB, body *accounting_models.InvoiceReceiptBody, at models.At) error {
	for i := range body.InvoiceReceiptDetails {
		detail := &body.InvoiceReceiptDetails[i]
		detail.InvoiceReceiptID = body.InvoiceReceipt.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating invoice receipt details")
		}

		// Audit trail for each detail
		atdataDetail := accounting_models.InvoiceReceiptDetailsAt{
			RefId:                        detail.ID,
			InvoiceReceiptDetailsContent: detail.InvoiceReceiptDetailsContent,
			At:                           at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating invoice receipt details at")
		}
	}
	return nil
}
