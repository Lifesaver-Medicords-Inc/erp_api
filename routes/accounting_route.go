package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/invoice_receipt_handlers"
	"github.com/pierceperado/smpc/handlers/journal_entry_handlers2"
	"github.com/pierceperado/smpc/services/invoice_receipt_services"
	"github.com/pierceperado/smpc/services/journal_entry_services2"
)

func AccountingRoutes(router fiber.Router) {
	accountingApi := router.Group("/accounting")

	//Invoice Receipt Endpoints
	invoiceReceiptService := invoice_receipt_services.NewInvoiceReceiptService()
	invoiceReceiptHandler := invoice_receipt_handlers.NewInvoiceReceiptHandler(invoiceReceiptService)
	accountingApi.Get("/invoice_receipt", invoiceReceiptHandler.GetInvoiceReceipt)
	accountingApi.Post("/invoice_receipt", invoiceReceiptHandler.CreateInvoiceReceipt)
	accountingApi.Put("/invoice_receipt", invoiceReceiptHandler.UpdateInvoiceReceipt)
	accountingApi.Delete("/invoice_receipt", invoiceReceiptHandler.DeleteInvoiceReceipt)

	//Journal Entry Endpoints
	journalEntryService2 := journal_entry_services2.NewJournalEntryService2()
	journalEntryHandler2 := journal_entry_handlers2.NewJournalEntryHandler2(journalEntryService2)
	accountingApi.Get("/company_setup", journalEntryHandler2.GetCompanySetup)
	accountingApi.Get("/journal_entry2", journalEntryHandler2.GetJournalEntry)
	accountingApi.Post("/journal_entry2", journalEntryHandler2.CreateJournalEntry)
	accountingApi.Put("/journal_entry2", journalEntryHandler2.UpdateJournalEntry)
	accountingApi.Delete("/journal_entry2", journalEntryHandler2.DeleteJournalEntry)
}
