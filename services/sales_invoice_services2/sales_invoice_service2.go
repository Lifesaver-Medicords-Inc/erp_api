package sales_invoice_services2

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
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/services/journal_entry_services2"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type SalesInvoiceService struct{}

func NewSalesInvoiceService() *SalesInvoiceService {
	return &SalesInvoiceService{}
}

func (s *SalesInvoiceService) GetExchangeRate(baseCode string) (interface{}, int, error) {
	erService := adminservices.NewExchangeRateService()

	response, err := erService.GetCurrencyAPI(baseCode)
	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid currency code") {
			return nil, fiber.StatusBadRequest, err
		}
		return nil, fiber.StatusInternalServerError, err
	}

	return response, fiber.StatusOK, nil
}

func (s *SalesInvoiceService) GetCustomer(conditions map[string]interface{}) (interface{}, int, error) {
	var response []accounting_models.CustomerView

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting customer view")
	}

	return response, fiber.StatusOK, nil
}

func (s *SalesInvoiceService) GetCustomerSO(conditions map[string]interface{}) (interface{}, int, error) {
	var poParent []accounting_models.InvoiceSOView

	// Get Sales Order (Parent)
	if err := services.DbRaw(&poParent, "sp_GetSalesOrderInvoice", conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting sales order data")
	}

	if len(poParent) == 0 {
		return nil, fiber.StatusNotFound, errors.New("no sales order found")
	}

	var allChildren []accounting_models.InvoiceSODetailView

	for _, po := range poParent {
		var poChild []accounting_models.InvoiceSODetailView

		childConditions := map[string]interface{}{
			"SalesId": po.SalesOrderID,
		}

		if err := services.DbRaw(&poChild, "sp_GetSalesOrderDetailsInvoice", childConditions); err != nil {
			return nil, fiber.StatusInternalServerError, errors.New("failed getting sales order details data")
		}

		// Convert DateDeliver in child records
		for i := range poChild {
			if poChild[i].DateDeliver != "" {
				parsedDate, err := time.Parse("2006-01-02", poChild[i].DateDeliver)
				if err == nil {
					poChild[i].DateDeliver = parsedDate.Format("01/02/2006") // MM/dd/yyyy
				}
			}
		}

		// Append children to one slice
		allChildren = append(allChildren, poChild...)
	}

	// Convert DocDate in parent records
	for i := range poParent {
		if poParent[i].DocDate != "" {
			parsedDate, err := time.Parse("2006-01-02", poParent[i].DocDate)
			if err == nil {
				poParent[i].DocDate = parsedDate.Format("01/02/2006") // MM/dd/yyyy
			}
		}
	}

	// Response structure
	response := struct {
		SalesOrderView        []accounting_models.InvoiceSOView       `json:"sales_order_view"`
		SalesOrderDetailsView []accounting_models.InvoiceSODetailView `json:"sales_order_details_view"`
	}{
		SalesOrderView:        poParent,
		SalesOrderDetailsView: allChildren,
	}

	return response, fiber.StatusOK, nil
}

func (s *SalesInvoiceService) GetSalesInvoice(conditions map[string]interface{}) (interface{}, int, error) {
	var response accounting_models.SalesInvoice2Get

	if err := services.DbGet(&response.SalesInvoice2, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting sales invoice")
	}

	if err := services.DbGet(&response.SalesInvoiceDetails2, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting sales invoice details")
	}

	return response, fiber.StatusOK, nil
}

func (s *SalesInvoiceService) CreateSalesInvoice(body *accounting_models.SalesInvoice2Body, at models.At) (*accounting_models.SalesInvoice2Body, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback() // rollback unless committed

	nextDocNo, err := utils.NextDocNo(tx, new(accounting_models.SalesInvoice2), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	body.SalesInvoice2.DocNo = nextDocNo

	// Insert main Sales Invoice
	if err := services.DbInsert(tx, &body.SalesInvoice2); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales invoice")
	}

	// Insert Sales Invoice Details
	if err := s.CreateSalesInvoiceDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Find matching journal entry
	var journals []accounting_models.JournalEntry2
	if err := tx.Find(&journals).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching journal entries")
	}

	docDate, err := time.Parse("01/02/2006", strings.ToLower(strings.TrimSpace(body.SalesInvoice2.DocDate)))
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("invalid sales invoice doc date format")
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
		return body, fiber.StatusNotFound, errors.New("no journal entry found for the sales invoice period")
	}

	// Fetch debit and credit COAs
	var coaDEBIT, coaCREDIT accounting_models.ChartOfAccounts
	if err := tx.First(&coaDEBIT, 40029).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching debit chart of account")
	}
	if err := tx.First(&coaCREDIT, 60030).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed fetching credit chart of account")
	}

	// Helper to create journal entry
	createJournalEntry := func(account accounting_models.ChartOfAccounts, amount float64, isCredit bool) accounting_models.JournalEntryDetails2 {
		entry := accounting_models.JournalEntryDetails2{
			JournalEntryDetails2Content: accounting_models.JournalEntryDetails2Content{
				Origin:         "Sales Invoice",
				OriginId:       body.SalesInvoice2.ID,
				LineMemo:       "Auto entry - " + map[bool]string{true: "CREDIT", false: "DEBIT"}[isCredit],
				PostingDate:    body.SalesInvoice2.DocDate,
				CreatedBy:      body.SalesInvoice2.PreparedBy,
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
		createJournalEntry(coaCREDIT, body.SalesInvoice2.TotalAmountDue, true),
		createJournalEntry(coaDEBIT, body.SalesInvoice2.TotalAmountDue, false),
	} {
		if err := jeService.AutoInsertJournalEntry(&e, body.SalesInvoice2.DocDate, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	// Insert audit record
	atdata := accounting_models.SalesInvoice2At{
		RefId:                body.SalesInvoice2.ID,
		SalesInvoiceContent2: body.SalesInvoice2.SalesInvoiceContent2,
		At:                   at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales invoice at")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	setup_services.InvalidateItemCaches()

	if err := services.InvalidateCacheByModel(accounting_models.InvoiceSOView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(accounting_models.InvoiceSODetailView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	if err := services.InvalidateCacheByModel(accounting_models.SalesInvoiceReceiptView{}); err != nil {
		fmt.Println("Failed to invalidate cache:", err)
	}

	return body, fiber.StatusOK, nil
}

func (s *SalesInvoiceService) CreateSalesInvoiceDetails(tx *gorm.DB, body *accounting_models.SalesInvoice2Body, at models.At) error {
	for i := range body.SalesInvoiceDetails2 {
		detail := &body.SalesInvoiceDetails2[i]
		detail.SalesInvoiceID = body.SalesInvoice2.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating payment voucher details")
		}

		// Audit trail for each detail
		atdataDetail := accounting_models.SalesInvoiceDetails2At{
			RefId:                       detail.ID,
			SalesInvoiceDetails2Content: detail.SalesInvoiceDetails2Content,
			At:                          at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating sales invoice details at")
		}
	}
	return nil
}
