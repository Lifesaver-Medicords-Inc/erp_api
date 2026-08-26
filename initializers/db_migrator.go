package initializers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pierceperado/smpc/models"
	accounting_models "github.com/pierceperado/smpc/models/accounting_models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	inventory_models "github.com/pierceperado/smpc/models/inventory_models"
)

// models "github.com/pierceperado/smpc/models"
// accounting_models "github.com/pierceperado/smpc/models/accounting_models"
// dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"

// // MigrateAll runs all database migrations - Use for deployment
func MigrateAll() {
	// migrateAdmin()
	migrateCompanySettings()
	// migrateSetup()
	// migrateItemManagement()
	// migrateBomBoq()
	// migrateInventoryWarehouse()
	// migrateEngineering()
	migrateAccessControl()
	migrateSalesCrm()
	migrateSalesProject()
	// migratePurchasingVendor()
	migrateBpi()
	// migrateAccounting()
	// migrateJobOrder()
	migrateLogisticsDispatching()
	// migrateVehicleManagement()
	// migrateInventoryTransaction()
}

// // MigrateModel migrates specific models or categories by name
// // Usage: MigrateModel("admin"), MigrateModel("item", "sales"), MigrateModel("all")
func MigrateModel(categories ...string) {
	if len(categories) == 0 {
		return
	}

	for _, category := range categories {
		cat := strings.ToLower(strings.TrimSpace(category))
		switch cat {
		case "admin":
			migrateAdmin()
		case "setup":
			migrateSetup()
		case "item", "item_management":
			migrateItemManagement()
		case "bom", "boq", "bom_boq":
			migrateBomBoq()
		case "inventory", "warehouse", "inventory_warehouse":
			migrateInventoryWarehouse()
		case "item_request", "pick_activity":
			migrateEngineering()
		case "sales", "crm", "sales_crm":
			migrateSalesCrm()
		case "sales_project":
			migrateSalesProject()
		case "returns", "sales_return", "purchase_return":
			migrateReturns()
		case "memos", "credit_memo", "debit_memo":
			migrateMemos()
		case "purchasing", "vendor", "purchasing_vendor":
			migratePurchasingVendor()
		case "bpi", "business_partner":
			migrateBpi()
		case "accounting":
			migrateAccounting()
		case "job_order":
			migrateJobOrder()
		case "logistics", "dispatching", "logistics_dispatching":
			migrateLogisticsDispatching()
		case "vehicle", "vehicle_management":
			migrateVehicleManagement()
		case "all":
			MigrateAll()
			return
		}
	}
}

// ============================================
// ADMIN & ORGANIZATION
// ============================================
func migrateAdmin() {
	// fmt.Println("=== Migrating BPI Module ===")

	// migrateAndLog(
	// 	&models.User{}, &models.UserAt{},
	// 	&models.PositionModel{}, &models.PositionAt{},
	// 	&models.PositionAccessModel{}, &models.PositionAccessAt{},
	// 	&models.UserPermissionModel{}, &models.UserPermissionAt{},
	// 	&models.Social{}, &models.SocialAt{},
	// 	&models.Entity{}, &models.EntityAt{},
	// 	&models.Application{}, &models.ApplicationAt{},
	// 	&models.CompanyModel{}, &models.CompanyAt{},
	// 	&models.CompanyAddressModel{}, &models.CompanyAddressAt{},
	// 	&models.CompanyContactModel{}, &models.CompanyContactAt{},
	// )
}

// migrateAdmin() above is commented out wholesale (User/Position/PositionAccess/
// UserPermission/Social/Entity/Application/Company/CompanyAddress/CompanyContact) -
// those predate this selective-migration setup and re-enabling the whole group is
// a much bigger blast radius than this needs. tbl_company already exists with its
// base columns (Phase 3 item 3.4 built and used it successfully), so this migrates
// only what's actually new on CompanyModel since then: markup_multiplier_price
// (declared on the model, but this exact gap is why the column was never created -
// confirmed directly against the live DB, "Invalid column name") and
// vat_rate_percent (Sales_Quotation_Bug_Report_2026-08-03.md #18 - both now
// configurable instead of hardcoded, see company_model.go's own comment on these
// two fields).
func migrateCompanySettings() {
	migrateAndLog(&models.CompanyModel{}, &models.CompanyAt{})
}

