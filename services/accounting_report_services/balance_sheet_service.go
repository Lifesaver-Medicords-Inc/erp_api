package accounting_report_services

import (
	"errors"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services/item_stock_services"
	"github.com/pierceperado/smpc/services/setup_services"
)

type BalanceSheetService struct {
	TrialBalanceService *TrialBalanceService
	ItemStockService    *item_stock_services.ItemStockService
	FixedAssetService   *setup_services.FixedAssetService
}

func NewBalanceSheetService() *BalanceSheetService {
	return &BalanceSheetService{
		TrialBalanceService: NewTrialBalanceService(),
		ItemStockService:    item_stock_services.NewItemStockService(),
		FixedAssetService:   setup_services.NewFixedAssetService(),
	}
}

// GetBalanceSheet is increment 3 of the Admin "Reports" build (approved plan).
// Assets/Liabilities/Equity come from the Trial Balance engine with no
// periodStart (Balance Sheet accounts are cumulative since inception, unlike
// Income Statement accounts - see GetTrialBalance's own doc comment).
//
// Sign convention: ASSET accounts are debit-normal, so TrialBalanceRow.NetBalance
// is used as-is. LIABILITY/EQUITY accounts are credit-normal, so they're flipped
// to TotalCredit-TotalDebit, same convention the Income Statement uses for
// REVENUE.
func (s *BalanceSheetService) GetBalanceSheet(asOfDate string) (*accounting_models.BalanceSheetResult, int, error) {
	if _, err := time.Parse("2006-01-02", asOfDate); err != nil {
		return nil, fiber.StatusBadRequest, errors.New("as_of must be in YYYY-MM-DD format")
	}

	rows, status, err := s.TrialBalanceService.GetTrialBalance("", asOfDate)
	if err != nil {
		return nil, status, err
	}

	result := &accounting_models.BalanceSheetResult{
		AsOf:                           asOfDate,
		InventoryIsFromInventoryModule: true,
	}

	for _, row := range rows {
		switch row.AccountClass {
		case "ASSET":
			result.CashAndOtherAssets += row.NetBalance
			result.AssetAccounts = append(result.AssetAccounts, row)
		case "LIABILITY":
			result.TotalLiabilities += row.TotalCredit - row.TotalDebit
			result.LiabilityAccounts = append(result.LiabilityAccounts, row)
		case "EQUITY":
			result.TotalEquity += row.TotalCredit - row.TotalDebit
			result.EquityAccounts = append(result.EquityAccounts, row)
		}
	}

	result.GlAssetsEqualsLiabilitiesPlusEquity = math.Abs(
		result.CashAndOtherAssets-(result.TotalLiabilities+result.TotalEquity),
	) < 0.01

	inventoryValue, err := s.ItemStockService.GetInventoryValue()
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting inventory value")
	}
	result.Inventory = inventoryValue

	// PropertyAndEquipment: FixedAssetService stores dates as MM/dd/yyyy
	// (this codebase's usual free-text date convention), so re-format the
	// ISO as_of this handler already validated above before calling it.
	asOfParsed, _ := time.Parse("2006-01-02", asOfDate)
	ppe, err := s.FixedAssetService.GetPPEAsOf(asOfParsed.Format("01/02/2006"))
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting property and equipment")
	}
	result.PropertyAndEquipment = ppe.TotalNetBookValue
	result.PropertyAndEquipmentCategories = ppe.Categories
	result.PropertyAndEquipmentIsTracked = len(ppe.Categories) > 0

	result.TotalAssets = result.CashAndOtherAssets + result.Inventory + result.PropertyAndEquipment
	result.TotalLiabilitiesAndEquity = result.TotalLiabilities + result.TotalEquity

	return result, fiber.StatusOK, nil
}
