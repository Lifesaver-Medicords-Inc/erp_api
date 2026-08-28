package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/accounting_report_handlers"
	"github.com/pierceperado/smpc/handlers/ap_voucher_handlers"
	"github.com/pierceperado/smpc/handlers/bulk_invoice_receipt_handlers"
	"github.com/pierceperado/smpc/handlers/invoice_receipt_handlers"
	"github.com/pierceperado/smpc/handlers/journal_entry_handlers"
	"github.com/pierceperado/smpc/handlers/payment_receipt_handlers"
	"github.com/pierceperado/smpc/handlers/payment_voucher_handlers"
	"github.com/pierceperado/smpc/handlers/sales_invoice_handlers"
	"github.com/pierceperado/smpc/handlers/setup_handlers"
	"github.com/pierceperado/smpc/services/accounting_report_services"
	"github.com/pierceperado/smpc/services/ap_voucher_services"
	"github.com/pierceperado/smpc/services/bulk_invoice_receipt_services"
	"github.com/pierceperado/smpc/services/invoice_receipt_services"
	"github.com/pierceperado/smpc/services/journal_entry_services"
	"github.com/pierceperado/smpc/services/payment_receipt_services"
	"github.com/pierceperado/smpc/services/payment_voucher_services"
	"github.com/pierceperado/smpc/services/sales_invoice_services"
	"github.com/pierceperado/smpc/services/setup_services"
)

func AccountingRoutes(router fiber.Router) {
	accountingApi := router.Group("/accounting")

	setupInvoiceReceiptRoutes(accountingApi)
	setupPaymentReceiptRoutes(accountingApi)
	setupSalesInvoiceRoutes(accountingApi)
	setupBulkInvoiceReceiptRoutes(accountingApi)
	setupPaymentVoucherRoutes(accountingApi)
	setupAPVoucherRoutes(accountingApi)
	setupJournalEntryRoutes(accountingApi)
	setupChartClassRoutes(accountingApi)
	setupChartOfAccountRoutes(accountingApi)
	setupTaxSetupRoutes(accountingApi)
	setupAccountingReportRoutes(accountingApi)
}

// setupAccountingReportRoutes - Admin "Reports" build (Balance Sheet, Income
// Statement, Statement of Cash Flows, Statement of Changes in Equity, then the 4
// ratio categories - see the approved plan). Increment 1: the Trial Balance
// aggregation every later statement composes on top of. Grouped under
// /accounting/reports/* so later increments (income_statement, balance_sheet,
// ...) land as siblings here rather than scattered across unrelated route groups.
func setupAccountingReportRoutes(api fiber.Router) {
	handler := accounting_report_handlers.NewAccountingReportHandler(
		accounting_report_services.NewTrialBalanceService(),
		accounting_report_services.NewIncomeStatementService(),
		accounting_report_services.NewBalanceSheetService(),
	)
	api.Get("/reports/trial_balance", handler.GetTrialBalance)
	api.Get("/reports/income_statement", handler.GetIncomeStatement)
	api.Get("/reports/balance_sheet", handler.GetBalanceSheet)
}

func setupInvoiceReceiptRoutes(api fiber.Router) {
	handler := invoice_receipt_handlers.NewInvoiceReceiptHandler(invoice_receipt_services.NewInvoiceReceiptService())
	api.Get("/invoice_receipt", handler.GetInvoiceReceipt)
	api.Get("/invoice_receipt/tax_view", handler.GetTaxView)
	api.Get("/invoice_receipt/supplier_trade", handler.GetSupplierTradeView)
	api.Get("/invoice_receipt/supplier_po/:supplier_id", handler.GetSupplierPO)
	api.Post("/invoice_receipt", handler.CreateInvoiceReceipt)
}

func setupBulkInvoiceReceiptRoutes(api fiber.Router) {
	handler := bulk_invoice_receipt_handlers.NewBulkInvoiceReceiptHandler(bulk_invoice_receipt_services.NewBulkInvoiceReceiptService())
	api.Post("/bulk_invoice_receipt", handler.CreateBulkInvoiceReceipt)
	api.Get("/bulk_invoice_receipt/search", handler.GetBulkInvoiceReceiptSearch)
	api.Get("/bulk_invoice_receipt", handler.GetBulkInvoiceReceipt)
}