// ============================================
// ACCESS CONTROL CATALOG
// ============================================
// Only the new AccessModuleModel (tbl_access_modules) - the master catalog of every
// grantable screen/action, seeded from SMPC_User_Access_Level_List.xlsx (see
// SeedAccessModules in seed_access_modules.go, called once from main.go's init() right
// after this). Deliberately its own function, not folded into migrateAdmin() above,
// since migrateAdmin() itself is commented out (Position/User/PositionAccess tables
// already exist and are intentionally left alone here) - this only touches the one new
// table, same narrow-scope reasoning as the ItemStocksAt/StockReservation calls in
// migrateInventoryWarehouse().
func migrateAccessControl() {
	migrateAndLog(
		&models.AccessModuleModel{},
	)
}

// ============================================
// SETUP / MASTER DATA
// ============================================
func migrateSetup() {
	// fmt.Println("=== Migrating SETUP Module ===")

	// migrateAndLog(
	// 	&models.Class{}, &models.ClassAt{},
	// 	&models.Name{}, &models.NameAt{},
	// 	&models.Type{}, &models.TypeAt{},
	// 	&models.Material{}, &models.MaterialAt{},
	// 	&models.ValuationMethod{}, &models.ValuationMethodAt{},
	// 	&models.TradeType{}, &models.TradeTypeAt{},
	// 	&models.Brand{}, &models.BrandAt{},
	// 	&models.UnitMeasurement{}, &models.UnitMeasurementAt{},
	// 	&models.PaymentTerms{}, &models.PaymentTermsAt{},
	// 	&models.Class{}, &models.ClassAt{},
	// 	&models.Name{}, &models.NameAt{},
	// 	&models.Type{}, &models.TypeAt{},
	// 	&models.Material{}, &models.MaterialAt{},
	// 	&models.ValuationMethod{}, &models.ValuationMethodAt{},
	// 	&models.TradeType{}, &models.TradeTypeAt{},
	// )
}

// ============================================
// ITEM MANAGEMENT
// ============================================
func migrateItemManagement() {
	// fmt.Println("=== Migrating ITEM MANAGEMENT Module ===")

	// migrateAndLog(
	// 	&models.ItemSpecsTemplate{}, &models.ItemSpecsTemplateAt{},
	// 	&models.AdditionalSpecs{}, &models.AdditionalSpecsAt{},
	// 	&models.AdditionalSpecsPumpType{}, &models.AdditionalSpecsPumpTypeAt{},
	// 	&models.ItemImage{}, &models.ItemImageAt{},
	// 	&models.ItemInventory{}, &models.ItemInventoryAt{},
	// 	&models.Model{}, &models.ModelAt{},
	// 	&models.PumpType{}, &models.PumpTypeAt{},
	// 	&models.PumpCount{}, &models.PumpCountAt{},
	// 	&models.ItemTradeType{}, &models.ItemTradeTypeAt{},
	// )
}

// ============================================
// BOM & BOQ MANAGEMENT
// ============================================
func migrateBomBoq() {
	// fmt.Println("=== Migrating BOM & BOQ Module ===")
	// migrateAndLog(
	// 	&models.SetupItemBom{}, &models.SetupItemBomAt{},
	// 	&models.SetupItemBomDetails{}, &models.SetupItemBomDetailsAt{},
	// 	&models.ItemBoq{}, &models.ItemBoqAt{},
	// 	&models.ItemBoqDetails{}, &models.ItemBoqDetailsAt{},
	// 	&models.BoqNotes{}, &models.BoqNotesAt{},
	// 	&models.WiringUserInput{}, &models.WiringUserInputAt{},
	// )
}

