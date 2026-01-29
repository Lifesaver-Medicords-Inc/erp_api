package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/bulk_invoice_receipt_handlers"
	"github.com/pierceperado/smpc/handlers/invoice_receipt_handlers"
	"github.com/pierceperado/smpc/handlers/journal_entry_handlers2"
	"github.com/pierceperado/smpc/handlers/setup_handlers"
	"github.com/pierceperado/smpc/services/bulk_invoice_receipt_services"
	"github.com/pierceperado/smpc/services/invoice_receipt_services"
	"github.com/pierceperado/smpc/services/journal_entry_services2"
	"github.com/pierceperado/smpc/services/setup_services"
)

func AccountingRoutes(router fiber.Router) {
	accountingApi := router.Group("/accounting")

	//Invoice Receipt Endpoints
	invoiceReceiptService := invoice_receipt_services.NewInvoiceReceiptService()
	invoiceReceiptHandler := invoice_receipt_handlers.NewInvoiceReceiptHandler(invoiceReceiptService)
	accountingApi.Get("/invoice_receipt", invoiceReceiptHandler.GetInvoiceReceipt)
	accountingApi.Get("/invoice_receipt/tax_view", invoiceReceiptHandler.GetTaxView)
	accountingApi.Get("/invoice_receipt/supplier_trade", invoiceReceiptHandler.GetSupplierTradeView)
	accountingApi.Get("/invoice_receipt/supplier_po/:supplier_id", invoiceReceiptHandler.GetSupplierPO)
	accountingApi.Post("/invoice_receipt", invoiceReceiptHandler.CreateInvoiceReceipt)

	//Bulk Invoice Receipt Endpoints
	bulkInvoiceReceiptService := bulk_invoice_receipt_services.NewBulkInvoiceReceiptService()
	bulkInvoiceReceiptHandler := bulk_invoice_receipt_handlers.NewBulkInvoiceReceiptHandler(bulkInvoiceReceiptService)
	accountingApi.Get("/bulk_invoice_receipt", bulkInvoiceReceiptHandler.GetBulkInvoiceReceipt)
	accountingApi.Post("/bulk_invoice_receipt", bulkInvoiceReceiptHandler.CreateBulkInvoiceReceipt)

	//Journal Entry Endpoints
	journalEntryService2 := journal_entry_services2.NewJournalEntryService2()
	journalEntryHandler2 := journal_entry_handlers2.NewJournalEntryHandler2(journalEntryService2)
	accountingApi.Get("/company_setup", journalEntryHandler2.GetCompanySetup)
	accountingApi.Get("/journal_entry2", journalEntryHandler2.GetJournalEntry)
	accountingApi.Post("/journal_entry2", journalEntryHandler2.CreateJournalEntry)
	accountingApi.Put("/journal_entry2", journalEntryHandler2.UpdateJournalEntry)
	accountingApi.Delete("/journal_entry2", journalEntryHandler2.DeleteJournalEntry)

	// Chart Class Endpoints
	chartClassService := setup_services.NewChartClassService()
	chartClassHandler := setup_handlers.NewChartClassHandler(chartClassService)
	accountingApi.Get("/chart_class", chartClassHandler.GetChartClasses)
	accountingApi.Get("/chart_class:/id", chartClassHandler.GetChartClass)
	accountingApi.Post("/chart_class", chartClassHandler.CreateChartClass)
	accountingApi.Put("/chart_class", chartClassHandler.UpdateChartClass)
	accountingApi.Delete("/chart_class", chartClassHandler.DeleteChartClass)

	//Chart of Account Endpoints
	chartOfAccountService := setup_services.NewChartOfAccountService()
	chartOfAccountHandler := setup_handlers.NewChartOfAccountHandler(chartOfAccountService)
	accountingApi.Get("/chart_of_account", chartOfAccountHandler.GetChartOfAccounts)
	accountingApi.Post("/chart_of_account", chartOfAccountHandler.CreateChartOfAccount)
	accountingApi.Put("/chart_of_account", chartOfAccountHandler.UpdateChartOfAccount)
	accountingApi.Delete("/chart_of_account", chartOfAccountHandler.DeleteChartOfAccount)
	accountingApi.Get("/chart_of_account_classification/:code", chartOfAccountHandler.GetChartOfAccountClassification)

	// Tax Setup Endpoints
	taxSetupService := setup_services.NewTaxSetupService()
	taxSetupHandler := setup_handlers.NewTaxSetupHandler(taxSetupService)
	accountingApi.Get("/tax", taxSetupHandler.GetTaxSetup)
	accountingApi.Get("/tax/coa", taxSetupHandler.GetChartOfAccountSetup)
	accountingApi.Get("/tax_setup/:code", taxSetupHandler.GetTaxClassificationSetup)
	accountingApi.Post("/tax", taxSetupHandler.CreateTaxSetup)
	accountingApi.Put("/tax", taxSetupHandler.UpdateTaxSetup)
	accountingApi.Delete("/tax", taxSetupHandler.DeleteTaxSetup)
}
