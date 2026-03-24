package initializers

import "strings"

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
	// DB.AutoMigrate(
	// 	&models.User{}, &models.UserAt{},
	// 	&models.PositionModel{}, &models.PositionAt{},
	// 	&models.PositionAccessModel{}, &models.PositionAccessAt{},
	// 	&models.UserPermissionModel{}, &models.UserPermissionAt{},
	// )
	// DB.AutoMigrate(&models.Social{}, &models.SocialAt{})
	// DB.AutoMigrate(&models.Entity{}, &models.EntityAt{})
	// DB.AutoMigrate(&models.Application{}, &models.ApplicationAt{})
	// DB.AutoMigrate(&models.CompanyModel{}, &models.CompanyAt{})
	// DB.AutoMigrate(&models.CompanyAddressModel{}, &models.CompanyAddressAt{})
	// DB.AutoMigrate(&models.CompanyContactModel{}, &models.CompanyContactAt{})
}

// ============================================
// SETUP / MASTER DATA
// ============================================
func migrateSetup() {
	// DB.AutoMigrate(
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
	// DB.AutoMigrate(
	// 	&models.Item{}, &models.ItemAt{},
	// 	&models.ItemSpecs{}, &models.ItemSpecsAt{},
	// 	&models.ItemSpecsTemplate{}, &models.ItemSpecsTemplateAt{},
	// 	&models.AdditionalSpecs{}, &models.AdditionalSpecsAt{},
	// 	&models.ItemImage{}, &models.ItemImageAt{},
	// 	&models.ItemInventory{}, &models.ItemInventoryAt{},
	// 	&models.Model{}, &models.ModelAt{},
	// 	&models.PumpCount{}, &models.PumpCountAt{},
	// 	&models.ItemTradeType{}, &models.ItemTradeTypeAt{},
	// )
}

// ============================================
// BOM & BOQ MANAGEMENT
// ============================================
func migrateBomBoq() {
	// DB.AutoMigrate(
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
	// DB.AutoMigrate(
	// 	&models.WarehouseUseType{}, &models.WarehouseUseTypeAt{},
	// 	&models.WarehouseName{}, &models.WarehouseNameAt{},
	// 	&models.WarehouseAddress{}, &models.WarehouseAddressAt{},
	// 	&models.WarehouseArea{}, &models.WarehouseAreaAt{},
	// )
	// DB.AutoMigrate(
	// 	&models.InvTracker{}, &models.InvTrackerAt{},
	// 	&models.InventoryStocks{}, &models.InventoryStocksAt{},
	// 	&models.InventoryStocksHistory{}, &models.InventoryStocksHistoryAt{},
	// )
	// DB.AutoMigrate(
	// 	&models.ReceivingReport{}, &models.ReceivingReportAt{},
	// 	&models.ReceivingReportDetails{}, &models.ReceivingReportDetailsAt{},
	// 	&models.ReceivingReportInventory{}, &models.ReceivingReportInventoryAt{},
	// )
	// DB.AutoMigrate(
	// 	&models.ReceivingReport2{}, &models.ReceivingReportAt2{},
	// 	&models.ReceivingReportDetails2{}, &models.ReceivingReportDetailsAt2{},
	// 	&models.ReceivingHistory{}, &models.ReceivingHistoryAt{},
	// )
}

// ============================================
// ITEM REQUEST & PICK ACTIVITY
// ============================================
func migrateItemRequestPickActivity() {
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
	// DB.AutoMigrate(
	// 	&models.Order{}, &models.OrderAt{},
	// 	&models.OrderDetails{}, &models.OrderDetailsAt{},
	// 	&models.CRM{}, &models.CRMAt{},
	// 	&models.Status{}, &models.StatusAt{},
	// 	&models.Opportunity{}, &models.OpportunityAt{},
	// )
	// DB.AutoMigrate(
	// 	&models.SalesQuotation{}, &models.SalesQuotationAt{},
	// 	&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{},
	// 	&models.SalesQuotationSelectedImage{}, &models.SalesQuotationSelectedImageAt{},
	// )
	// DB.AutoMigrate(&models.SalesCanvasSheet{})
}

// ============================================
// SALES PROJECT
// ============================================
func migrateSalesProject() {
	// DB.AutoMigrate(
	// 	&models.SalesProjectTemplate{}, &models.SalesProjectTemplateAt{},
	// 	&models.SalesProjectTemplateChild{}, &models.SalesProjectTemplateChildAt{},
	// 	&models.SalesProjectMultiplier{}, &models.SalesProjectMultiplierAt{},
	// 	&models.SalesProjectHistory{}, &models.SalesProjectHistoryAt{},
	// 	&models.SalesProjectItemSet{}, &models.SalesProjectItemSetAt{},
	// 	&models.SalesProjectContent{}, &models.SalesProjectContentAt{},
	// 	&models.SalesProjectContentFinal{}, &models.SalesProjectContentFinalAt{},
	// 	&models.SalesProjectAdvancedConditions{}, &models.SalesProjectAdvancedConditionsAt{},
	// 	&models.SalesProjectItems{}, &models.SalesProjectItemsAt{},
	// 	&models.SalesProjectWiring{}, &models.SalesProjectWiringAt{},
	// )
}

