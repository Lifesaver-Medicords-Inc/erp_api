package initializers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
)

// models "github.com/pierceperado/smpc/models"
// accounting_models "github.com/pierceperado/smpc/models/accounting_models"
// dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"

// MigrateAll runs all database migrations - Use for deployment
func MigrateAll() {
	migrateAdmin()
	migrateSetup()
	migrateItemManagement()
	migrateBomBoq()
	migrateInventoryWarehouse()
	migrateItemRequestPickActivity()
	migrateSalesCrm()
	migrateSalesProject()
	migratePurchasingVendor()
	migrateBpi()
	migrateAccounting()
	migrateJobOrder()
	migrateLogisticsDispatching()
	migrateVehicleManagement()
}

// MigrateModel migrates specific models or categories by name
// Usage: MigrateModel("admin"), MigrateModel("item", "sales"), MigrateModel("all")
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
		case "item_request", "pick_activity", "item_request_pick":
			migrateItemRequestPickActivity()
		case "sales", "crm", "sales_crm":
			migrateSalesCrm()
		case "sales_project":
			migrateSalesProject()
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
	fmt.Println("=== Migrating BPI Module ===")

	migrateAndLog(
		&models.User{}, &models.UserAt{},
		&models.PositionModel{}, &models.PositionAt{},
		&models.PositionAccessModel{}, &models.PositionAccessAt{},
		&models.UserPermissionModel{}, &models.UserPermissionAt{},
		&models.Social{}, &models.SocialAt{},
		&models.Entity{}, &models.EntityAt{},
		&models.Application{}, &models.ApplicationAt{},
		&models.CompanyModel{}, &models.CompanyAt{},
		&models.CompanyAddressModel{}, &models.CompanyAddressAt{},
		&models.CompanyContactModel{}, &models.CompanyContactAt{},
	)
}

// ============================================
// SETUP / MASTER DATA
// ============================================
func migrateSetup() {

	fmt.Println("=== Migrating SETUP Module ===")

	migrateAndLog(
		&models.Class{}, &models.ClassAt{},
		&models.Name{}, &models.NameAt{},
		&models.Type{}, &models.TypeAt{},
		&models.Material{}, &models.MaterialAt{},
		&models.ValuationMethod{}, &models.ValuationMethodAt{},
		&models.TradeType{}, &models.TradeTypeAt{},
		&models.Brand{}, &models.BrandAt{},
		&models.UnitMeasurement{}, &models.UnitMeasurementAt{},
		&models.PaymentTerms{}, &models.PaymentTermsAt{},
		&models.Class{}, &models.ClassAt{},
		&models.Name{}, &models.NameAt{},
		&models.Type{}, &models.TypeAt{},
		&models.Material{}, &models.MaterialAt{},
		&models.ValuationMethod{}, &models.ValuationMethodAt{},
		&models.TradeType{}, &models.TradeTypeAt{},
	)
}

// ============================================
// ITEM MANAGEMENT
// ============================================
func migrateItemManagement() {
	fmt.Println("=== Migrating ITEM MANAGEMENT Module ===")

	migrateAndLog(
		&models.ItemSpecsTemplate{}, &models.ItemSpecsTemplateAt{},
		&models.AdditionalSpecs{}, &models.AdditionalSpecsAt{},
		&models.AdditionalSpecsPumpType{}, &models.AdditionalSpecsPumpTypeAt{},
		&models.ItemImage{}, &models.ItemImageAt{},
		&models.ItemInventory{}, &models.ItemInventoryAt{},
		&models.Model{}, &models.ModelAt{},
		&models.PumpType{}, &models.PumpTypeAt{},
		&models.PumpCount{}, &models.PumpCountAt{},
		&models.ItemTradeType{}, &models.ItemTradeTypeAt{},
	)
}

// ============================================
// BOM & BOQ MANAGEMENT
// ============================================
func migrateBomBoq() {
	fmt.Println("=== Migrating BOM & BOQ Module ===")
	migrateAndLog(
		&models.SetupItemBom{}, &models.SetupItemBomAt{},
		&models.SetupItemBomDetails{}, &models.SetupItemBomDetailsAt{},
		&models.ItemBoq{}, &models.ItemBoqAt{},
		&models.ItemBoqDetails{}, &models.ItemBoqDetailsAt{},
		&models.BoqNotes{}, &models.BoqNotesAt{},
		&models.WiringUserInput{}, &models.WiringUserInputAt{},
	)
}