// ============================================
// INVENTORY & WAREHOUSE
// ============================================
func migrateInventoryWarehouse() {
	// fmt.Println("=== Migrating INVENTORY & WAREHOUSE Module ===")
	// migrateAndLog(
	// 	&models.WarehouseUseType{}, &models.WarehouseUseTypeAt{},
	// 	&models.WarehouseName{}, &models.WarehouseNameAt{},
	// 	&models.WarehouseAddress{}, &models.WarehouseAddressAt{},
	// 	&models.WarehouseArea{}, &models.WarehouseAreaAt{},
	// )
	// migrateAndLog(
	// 	&models.InvTracker{}, &models.InvTrackerAt{},
	// 	&models.InventoryStocks{}, &models.InventoryStocksAt{},
	// 	&models.InventoryStocksHistory{}, &models.InventoryStocksHistoryAt{},
	//	&inventory_models.ItemStocks{}, &inventory_models.ItemStocksAt{},
	// )
	// Scoped narrowly to just the new Remarks column on the audit table (added for the
	// Inventory Item Stocks manual-adjustment feature) - not the broader commented-out
	// block above, which covers several other legacy models this isn't touching.
	migrateAndLog(
		&inventory_models.ItemStocksAt{},
	)
	// StockTransaction (tbl_inv_stock_transactions) - the trigger-written stock ledger.
	// Must run before migrations.RunSQLMigrations() (called later in main.go's init())
	// creates tr_inv_item_stocks_ledger, since the trigger inserts into this table.
	migrateAndLog(
		&inventory_models.StockTransaction{},
	)
	// StockLot / StockLotConsumption - FIFO purchase-cost tracking. Written directly by
	// item_stock_services (CreateStockLot/ConsumeLotsFIFO/ReleaseLotsFIFO), not by a
	// trigger, since FIFO consumption is inherently sequential/procedural.
	migrateAndLog(
		&inventory_models.StockLot{},
		&inventory_models.StockLotConsumption{},
	)
	// StockReservation - soft holds placed by Sales Quotation lines (see
	// quick_quotation_service.go). Never touched by the trigger; swept on a timer by
	// main.go's startStockReservationSweep, not here.
	migrateAndLog(
		&inventory_models.StockReservation{},
	)
	// migrateAndLog(
	//	&inventory_models.ReceivingReport{}, &inventory_models.ReceivingReportAt{},
	//	&inventory_models.ReceivingReportDetails{}, &inventory_models.ReceivingReportDetailsAt{},
	// )
	// migrateAndLog(
	// 	&models.ReceivingReport2{}, &models.ReceivingReportAt2{},
	// 	&models.ReceivingReportDetails2{}, &models.ReceivingReportDetailsAt2{},
	// 	&models.ReceivingHistory{}, &models.ReceivingHistoryAt{},
	// )
}

// ============================================
// INVENTORY TRANSACTIONS
// ============================================

func migrateInventoryTransaction() {
	// fmt.Println("=== Migrating INVENTORY TRANSACTIONS Modules ===")
	// migrateAndLog(
	// 	&models.ItemRequest{}, &models.ItemRequestAt{},
	// 	&models.ItemRequestDetails{}, &models.ItemRequestDetailsAt{},
	// 	&models.ItemRequestLocation{}, &models.ItemRequestLocationAt{},
	// 	&models.ItemRequestHistory{}, &models.ItemRequestHistoryAt{},
	// )
	migrateAndLog(
	// &inventory_models.ItemRequest{}, &inventory_models.ItemRequestAt{},
	// &inventory_models.ItemRequestDetails{}, &inventory_models.ItemRequestDetailsAt{},
	// &inventory_models.ItemRequestLocations{}, &inventory_models.ItemRequestLocationsAt{},
	)

	migrateAndLog(
	// &inventory_models.PickActivity{}, &inventory_models.PickActivityAt{},
	// &inventory_models.PickActivityDetails{}, &inventory_models.PickActivityDetailsAt{},
	// &inventory_models.PickActivityLocations{}, &inventory_models.PickActivityLocationsAt{},
	)
	// migrateAndLog(
	// 	&models.PickActivity{}, &models.PickActivityAt{},
	// 	&models.PickActivityDetails{}, &models.PickActivityDetailsAt{},
	// 	&models.PickActivityLocation{}, &models.PickActivityLocationAt{},
	// 	&models.PickActivityHistory{}, &models.PickActivityHistoryAt{},
	// )

	// migrateAndLog(
	// 	&inventory_models.ReceivingReport{}, &inventory_models.ReceivingReportAt{},
	// 	&inventory_models.ReceivingReportDetails{}, &inventory_models.ReceivingReportDetailsAt{},
	// )
}

