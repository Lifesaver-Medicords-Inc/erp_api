package accounting_report_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services/accounting_report_services"
	"github.com/pierceperado/smpc/utils"
)

type AccountingReportHandler struct {
	TrialBalanceService    *accounting_report_services.TrialBalanceService
	IncomeStatementService *accounting_report_services.IncomeStatementService
	BalanceSheetService    *accounting_report_services.BalanceSheetService
	CashFlowService        *accounting_report_services.CashFlowService
	FinancialRatiosService *accounting_report_services.FinancialRatiosService
}

func NewAccountingReportHandler(
	trialBalanceService *accounting_report_services.TrialBalanceService,
	incomeStatementService *accounting_report_services.IncomeStatementService,
	balanceSheetService *accounting_report_services.BalanceSheetService,
	cashFlowService *accounting_report_services.CashFlowService,
	financialRatiosService *accounting_report_services.FinancialRatiosService,
) *AccountingReportHandler {
	return &AccountingReportHandler{
		TrialBalanceService:    trialBalanceService,
		IncomeStatementService: incomeStatementService,
		BalanceSheetService:    balanceSheetService,
		CashFlowService:        cashFlowService,
		FinancialRatiosService: financialRatiosService,
	}
}

// GetTrialBalance - GET /accounting/reports/trial_balance?as_of=YYYY-MM-DD&period_start=YYYY-MM-DD
//
// as_of is required, not defaulted to the server's own clock - CLAUDE.md's "'Now'
// is the user's PC clock" rule means the caller (the WinForms client) is the one
// that owns what "today" means, the same as everywhere else in this system; this
// endpoint just answers the question it's asked instead of silently picking its
// own answer. period_start is optional - omit it for a since-inception balance
// (Balance Sheet accounts), pass it for period-only activity (Income Statement
// accounts) - see GetTrialBalance's own doc comment.
func (h *AccountingReportHandler) GetTrialBalance(c *fiber.Ctx) error {
	asOf := c.Query("as_of")
	if asOf == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "as_of query parameter is required (YYYY-MM-DD)")
	}
	periodStart := c.Query("period_start")

	data, status, err := h.TrialBalanceService.GetTrialBalance(periodStart, asOf)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GetIncomeStatement - GET /accounting/reports/income_statement?period_start=YYYY-MM-DD&period_end=YYYY-MM-DD
// Both required, for the same "caller owns 'today'" reason as GetTrialBalance's as_of.
func (h *AccountingReportHandler) GetIncomeStatement(c *fiber.Ctx) error {
	periodStart := c.Query("period_start")
	periodEnd := c.Query("period_end")
	if periodStart == "" || periodEnd == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "period_start and period_end query parameters are required (YYYY-MM-DD)")
	}

	data, status, err := h.IncomeStatementService.GetIncomeStatement(periodStart, periodEnd)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GetBalanceSheet - GET /accounting/reports/balance_sheet?as_of=YYYY-MM-DD
// Required, for the same "caller owns 'today'" reason as GetTrialBalance's as_of.
func (h *AccountingReportHandler) GetBalanceSheet(c *fiber.Ctx) error {
	asOf := c.Query("as_of")
	if asOf == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "as_of query parameter is required (YYYY-MM-DD)")
	}

	data, status, err := h.BalanceSheetService.GetBalanceSheet(asOf)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GetCashFlow - GET /accounting/reports/cash_flow?period_start=YYYY-MM-DD&period_end=YYYY-MM-DD
// Both required, for the same "caller owns 'today'" reason as GetTrialBalance's as_of.
func (h *AccountingReportHandler) GetCashFlow(c *fiber.Ctx) error {
	periodStart := c.Query("period_start")
	periodEnd := c.Query("period_end")
	if periodStart == "" || periodEnd == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "period_start and period_end query parameters are required (YYYY-MM-DD)")
	}

	data, status, err := h.CashFlowService.GetCashFlow(periodStart, periodEnd)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GetFinancialRatios - GET /accounting/reports/financial_ratios?period_start=YYYY-MM-DD&period_end=YYYY-MM-DD
// Both required, for the same "caller owns 'today'" reason as GetTrialBalance's as_of.
func (h *AccountingReportHandler) GetFinancialRatios(c *fiber.Ctx) error {
	periodStart := c.Query("period_start")
	periodEnd := c.Query("period_end")
	if periodStart == "" || periodEnd == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "period_start and period_end query parameters are required (YYYY-MM-DD)")
	}

	data, status, err := h.FinancialRatiosService.GetFinancialRatios(periodStart, periodEnd)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
