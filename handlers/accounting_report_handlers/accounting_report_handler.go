package accounting_report_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services/accounting_report_services"
	"github.com/pierceperado/smpc/utils"
)

type AccountingReportHandler struct {
	TrialBalanceService *accounting_report_services.TrialBalanceService
}

func NewAccountingReportHandler(trialBalanceService *accounting_report_services.TrialBalanceService) *AccountingReportHandler {
	return &AccountingReportHandler{TrialBalanceService: trialBalanceService}
}

// GetTrialBalance - GET /accounting/reports/trial_balance?as_of=YYYY-MM-DD
//
// as_of is required, not defaulted to the server's own clock - CLAUDE.md's "'Now'
// is the user's PC clock" rule means the caller (the WinForms client) is the one
// that owns what "today" means, the same as everywhere else in this system; this
// endpoint just answers the question it's asked instead of silently picking its
// own answer.
func (h *AccountingReportHandler) GetTrialBalance(c *fiber.Ctx) error {
	asOf := c.Query("as_of")
	if asOf == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "as_of query parameter is required (YYYY-MM-DD)")
	}

	data, status, err := h.TrialBalanceService.GetTrialBalance(asOf)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