func migrateEngineering() {
	// DB.AutoMigrate(
	// 	&models.ItemRequest{}, &models.ItemRequestAt{},
	// 	&models.ItemRequestDetails{}, &models.ItemRequestDetailsAt{},
	// 	&models.ItemRequestLocation{}, &models.ItemRequestLocationAt{},
	// 	&models.ItemRequestHistory{}, &models.ItemRequestHistoryAt{},
	// )
	// DB.AutoMigrate(
	// 	&models.PickActivity{}, &models.PickActivityAt{},
	// 	&models.PickActivityDetails{}, &models.PickActivityDetailsAt{},
	// 	&models.PickActivityLocation{}, &models.PickActivityLocationAt{},
	// 	&models.PickActivityHistory{}, &models.PickActivityHistoryAt{},
	// )
}

// ============================================
// SALES & CRM
// ============================================
func migrateSalesCrm() {
	// fmt.Println("=== Migrating SALES & CRM Module ===")
	// migrateAndLog(
	// 	&models.Order{}, &models.OrderAt{},
	// 	&models.OrderDetails{}, &models.OrderDetailsAt{},
	// 	&models.CRM{}, &models.CRMAt{},
	// 	&models.Status{}, &models.StatusAt{},
	// 	&models.Opportunity{}, &models.OpportunityAt{},
	// 	&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{},
	// 	&models.SalesQuotationSelectedImage{}, &models.SalesQuotationSelectedImageAt{},
	// )
	// migrateAndLog(&models.SalesCanvasSheet{})

	// Only this pair is active for now - contact_1/contact_2 (SalesQuotationContent) were
	// missing an explicit column tag, so applyQuotationFieldChanges' hand-built update map
	// (which uses the literal keys "contact_1"/"contact_2") didn't match whatever column
	// name GORM would otherwise assume, causing "Invalid column name 'contact_1'" on save.
	// The rest of the Sales & CRM module above stays untouched/commented out.
	migrateAndLog(
		&models.SalesQuotation{}, &models.SalesQuotationAt{},
	)

	// OrderDetailsContent gained ItemSetHeader (item_set_header) so project-sourced
	// order lines can carry their itemset tab name for print reconstruction. AutoMigrate
	// is additive-only (adds the missing column, doesn't touch existing rows/data), so
	// running this on every startup is safe - same reasoning as migrateSalesProject.
	migrateAndLog(
		&models.OrderDetails{}, &models.OrderDetailsAt{},
	)
}

// ============================================
// RETURNS (SALES & PURCHASE) - spec §5.13, §5.8
// ============================================
// New models, no existing data to disturb. AutoMigrate creates these four
// tables (+ their _at audit pairs) the first time this runs and is a no-op
// on every run after - same additive-only guarantee as every other
// migrateAndLog call in this file.
func migrateReturns() {
	migrateAndLog(
		&models.SalesReturn{}, &models.SalesReturnAt{},
		&models.SalesReturnDetails{}, &models.SalesReturnDetailsAt{},
		&models.PurchaseReturn{}, &models.PurchaseReturnAt{},
		&models.PurchaseReturnDetails{}, &models.PurchaseReturnDetailsAt{},
	)
}

// ============================================
// MEMOS (CREDIT & DEBIT) - spec §5.18, §5.19
// ============================================
// New models, no existing data to disturb - same additive-only guarantee as
// migrateReturns() above. This was missing entirely (models/service/handler/
// routes/client UI all existed, but nothing ever created these tables), so
// every Credit Memo/Debit Memo API call has been failing against a real
// database since those were built - found while setting up Phase 1's
// integration testing (item 1.10).
func migrateMemos() {
	migrateAndLog(
		&models.CreditMemo{}, &models.CreditMemoAt{},
		&models.DebitMemo{}, &models.DebitMemoAt{},
		&models.DebitMemoDetails{}, &models.DebitMemoDetailsAt{},
	)
}

