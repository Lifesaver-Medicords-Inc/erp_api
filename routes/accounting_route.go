package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/ap_voucher_handlers"
	"github.com/pierceperado/smpc/handlers/bulk_invoice_receipt_handlers"
	"github.com/pierceperado/smpc/handlers/invoice_receipt_handlers"
	"github.com/pierceperado/smpc/handlers/journal_entry_handlers2"
	"github.com/pierceperado/smpc/handlers/payment_receipt_handlers"
	"github.com/pierceperado/smpc/handlers/payment_voucher_handlers"
	"github.com/pierceperado/smpc/handlers/sales_invoice_handlers2"
	"github.com/pierceperado/smpc/handlers/setup_handlers"
	"github.com/pierceperado/smpc/services/ap_voucher_services"
	"github.com/pierceperado/smpc/services/bulk_invoice_receipt_services"
	"github.com/pierceperado/smpc/services/invoice_receipt_services"
	"github.com/pierceperado/smpc/services/journal_entry_services2"
	"github.com/pierceperado/smpc/services/payment_receipt_services"
	"github.com/pierceperado/smpc/services/payment_voucher_services"
	"github.com/pierceperado/smpc/services/sales_invoice_services2"
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

	//Payment Receipt Endpoints
	paymentReceiptService := payment_receipt_services.NewPaymentReceiptService()
	paymentReceiptHandler := payment_receipt_handlers.NewPaymentReceiptHandler(paymentReceiptService)
	accountingApi.Get("/payment_receipt", paymentReceiptHandler.GetPaymentReceipt)
	accountingApi.Post("/payment_receipt", paymentReceiptHandler.CreatePaymentReceipt)

	//Sales Invoice Endpoints
	salesInvoiceService := sales_invoice_services2.NewSalesInvoiceService()
	salesInvoiceHandler := sales_invoice_handlers2.NewSalesInvoiceHandler(salesInvoiceService)
	accountingApi.Get("/customer", salesInvoiceHandler.GetCustomer)
	accountingApi.Get("/customer_so/:customer_id", salesInvoiceHandler.GetCustomerSO)
	accountingApi.Get("/exchange_rate/:base_code", salesInvoiceHandler.GetExchangeRate)
	accountingApi.Get("/sales_invoice2", salesInvoiceHandler.GetSalesInvoice)
	accountingApi.Post("/sales_invoice2", salesInvoiceHandler.CreateSalesInvoice)

	//Bulk Invoice Receipt Endpoints
	bulkInvoiceReceiptService := bulk_invoice_receipt_services.NewBulkInvoiceReceiptService()
	bulkInvoiceReceiptHandler := bulk_invoice_receipt_handlers.NewBulkInvoiceReceiptHandler(bulkInvoiceReceiptService)
	accountingApi.Get("/bulk_invoice_receipt", bulkInvoiceReceiptHandler.GetBulkInvoiceReceipt)
	accountingApi.Post("/bulk_invoice_receipt", bulkInvoiceReceiptHandler.CreateBulkInvoiceReceipt)

	//Payment Voucher Endpoints
	paymentVoucherService := payment_voucher_services.NewPaymentVoucherService()
	paymentVoucherHandler := payment_voucher_handlers.NewPaymentVoucherHandler(paymentVoucherService)
	accountingApi.Get("/payment_voucher", paymentVoucherHandler.GetPaymentVoucher)
	accountingApi.Post("/payment_voucher", paymentVoucherHandler.CreatePaymentVoucher)
	accountingApi.Get("/payment_voucher/ap_voucher/:supplier_id", paymentVoucherHandler.GetSupplierAPVoucher)

	//AP Voucher Endpoints
	apVoucherService := ap_voucher_services.NewApVoucherService()
	apVoucherHandler := ap_voucher_handlers.NewApVoucherHandler(apVoucherService)
	accountingApi.Get("/ap_voucher/invoice/:supplier_id", apVoucherHandler.GetInvoiceView)
	accountingApi.Get("/ap_voucher", apVoucherHandler.GetApVoucher)
	accountingApi.Post("/ap_voucher", apVoucherHandler.CreateApVoucher)

	//Journal Entry Endpoints
	journalEntryService2 := journal_entry_services2.NewJournalEntryService2()
	journalEntryHandler2 := journal_entry_handlers2.NewJournalEntryHandler2(journalEntryService2)
	accountingApi.Get("/company_setup", journalEntryHandler2.GetCompanySetup)
	accountingApi.Get("/current_journal", journalEntryHandler2.GetCurrentJournal)
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
