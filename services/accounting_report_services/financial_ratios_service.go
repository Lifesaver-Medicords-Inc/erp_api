package accounting_report_services

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models/accounting_models"
)

// tradeReceivableCode - the CODE of the Trade Receivable account, not its id.
//
// This used to be a hardcoded id (70032 on the old chart). That silently broke
// the moment the chart of accounts changed: the chart is Setup data that
// management owns and is expected to rebuild, and a rebuilt account gets a new
// identity. Nothing errored - the id simply matched no row, beginning and
// ending A/R both read 0, and Receivables Turnover printed 0.00 on a financial
// report as though that were a real measurement.
//
// The code is the part management actually controls and can see, and it
// follows the ASSET prefix convention enforced in Chart Of Account Setup. It
// is still a fixed value, so if no account carries it the ratio is reported as
// NOT COMPUTED rather than as zero - see resolveTradeReceivable. The permanent
// answer is the GL mapping layer (§12.8), which does not exist yet.
const tradeReceivableCode = "100002"

// resolveTradeReceivable looks the account up by code at report time rather
// than trusting a compiled-in identity.
//
// Deliberately queries the chart of accounts rather than scanning the trial
// balance rows: the trial balance's HAVING clause drops accounts with no
// activity, so "absent from the trial balance" cannot tell "this account does
// not exist" apart from "this account exists and had no movement". Those need
// opposite treatment - the first must block the ratio, the second is a
// legitimate zero.
// Limit(1).Find rather than First: a missing account is an EXPECTED condition
// here - it is the whole reason this function returns a bool - and First logs
// gorm.ErrRecordNotFound at error level, which would put a red "record not
// found" in the server log every time this report ran on a chart that had been
// rebuilt. Find leaves an empty result empty.
func resolveTradeReceivable() (uint, bool) {
	var accounts []accounting_models.ChartOfAccounts

	if err := initializers.DB.
		Where("code = ?", tradeReceivableCode).
		Limit(1).
		Find(&accounts).Error; err != nil {
		return 0, false
	}

	if len(accounts) == 0 {
		return 0, false
	}

	return accounts[0].ID, true
}

type FinancialRatiosService struct {
	BalanceSheetService    *BalanceSheetService
	IncomeStatementService *IncomeStatementService
	TrialBalanceService    *TrialBalanceService
}

func NewFinancialRatiosService() *FinancialRatiosService {
	return &FinancialRatiosService{
		BalanceSheetService:    NewBalanceSheetService(),
		IncomeStatementService: NewIncomeStatementService(),
		TrialBalanceService:    NewTrialBalanceService(),
	}
}