// ============================================
// SALES PROJECT
// ============================================
func migrateSalesProject() {
	fmt.Println("=== Migrating SALES PROJECT Module ===")
	// Was fully commented out, so the live DB never picked up schema changes made to
	// these models over time (e.g. SalesProjectContent.IsWiring, added to the struct but
	// never applied to the actual table - "Invalid column name 'is_wiring'"). AutoMigrate
	// is additive-only (creates missing tables/columns/indexes, never drops or alters
	// existing data), and migrateAndLog logs+continues per model instead of aborting on
	// a single failure, so re-enabling this is safe to run on every startup - it'll pick
	// up is_wiring here and self-heals the same way for any other drift in this module.
	migrateAndLog(
		&models.SalesProjectTemplate{}, &models.SalesProjectTemplateAt{},
		&models.SalesProjectTemplateChild{}, &models.SalesProjectTemplateChildAt{},
		&models.SalesProjectMultiplier{}, &models.SalesProjectMultiplierAt{},
		&models.SalesProjectHistory{}, &models.SalesProjectHistoryAt{},
		&models.SalesProjectItemSet{}, &models.SalesProjectItemSetAt{},
		&models.SalesProjectContent{}, &models.SalesProjectContentAt{},
		&models.SalesProjectContentFinal{}, &models.SalesProjectContentFinalAt{},
		&models.SalesProjectAdvancedConditions{}, &models.SalesProjectAdvancedConditionsAt{},
		&models.SalesProjectItems{}, &models.SalesProjectItemsAt{},
		&models.SalesProjectWiring{}, &models.SalesProjectWiringAt{},
	)
}

// ============================================
// PURCHASING & VENDOR
// ============================================
func migratePurchasingVendor() {
	// fmt.Println("=== Migrating PURCHASING & VENDOR Module ===")
	// migrateAndLog(
	// 	&models.PurchaseRequisition{}, &models.PurchaseRequisitionAt{},
	// 	&models.PROrders{}, &models.PROrdersAt{},
	// 	&models.PurchasingCanvassSheet{}, &models.PurchasingCanvassSheetAt{},
	// 	&models.PurchaseOrder{}, &models.PurchaseOrderAt{},
	// 	&models.PurchaseOrderDetails{}, &models.PurchaseOrderDetailsAt{},
	// )
}

// ============================================
// BUSINESS PARTNER / BPI
// ============================================
func migrateBpi() {
	// fmt.Println("=== Migrating BPI Module ===")
	// migrateAndLog(
	// 	&models.Bpi{}, &models.BpiAt{},
	// 	&models.BpiGeneral{}, &models.BpiGeneralAt{},
	// 	&models.BpiContacts{}, &models.BpiContactsAt{},
	// 	&models.BpiIndustries{}, &models.BpiIndustriesAt{},
	// 	&models.Industries{}, &models.IndustriesAt{},
	// 	&models.BpiBranchIndustries{}, &models.BpiBranchIndustriesAt{},
	// 	&models.BpiEntity{}, &models.BpiEntityAt{},
	// 	&models.BpiAddress{}, &models.BpiAddressAt{},
	// 	&models.BpiItems{}, &models.BpiItemsAt{},
	// 	&models.BpiFinance{}, &models.BpiFinanceAt{},
	// 	&models.BpiAccreditation{}, &models.BpiAccreditationAt{},
	// 	&models.BpiHistory{}, &models.BpiHistoryAt{},
	// )
	// Only this pair is active for now - z_tbl_accounting_bpi_overpayment_at
	// was missing AT_USER (and likely other At-embedded columns), causing
	// "Invalid column name 'AT_USER'" on insert. The rest of the BPI module
	// above stays untouched/commented out.
	migrateAndLog(
		&accounting_models.BpiOverpayment{}, &accounting_models.BpiOverpaymentAt{},
	)
}