func setupAPVoucherRoutes(api fiber.Router) {
	handler := ap_voucher_handlers.NewApVoucherHandler(ap_voucher_services.NewApVoucherService())
	api.Get("/ap_voucher/invoice/:supplier_id", handler.GetInvoiceView)
	api.Get("/ap_voucher", handler.GetApVoucher)
	api.Post("/ap_voucher", handler.CreateApVoucher)
}

func setupPaymentVoucherRoutes(api fiber.Router) {
	handler := payment_voucher_handlers.NewPaymentVoucherHandler(payment_voucher_services.NewPaymentVoucherService())
	api.Get("/payment_voucher", handler.GetPaymentVoucher)
	api.Post("/payment_voucher", handler.CreatePaymentVoucher)
	api.Get("/payment_voucher/ap_voucher/:supplier_id", handler.GetSupplierAPVoucher)
}

func setupSalesInvoiceRoutes(api fiber.Router) {
	handler := sales_invoice_handlers.NewSalesInvoiceHandler(sales_invoice_services.NewSalesInvoiceService())
	api.Get("/customer", handler.GetCustomer)
	api.Get("/customer_so/:customer_id", handler.GetCustomerSO)
	api.Get("/exchange_rate/:base_code", handler.GetExchangeRate)
	api.Get("/sales_invoice", handler.GetSalesInvoice)
	api.Post("/sales_invoice", handler.CreateSalesInvoice)
}

func setupPaymentReceiptRoutes(api fiber.Router) {
	handler := payment_receipt_handlers.NewPaymentReceiptHandler(payment_receipt_services.NewPaymentReceiptService())
	api.Get("/payment_receipt", handler.GetPaymentReceipt)
	api.Post("/payment_receipt", handler.CreatePaymentReceipt)
	api.Get("/payment_receipt/sales_invoice/:customer_id", handler.GetCustomerSalesInvoice)
}

func setupJournalEntryRoutes(api fiber.Router) {
	handler := journal_entry_handlers.NewJournalEntryHandler2(journal_entry_services.NewJournalEntryService2())
	api.Get("/company_setup", handler.GetCompanySetup)
	api.Get("/current_journal", handler.GetCurrentJournal)
	api.Get("/journal_entry", handler.GetJournalEntry)
	api.Post("/journal_entry", handler.CreateJournalEntry)
	api.Put("/journal_entry", handler.UpdateJournalEntry)
	api.Delete("/journal_entry", handler.DeleteJournalEntry)
}

func setupChartClassRoutes(api fiber.Router) {
	handler := setup_handlers.NewChartClassHandler(setup_services.NewChartClassService())
	api.Post("/chart_class", handler.CreateChartClass)
	api.Put("/chart_class", handler.UpdateChartClass)
	api.Delete("/chart_class", handler.DeleteChartClass)
	api.Get("/chart_class", handler.GetChartClasses)
}

func setupChartOfAccountRoutes(api fiber.Router) {
	handler := setup_handlers.NewChartOfAccountHandler(setup_services.NewChartOfAccountService())
	api.Get("/chart_of_account", handler.GetChartOfAccounts)
	api.Post("/chart_of_account", handler.CreateChartOfAccount)
	api.Put("/chart_of_account", handler.UpdateChartOfAccount)
	api.Delete("/chart_of_account", handler.DeleteChartOfAccount)
}

func setupTaxSetupRoutes(api fiber.Router) {
	handler := setup_handlers.NewTaxSetupHandler(setup_services.NewTaxSetupService())
	api.Get("/tax", handler.GetTaxSetup)
	api.Get("/tax/coa", handler.GetChartOfAccountSetup)
	api.Get("/tax_setup/:code", handler.GetTaxClassificationSetup)
	api.Post("/tax", handler.CreateTaxSetup)
	api.Put("/tax", handler.UpdateTaxSetup)
	api.Delete("/tax", handler.DeleteTaxSetup)
}