// ============================================
// PURCHASING & VENDOR
// ============================================
func migratePurchasingVendor() {
	// DB.AutoMigrate(
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
	// DB.AutoMigrate(
	// 	&models.Bpi{}, &models.BpiAt{},
	// 	&models.BpiGeneral{}, &models.BpiGeneralAt{},
	// 	&models.BpiContacts{}, &models.BpiContactsAt{},
	// 	&models.BpiIndustries{}, &models.BpiIndustriesAt{},
	// 	&models.BpiBranchIndustries{}, &models.BpiBranchIndustriesAt{},
	// 	&models.BpiEntity{}, &models.BpiEntityAt{},
	// 	&models.BpiAddress{}, &models.BpiAddressAt{},
	// 	&models.BpiItems{}, &models.BpiItemsAt{},
	// 	&models.BpiFinance{}, &models.BpiFinanceAt{},
	// 	&models.BpiAccreditation{}, &models.BpiAccreditationAt{},
	// 	&models.BpiHistory{}, &models.BpiHistoryAt{},
	// 	&accounting_models.BpiOverpayment{}, &accounting_models.BpiOverpaymentAt{},
	// )
}

// ============================================
// ACCOUNTING
// ============================================
func migrateAccounting() {
	// DB.AutoMigrate(
	// 	&accounting_models.ChartOfAccounts{}, &accounting_models.ChartOfAccountsAt{},
	// 	&accounting_models.Tax{}, &accounting_models.TaxAt{},
	// 	&accounting_models.TaxDetails{}, &accounting_models.TaxDetailsAt{},
	// )
	// DB.AutoMigrate(
	// 	&accounting_models.SalesInvoice{}, &accounting_models.SalesInvoiceAt{},
	// 	&accounting_models.SalesInvoiceDetail{}, &accounting_models.SalesInvoiceDetailAt{},
	// 	&accounting_models.SalesInvoice2{}, &accounting_models.SalesInvoice2At{},
	// 	&accounting_models.SalesInvoiceDetails2{}, &accounting_models.SalesInvoiceDetails2At{},
	// )
	// DB.AutoMigrate(
	// 	&accounting_models.JournalEntry{}, &accounting_models.JournalEntryAt{},
	// 	&accounting_models.JournalEntryDetails{}, &accounting_models.JournalEntryDetailsAt{},
	// 	&accounting_models.JournalEntry2{}, &accounting_models.JournalEntry2At{},
	// 	&accounting_models.JournalEntryDetails2{}, &accounting_models.JournalEntryDetails2At{},
	// )
	// DB.AutoMigrate(
	// 	&accounting_models.InvoiceReceipt{}, &accounting_models.InvoiceReceiptAt{},
	// 	&accounting_models.InvoiceReceiptDetails{}, &accounting_models.InvoiceReceiptDetailsAt{},
	// 	&accounting_models.BulkInvoiceReceipt{}, &accounting_models.BulkInvoiceReceiptAt{},
	// 	&accounting_models.BulkInvoiceReceiptDetails{}, &accounting_models.BulkInvoiceReceiptDetailsAt{},
	// )
	// DB.AutoMigrate(
	// 	&accounting_models.ApVoucher{}, &accounting_models.ApVoucherAt{},
	// 	&accounting_models.ApVoucherDetails{}, &accounting_models.ApVoucherDetailsAt{},
	// 	&accounting_models.PaymentVoucher{}, &accounting_models.PaymentVoucherAt{},
	// 	&accounting_models.PaymentVoucherDetails{}, &accounting_models.PaymentVoucherDetailsAt{},
	// )
	// DB.AutoMigrate(
	// 	&accounting_models.PaymentReceipt{}, &accounting_models.PaymentReceiptAt{},
	// 	&accounting_models.PaymentReceiptDetails{}, &accounting_models.PaymentReceiptDetailsAt{},
	// )
}

// ============================================
// JOB ORDER
// ============================================
func migrateJobOrder() {
	// DB.AutoMigrate(&models.JobOrder{}, &models.JobOrderAt{})
}

// ============================================
// LOGISTICS & DISPATCHING
// ============================================
func migrateLogisticsDispatching() {
	// DB.AutoMigrate(&models.ItemRelease{}, &models.ItemReleaseAt{})
	// DB.AutoMigrate(&models.ItemReleaseDetails{}, &models.ItemReleaseDetailsAt{})
	// DB.AutoMigrate(&dispatching_models.CalendarCategoryModel{}, &dispatching_models.CalendarCategoryAt{})
	// DB.AutoMigrate(&dispatching_models.CalendarCostTypeModel{}, &dispatching_models.CalendarCostTypeAt{})
	// DB.AutoMigrate(&models.CalendarScheduleModel{}, &models.CalendarScheduleAt{})
	// DB.AutoMigrate(
	// 	&dispatching_models.DeliveryReceipt{}, &dispatching_models.DeliveryReceiptAt{},
	// 	&dispatching_models.DeliveryReceiptItems{}, &dispatching_models.DeliveryReceiptItemsAt{},
	// 	&dispatching_models.DeliveryReceiptCosts{}, &dispatching_models.DeliveryReceiptCostsAt{},
	// 	&dispatching_models.ReceiptFile{}, &dispatching_models.ReceiptFileAt{},
	// )
}

// ============================================
// VEHICLE MANAGEMENT
// ============================================
func migrateVehicleManagement() {
	// DB.AutoMigrate(&models.VehicleModel{}, &models.VehicleAt{})
	// DB.AutoMigrate(&models.VehicleFileModel{}, &models.VehicleFileAt{})
}