// ============================================
// ACCOUNTING
// ============================================
func migrateAccounting() {
	// fmt.Println("=== Migrating ACCOUNTING Module ===")
	migrateAndLog(
		&accounting_models.ChartClass{}, &accounting_models.ChartClassAt{},
	)
	// migrateAndLog(
	// 	&accounting_models.ChartOfAccounts{}, &accounting_models.ChartOfAccountsAt{},
	// 	&accounting_models.Tax{}, &accounting_models.TaxAt{},
	// 	&accounting_models.TaxDetails{}, &accounting_models.TaxDetailsAt{},
	// )
	// migrateAndLog(
	// 	&accounting_models.SalesInvoice{}, &accounting_models.SalesInvoiceAt{},
	// 	&accounting_models.SalesInvoiceDetail{}, &accounting_models.SalesInvoiceDetailAt{},
	// 	&accounting_models.SalesInvoice2{}, &accounting_models.SalesInvoice2At{},
	// 	&accounting_models.SalesInvoiceDetails2{}, &accounting_models.SalesInvoiceDetails2At{},
	// )
	// migrateAndLog(
	// 	&accounting_models.JournalEntry{}, &accounting_models.JournalEntryAt{},
	// 	&accounting_models.JournalEntryDetails{}, &accounting_models.JournalEntryDetailsAt{},
	// 	&accounting_models.JournalEntry2{}, &accounting_models.JournalEntry2At{},
	// 	&accounting_models.JournalEntryDetails2{}, &accounting_models.JournalEntryDetails2At{},
	// )
	// migrateAndLog(
	// 	&accounting_models.InvoiceReceipt{}, &accounting_models.InvoiceReceiptAt{},
	// 	&accounting_models.InvoiceReceiptDetails{}, &accounting_models.InvoiceReceiptDetailsAt{},
	// 	&accounting_models.BulkInvoiceReceipt{}, &accounting_models.BulkInvoiceReceiptAt{},
	// 	&accounting_models.BulkInvoiceReceiptDetails{}, &accounting_models.BulkInvoiceReceiptDetailsAt{},
	// )
	// migrateAndLog(
	// 	&accounting_models.ApVoucher{}, &accounting_models.ApVoucherAt{},
	// 	&accounting_models.ApVoucherDetails{}, &accounting_models.ApVoucherDetailsAt{},
	// 	&accounting_models.PaymentVoucher{}, &accounting_models.PaymentVoucherAt{},
	// 	&accounting_models.PaymentVoucherDetails{}, &accounting_models.PaymentVoucherDetailsAt{},
	// )
	// migrateAndLog(
	// 	&accounting_models.PaymentReceipt{}, &accounting_models.PaymentReceiptAt{},
	// 	&accounting_models.PaymentReceiptDetails{}, &accounting_models.PaymentReceiptDetailsAt{},
	// )
}

// ============================================
// JOB ORDER
// ============================================
func migrateJobOrder() {
	fmt.Println("=== Migrating JOB ORDER Module ===")
	migrateAndLog(&models.JobOrder{}, &models.JobOrderAt{})
}

// ============================================
// LOGISTICS & DISPATCHING
// ============================================
func migrateLogisticsDispatching() {
	fmt.Println("=== Migrating LOGISTICS & DISPATCHING Module ===")
	migrateAndLog(&models.ItemRelease{}, &models.ItemReleaseAt{})
	migrateAndLog(&models.ItemReleaseDetails{}, &models.ItemReleaseDetailsAt{})
	migrateAndLog(&dispatching_models.CalendarCategoryModel{}, &dispatching_models.CalendarCategoryAt{})
	migrateAndLog(&dispatching_models.CalendarCostTypeModel{}, &dispatching_models.CalendarCostTypeAt{})
	migrateAndLog(&models.CalendarScheduleModel{}, &models.CalendarScheduleAt{})
	migrateAndLog(&dispatching_models.DeliveryReceipt{}, &dispatching_models.DeliveryReceiptAt{})
	migrateAndLog(&dispatching_models.DeliveryReceiptCosts{}, &dispatching_models.DeliveryReceiptCostsAt{})
	migrateAndLog(&dispatching_models.ReceiptFile{}, &dispatching_models.ReceiptFileAt{})
	migrateAndLog(&dispatching_models.SalesCalendarScheduleModel{}, &dispatching_models.SalesCalendarScheduleModelAt{})
	migrateAndLog(&dispatching_models.EngineeringCalendarScheduleModel{}, &dispatching_models.EngineeringCalendarScheduleModelAt{})
	migrateAndLog(&dispatching_models.LogisticsCalendarScheduleModel{}, &dispatching_models.LogisticsCalendarScheduleModelAt{})
	migrateAndLog(&dispatching_models.LogisticsRoute{}, &dispatching_models.LogisticsRouteAt{})
	migrateAndLog(&dispatching_models.LogisticsRouteCost{})
}

// ============================================
// VEHICLE MANAGEMENT
// ============================================
func migrateVehicleManagement() {
	// fmt.Println("=== Migrating VEHICLE MANAGEMENT Module ===")
	// migrateAndLog(&models.VehicleModel{}, &models.VehicleAt{})
	// migrateAndLog(&models.VehicleFileModel{}, &models.VehicleFileAt{})
}

func migrateAndLog(models ...interface{}) {
	for _, m := range models {
		name := getModelName(m)

		err := DB.AutoMigrate(m)
		if err != nil {
			fmt.Println("❌ Failed:", name, "| Error:", err)
		} else {
			fmt.Println("✅ Migrated:", name)
		}
	}
}

func getModelName(m interface{}) string {
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