// ============================================
// INVENTORY & WAREHOUSE
// ============================================
func migrateInventoryWarehouse() {
	fmt.Println("=== Migrating INVENTORY & WAREHOUSE Module ===")
	migrateAndLog(
		&models.WarehouseUseType{}, &models.WarehouseUseTypeAt{},
		&models.WarehouseName{}, &models.WarehouseNameAt{},
		&models.WarehouseAddress{}, &models.WarehouseAddressAt{},
		&models.WarehouseArea{}, &models.WarehouseAreaAt{},
	)
	migrateAndLog(
		&models.InvTracker{}, &models.InvTrackerAt{},
		&models.InventoryStocks{}, &models.InventoryStocksAt{},
		&models.InventoryStocksHistory{}, &models.InventoryStocksHistoryAt{},
	)
	migrateAndLog(
		&models.ReceivingReport{}, &models.ReceivingReportAt{},
		&models.ReceivingReportDetails{}, &models.ReceivingReportDetailsAt{},
		&models.ReceivingReportInventory{}, &models.ReceivingReportInventoryAt{},
	)
	migrateAndLog(
		&models.ReceivingReport2{}, &models.ReceivingReportAt2{},
		&models.ReceivingReportDetails2{}, &models.ReceivingReportDetailsAt2{},
		&models.ReceivingHistory{}, &models.ReceivingHistoryAt{},
	)
}

// ============================================
// ITEM REQUEST & PICK ACTIVITY
// ============================================
func migrateItemRequestPickActivity() {
	fmt.Println("=== Migrating ITEM REQUEST & PICK ACTIVITY Module ===")
	migrateAndLog(
		&models.ItemRequest{}, &models.ItemRequestAt{},
		&models.ItemRequestDetails{}, &models.ItemRequestDetailsAt{},
		&models.ItemRequestLocation{}, &models.ItemRequestLocationAt{},
		&models.ItemRequestHistory{}, &models.ItemRequestHistoryAt{},
	)
	migrateAndLog(
		&models.PickActivity{}, &models.PickActivityAt{},
		&models.PickActivityDetails{}, &models.PickActivityDetailsAt{},
		&models.PickActivityLocation{}, &models.PickActivityLocationAt{},
		&models.PickActivityHistory{}, &models.PickActivityHistoryAt{},
	)
}

// ============================================
// SALES & CRM
// ============================================
func migrateSalesCrm() {
	fmt.Println("=== Migrating SALES & CRM Module ===")
	migrateAndLog(
		&models.Order{}, &models.OrderAt{},
		&models.OrderDetails{}, &models.OrderDetailsAt{},
		&models.CRM{}, &models.CRMAt{},
		&models.Status{}, &models.StatusAt{},
		&models.Opportunity{}, &models.OpportunityAt{},
		&models.SalesQuotation{}, &models.SalesQuotationAt{},
		&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{},
		&models.SalesQuotationSelectedImage{}, &models.SalesQuotationSelectedImageAt{},
	)
	migrateAndLog(&models.SalesCanvasSheet{})
}

// ============================================
// SALES PROJECT
// ============================================
func migrateSalesProject() {
	fmt.Println("=== Migrating SALES PROJECT Module ===")
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
	fmt.Println("=== Migrating PURCHASING & VENDOR Module ===")
	migrateAndLog(
		&models.PurchaseRequisition{}, &models.PurchaseRequisitionAt{},
		&models.PROrders{}, &models.PROrdersAt{},
		&models.PurchasingCanvassSheet{}, &models.PurchasingCanvassSheetAt{},
		&models.PurchaseOrder{}, &models.PurchaseOrderAt{},
		&models.PurchaseOrderDetails{}, &models.PurchaseOrderDetailsAt{},
	)
}