func safeDiv(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// §12.10's classification, as stored in
// tbl_setup_chart_of_accounts.liquidity_class. Fixed in code rather than
// maintained in Setup, matching CashFlowCategory's two values and the five
// class types ChartClassPage hardcodes - §17.5's precedent for a list that
// is not expected to change. §12.10 names CURRENT and NON-CURRENT itself.
const (
	liquidityCurrent    = "CURRENT"
	liquidityNonCurrent = "NON-CURRENT"
	liquidityCash       = "CASH"
)

// isCurrent - CASH is a SUBSET of CURRENT, not an alternative to it, so a
// cash account counts toward current assets as well as toward the cash
// ratio. NON-CURRENT and "" both return false, but they are not the same
// thing: "" means nobody has decided yet, and the caller reports it.
func isCurrent(liquidityClass string) bool {
	return liquidityClass == liquidityCurrent || liquidityClass == liquidityCash
}

// computeLiquidity fills in §12.10's four figures from the classification
// management set on each account.
//
// It reads the Balance Sheet's own row lists rather than re-querying, so the
// two reports cannot disagree about what an account's balance is, and it
// repeats the Balance Sheet's sign convention exactly: ASSET accounts are
// debit-normal and use NetBalance as-is, LIABILITY accounts are credit-normal
// and are flipped.
//
// Nothing here looks at an account's code, class or name to decide whether it
// is current - §12.10 forbids inferring it, and an unclassified account is
// excluded and named rather than assumed.
func computeLiquidity(balanceSheet *accounting_models.BalanceSheetResult, result *accounting_models.FinancialRatiosResult) {
	var classifiedAssets, classifiedLiabilities int

	for _, row := range balanceSheet.AssetAccounts {
		switch {
		case row.LiquidityClass == "":
			result.UnclassifiedAccounts = append(result.UnclassifiedAccounts, row)
		case isCurrent(row.LiquidityClass):
			classifiedAssets++
			result.CurrentAssets += row.NetBalance
			if row.LiquidityClass == liquidityCash {
				result.CashAndEquivalents += row.NetBalance
			}
		case row.LiquidityClass == liquidityNonCurrent:
			classifiedAssets++
		}
	}

	for _, row := range balanceSheet.LiabilityAccounts {
		switch {
		case row.LiquidityClass == "":
			result.UnclassifiedAccounts = append(result.UnclassifiedAccounts, row)
		case isCurrent(row.LiquidityClass):
			classifiedLiabilities++
			result.CurrentLiabilities += row.TotalCredit - row.TotalDebit
		case row.LiquidityClass == liquidityNonCurrent:
			classifiedLiabilities++
		}
	}

	// Blocked only when one whole side is unclassified - there is then no
	// denominator, or no numerator, and any figure would be fiction. A
	// PARTIAL classification still computes: the ratios reflect what is
	// classified and UnclassifiedAccounts names the rest, which is what
	// §12.10 asks for ("excluded and reported as unclassified").
	if classifiedAssets == 0 || classifiedLiabilities == 0 {
		result.LiquidityBlocked = true
		result.LiquidityNote = "Liquidity needs a CURRENT / NON-CURRENT classification on the Chart " +
			"of Accounts (§12.10), and one whole side has none yet - classify the asset and liability " +
			"accounts under Setup > Chart of Accounts. Nothing is computed in the meantime; a guess at " +
			"which accounts are current would overstate every ratio."
		return
	}

	// QuickRatioInventory is whatever the Balance Sheet reports, so the two
	// reports agree by construction. Today that is the FIFO inventory module
	// rather than the ledger, because INVENTORY has no postings - see
	// BalanceSheetResult.InventoryIsFromInventoryModule.
	result.QuickRatioInventory = balanceSheet.Inventory

	result.CurrentRatio = safeDiv(result.CurrentAssets, result.CurrentLiabilities)
	result.QuickRatio = safeDiv(result.CurrentAssets-result.QuickRatioInventory, result.CurrentLiabilities)
	result.CashRatio = safeDiv(result.CashAndEquivalents, result.CurrentLiabilities)
	result.WorkingCapital = result.CurrentAssets - result.CurrentLiabilities

	if len(result.UnclassifiedAccounts) > 0 {
		result.LiquidityNote = "Computed from the classified accounts only. " +
			strconv.Itoa(len(result.UnclassifiedAccounts)) + " account(s) carrying a balance are still " +
			"unclassified and are EXCLUDED from these figures - see unclassified_accounts. Classify them " +
			"under Setup > Chart of Accounts for a complete picture."
	}
}

// GetFinancialRatios — Liquidity, Solvency, Profitability, and Efficiency,
// computed over the Balance Sheet as of periodEnd and the Income Statement
// for [periodStart, periodEnd].
//
// Liquidity (§12.10) is no longer unconditionally blocked: it computes from
// the CURRENT / NON-CURRENT / CASH classification management sets per
// account, and blocks only when a whole side of that classification is still
// missing. See computeLiquidity.
func (s *FinancialRatiosService) GetFinancialRatios(periodStart, periodEnd string) (*accounting_models.FinancialRatiosResult, int, error) {
	startDate, err := time.Parse("2006-01-02", periodStart)
	if err != nil {
		return nil, fiber.StatusBadRequest, errors.New("period_start must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", periodEnd); err != nil {
		return nil, fiber.StatusBadRequest, errors.New("period_end must be in YYYY-MM-DD format")
	}

	balanceSheet, status, err := s.BalanceSheetService.GetBalanceSheet(periodEnd)
	if err != nil {
		return nil, status, err
	}

	incomeStatement, status, err := s.IncomeStatementService.GetIncomeStatement(periodStart, periodEnd)
	if err != nil {
		return nil, status, err
	}

	result := &accounting_models.FinancialRatiosResult{
		AsOf:        periodEnd,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,

		DebtToEquity: safeDiv(balanceSheet.TotalLiabilities, balanceSheet.TotalEquity),
		DebtRatio:    safeDiv(balanceSheet.TotalLiabilities, balanceSheet.TotalAssets),

		GrossProfitMargin: safeDiv(incomeStatement.GrossProfit, incomeStatement.Revenue),
		NetProfitMargin:   safeDiv(incomeStatement.NetIncome, incomeStatement.Revenue),
		ReturnOnAssets:    safeDiv(incomeStatement.NetIncome, balanceSheet.TotalAssets),
		ReturnOnEquity:    safeDiv(incomeStatement.NetIncome, balanceSheet.TotalEquity),

		InventoryTurnover:                  safeDiv(incomeStatement.CostOfSales, balanceSheet.Inventory),
		InventoryTurnoverUsesEndingBalance: true,
		AssetTurnover:                      safeDiv(incomeStatement.Revenue, balanceSheet.TotalAssets),
		AssetTurnoverUsesEndingBalance:     true,
	}

	// Liquidity - §12.10. Reads the classification management set on each
	// account; blocks rather than guesses when a whole side is unclassified.
	computeLiquidity(balanceSheet, result)

	// Receivables Turnover: Trade Receivable is a real ledger account, so
	// unlike Inventory/Total Assets above, a true period average is
	// genuinely computable - fetch its balance the instant before the
	// period opens as well as at period end.
	//
	// Resolved by code at report time. If management's chart carries no such
	// account the ratio is BLOCKED and says so, rather than dividing by an
	// average of two zeros and printing 0.00 as if it had measured something.
	tradeReceivableId, found := resolveTradeReceivable()
	if !found {
		result.ReceivablesTurnoverBlocked = true
		result.ReceivablesTurnoverNote = "Not computed - no account with code " + tradeReceivableCode +
			" (TRADE RECEIVABLE) exists in the Chart of Accounts. Receivables Turnover is the only ratio " +
			"that needs a specific named account; every other figure on this report is derived from account " +
			"classes. Add the account under Setup > Chart of Accounts, or tell the developers which code " +
			"replaced it."
		return result, fiber.StatusOK, nil
	}

	dayBeforeStart := startDate.AddDate(0, 0, -1).Format("2006-01-02")
	beginRows, status, err := s.TrialBalanceService.GetTrialBalance("", dayBeforeStart)
	if err != nil {
		return nil, status, err
	}
	var beginAR float64
	for _, r := range beginRows {
		if r.AccountId == tradeReceivableId {
			beginAR = r.NetBalance
			break
		}
	}
	var endAR float64
	for _, r := range balanceSheet.AssetAccounts {
		if r.AccountId == tradeReceivableId {
			endAR = r.NetBalance
			break
		}
	}
	averageAR := (beginAR + endAR) / 2
	result.ReceivablesTurnover = safeDiv(incomeStatement.Revenue, averageAR)

	return result, fiber.StatusOK, nil
}