// ============================================
// BUSINESS PARTNER / BPI
// ============================================
func migrateBpi() {
	fmt.Println("=== Migrating BPI Module ===")
	migrateAndLog(
		&models.Bpi{}, &models.BpiAt{},
		&models.BpiGeneral{}, &models.BpiGeneralAt{},
		&models.BpiContacts{}, &models.BpiContactsAt{},
		&models.BpiIndustries{}, &models.BpiIndustriesAt{},
		&models.Industries{}, &models.IndustriesAt{},
		&models.BpiBranchIndustries{}, &models.BpiBranchIndustriesAt{},
		&models.BpiEntity{}, &models.BpiEntityAt{},
		&models.BpiAddress{}, &models.BpiAddressAt{},
		&models.BpiItems{}, &models.BpiItemsAt{},
		&models.BpiFinance{}, &models.BpiFinanceAt{},
		&models.BpiAccreditation{}, &models.BpiAccreditationAt{},
		&models.BpiHistory{}, &models.BpiHistoryAt{},
		&accounting_models.BpiOverpayment{}, &accounting_models.BpiOverpaymentAt{},
	)
}

// ============================================
// ACCOUNTING
// ============================================
func migrateAccounting() {
	fmt.Println("=== Migrating ACCOUNTING Module ===")
	migrateAndLog(
		&accounting_models.ChartOfAccounts{}, &accounting_models.ChartOfAccountsAt{},
		&accounting_models.Tax{}, &accounting_models.TaxAt{},
		&accounting_models.TaxDetails{}, &accounting_models.TaxDetailsAt{},
	)
	migrateAndLog(
		&accounting_models.SalesInvoice{}, &accounting_models.SalesInvoiceAt{},
		&accounting_models.SalesInvoiceDetail{}, &accounting_models.SalesInvoiceDetailAt{},
		&accounting_models.SalesInvoice2{}, &accounting_models.SalesInvoice2At{},
		&accounting_models.SalesInvoiceDetails2{}, &accounting_models.SalesInvoiceDetails2At{},
	)
	migrateAndLog(
		&accounting_models.JournalEntry{}, &accounting_models.JournalEntryAt{},
		&accounting_models.JournalEntryDetails{}, &accounting_models.JournalEntryDetailsAt{},
		&accounting_models.JournalEntry2{}, &accounting_models.JournalEntry2At{},
		&accounting_models.JournalEntryDetails2{}, &accounting_models.JournalEntryDetails2At{},
	)
	migrateAndLog(
		&accounting_models.InvoiceReceipt{}, &accounting_models.InvoiceReceiptAt{},
		&accounting_models.InvoiceReceiptDetails{}, &accounting_models.InvoiceReceiptDetailsAt{},
		&accounting_models.BulkInvoiceReceipt{}, &accounting_models.BulkInvoiceReceiptAt{},
		&accounting_models.BulkInvoiceReceiptDetails{}, &accounting_models.BulkInvoiceReceiptDetailsAt{},
	)
	migrateAndLog(
		&accounting_models.ApVoucher{}, &accounting_models.ApVoucherAt{},
		&accounting_models.ApVoucherDetails{}, &accounting_models.ApVoucherDetailsAt{},
		&accounting_models.PaymentVoucher{}, &accounting_models.PaymentVoucherAt{},
		&accounting_models.PaymentVoucherDetails{}, &accounting_models.PaymentVoucherDetailsAt{},
	)
	migrateAndLog(
		&accounting_models.PaymentReceipt{}, &accounting_models.PaymentReceiptAt{},
		&accounting_models.PaymentReceiptDetails{}, &accounting_models.PaymentReceiptDetailsAt{},
	)
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
	migrateAndLog(
		&dispatching_models.DeliveryReceipt{}, &dispatching_models.DeliveryReceiptAt{},
		&dispatching_models.DeliveryReceiptItems{}, &dispatching_models.DeliveryReceiptItemsAt{},
		&dispatching_models.DeliveryReceiptCosts{}, &dispatching_models.DeliveryReceiptCostsAt{},
		&dispatching_models.ReceiptFile{}, &dispatching_models.ReceiptFileAt{},
	)
}

// ============================================
// VEHICLE MANAGEMENT
// ============================================
func migrateVehicleManagement() {
	fmt.Println("=== Migrating VEHICLE MANAGEMENT Module ===")
	migrateAndLog(&models.VehicleModel{}, &models.VehicleAt{})
	migrateAndLog(&models.VehicleFileModel{}, &models.VehicleFileAt{})
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
