package initializers

import "github.com/pierceperado/smpc/models"

func MigrateDb() {
	// // Auth

	// DB.AutoMigrate(&models.User{}, &models.UserAt{})
	// DB.AutoMigrate(&models.User{}, &models.UserAt{})

	// // Setup

	// DB.AutoMigrate(&models.Brand{}, &models.BrandAt{})
	// DB.AutoMigrate(&models.UnitMeasurement{}, &models.UnitMeasurementAt{})
	// DB.AutoMigrate(&models.PaymentTerms{}, &models.PaymentTermsAt{})

	// DB.AutoMigrate(&models.Class{}, &models.ClassAt{})
	// DB.AutoMigrate(&models.Name{}, &models.NameAt{})
	// DB.AutoMigrate(&models.Type{}, &models.TypeAt{})
	// DB.AutoMigrate(&models.PositionModel{}, &models.PositionAt{})
	// DB.AutoMigrate(&models.SetupItemBom{}, &models.SetupItemBomAt{})
	// DB.AutoMigrate(&models.SetupItemBomDetails{}, &models.SetupItemBomDetailsAt{})
	// DB.AutoMigrate(&models.ItemBoq{}, &models.ItemBoqAt{})
	// DB.AutoMigrate(&models.ItemBoqDetails{}, &models.ItemBoqDetailsAt{})
	// DB.AutoMigrate(&models.WiringUserInput{}, &models.WiringUserInputAt{})
	// DB.AutoMigrate(&models.BoqNotes{}, &models.BoqNotesAt{})
	// DB.AutoMigrate(&models.Item{}, &models.ItemAt{})
	// DB.AutoMigrate(&models.ItemSpecs{}, &models.ItemSpecsAt{})
	// DB.AutoMigrate(&models.AdditionalSpecs{}, &models.AdditionalSpecsAt{})
	// DB.AutoMigrate(&models.ItemImage{}, &models.ItemImageAt{})
	// DB.AutoMigrate(&models.TradeType{}, models.TradeTypeAt{})
	// DB.AutoMigrate(&models.Model{}, &models.ModelAt{})
	//DB.AutoMigrate(&models.ItemRequest{}, &models.ItemRequestAt{})
	//DB.AutoMigrate(&models.ItemRequestDetails{}, &models.ItemRequestDetailsAt{})
	//DB.AutoMigrate(&models.ItemRequestLocation{}, &models.ItemRequestLocationAt{})
	//DB.AutoMigrate(&models.ItemRequestHistory{}, &models.ItemRequestHistoryAt{})
	//DB.AutoMigrate(&models.PickActivity{}, &models.PickActivityAt{})
	//DB.AutoMigrate(&models.PickActivityDetails{}, &models.PickActivityDetailsAt{})
	//DB.AutoMigrate(&models.PickActivityLocation{}, &models.PickActivityLocationAt{})
	//DB.AutoMigrate(&models.PickActivityHistory{}, &models.PickActivityHistoryAt{})
	//DB.AutoMigrate(&models.InventoryStocks{}, &models.InventoryStocksAt{})
	//DB.AutoMigrate(&models.InventoryStocksHistory{}, &models.InventoryStocksHistoryAt{})

	// //warehouse
	// DB.AutoMigrate(&models.WarehouseUseType{}, &models.WarehouseUseTypeAt{})
	// DB.AutoMigrate(&models.WarehouseName{}, &models.WarehouseNameAt{})
	// DB.AutoMigrate(&models.WarehouseAddress{}, &models.WarehouseAddressAt{})
	// DB.AutoMigrate(&models.WarehouseArea{}, &models.WarehouseAreaAt{})

	// DB.AutoMigrate(&models.ReceivingReport{}, &models.ReceivingReportAt{})
	// DB.AutoMigrate(&models.ReceivingReportDetails{}, &models.ReceivingReportDetailsAt{})
	// DB.AutoMigrate(&models.ReceivingReportInventory{}, &models.ReceivingReportInventoryAt{})

	// DB.AutoMigrate(&models.ReceivingReport2{}, &models.ReceivingReportAt2{})
	// DB.AutoMigrate(&models.ReceivingReportDetails2{}, &models.ReceivingReportDetailsAt2{})
	// DB.AutoMigrate(&models.ReceivingHistory{}, &models.ReceivingHistoryAt{})

	// DB.AutoMigrate(&models.Social{}, &models.SocialAt{})
	// DB.AutoMigrate(&models.Entity{}, &models.EntityAt{})
	// DB.AutoMigrate(&models.Application{}, &models.ApplicationAt{})
	// DB.AutoMigrate(&models.User{}, &models.UserAt{})

	// //inventory
	// DB.AutoMigrate(&models.InvTracker{}, &models.InvTrackerAt{})

	// // Setup

	// DB.AutoMigrate(&models.Brand{}, &models.BrandAt{})
	// DB.AutoMigrate(&models.UnitMeasurement{}, &models.UnitMeasurementAt{})
	// DB.AutoMigrate(&models.PaymentTerms{}, &models.PaymentTermsAt{})

	// DB.AutoMigrate(&models.Class{}, &models.ClassAt{})
	// DB.AutoMigrate(&models.Name{}, &models.NameAt{})
	// DB.AutoMigrate(&models.Type{}, &models.TypeAt{})
	// DB.AutoMigrate(&models.Item{}, &models.ItemAt{})
	// DB.AutoMigrate(&models.Model{}, &models.ModelAt{})
	//DB.AutoMigrate(&models.ValuationMethod{}, &models.ValuationMethodAt{})

	// DB.AutoMigrate(&models.Application{}, &models.ApplicationAt{})

	// // // Sales

	//DB.AutoMigrate(&models.Order{}, &models.OrderAt{})
	//DB.AutoMigrate(&models.OrderDetails{}, &models.OrderDetailsAt{})
	// DB.AutoMigrate(&models.Opportunity{}, &models.OpportunityAt{})

	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})

	// DB.AutoMigrate(&models.Order{}, &models.OrderAt{})

	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})

	//DB.AutoMigrate(&models.SalesProjectTemplate{}, &models.SalesProjectTemplateAt{})
	//DB.AutoMigrate(&models.SalesProjectTemplateChild{}, &models.SalesProjectTemplateChildAt{})

	// DB.AutoMigrate(
	// 	&models.SalesProjectMultiplier{},
	// 	&models.SalesProjectMultiplierAt{},
	// 	&models.SalesProjectHistory{},
	// 	&models.SalesProjectHistoryAt{},
	// 	&models.SalesProjectItemSet{},
	// 	&models.SalesProjectItemSetAt{},
	// 	&models.SalesProjectContent{},
	// 	&models.SalesProjectContentAt{},
	// 	&models.SalesProjectAdvancedConditions{},
	// 	&models.SalesProjectAdvancedConditionsAt{},
	// 	&models.SalesProjectItems{},
	// 	&models.SalesProjectItemsAt{},
	// )

	// DB.AutoMigrate(
	// 	&models.SalesQuotation{},
	// 	&models.SalesQuotationAt{},
	// 	&models.SalesQuotationQuick{},
	// 	&models.SalesQuotationQuickAt{},
	// 	&models.SalesProjectMultiplier{},
	// 	&models.SalesProjectMultiplierAt{},
	// 	&models.SalesProjectHistory{},
	// 	&models.SalesProjectHistoryAt{},
	// 	&models.SalesProjectItemSet{},
	// 	&models.SalesProjectItemSetAt{},
	// 	&models.SalesProjectAdvancedConditions{},
	// 	&models.SalesProjectAdvancedConditionsAt{},
	// 	&models.SalesProjectItems{},
	// 	&models.SalesProjectItemsAt{},
	// 	&models.SalesProjectWiring{},
	// 	&models.SalesProjectWiringAt{},
	// )

	// DB.AutoMigrate(&models.SalesCanvasSheet{})
	// DB.AutoMigrate(&models.Order{}, &models.OrderAt{})

	// DB.AutoMigrate(&models.Order{}, &models.OrderAt{})
	// DB.AutoMigrate(&models.OrderDetails{}, &models.OrderDetailsAt{})
	// DB.AutoMigrate(&models.Opportunity{}, &models.OpportunityAt{})
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})
	// DB.AutoMigrate(&models.Order{}, &models.OrderAt{})

	// DB.AutoMigrate(&models.OrderDetails{}, &models.OrderDetailsAt{})
	// DB.AutoMigrate(&models.CRM{}, &models.CRMAt{})
	// DB.AutoMigrate(&models.Status{}, &models.StatusAt{})
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})
	// DB.AutoMigrate(&models.SalesQuotationSelectedImage{}, &models.SalesQuotationSelectedImageAt{})
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})

	// // Purchasing
	// DB.AutoMigrate(&models.PurchaseRequisition{}, &models.PurchaseRequisitionAt{})
	// DB.AutoMigrate(&models.PROrders{}, &models.PROrdersAt{})
	// DB.AutoMigrate(&models.PurchasingCanvassSheet{}, &models.PurchasingCanvassShInventoryStocksHistory
	// DB.AutoMigrate(&models.PurchaseOrder{}, &models.PurchaseOrderAt{})
	// DB.AutoMigrate(&models.PurchaseOrderDetails{}, &models.PurchaseOrderDetailsAt{})
	// // Sales
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.Order{}, &models.OrderAt{})
	// DB.AutoMigrate(&models.OrderDetails{}, &models.OrderDetailsAt{})
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})

	// DB.AutoMigrate(&models.SalesProjectItems{}, &models.SalesProjectItemsAt{})

	// //BPI
	// DB.AutoMigrate(&models.Bpi{}, &models.BpiAt{})
	// DB.AutoMigrate(&models.BpiGeneral{}, &models.BpiGeneralAt{})
	// DB.AutoMigrate(&models.BpiContacts{}, &models.BpiContactsAt{})
	// DB.AutoMigrate(&models.BpiIndustries{}, &models.BpiIndustriesAt{})
	// DB.AutoMigrate(&models.BpiBranchIndustries{}, &models.BpiBranchIndustriesAt{})
	// DB.AutoMigrate(&models.BpiEntity{}, &models.BpiEntityAt{})
	// DB.AutoMigrate(&models.BpiAddress{}, &models.BpiAddressAt{})
	// DB.AutoMigrate(&models.BpiItems{}, &models.BpiItemsAt{})
	// DB.AutoMigrate(&models.BpiFinance{}, &models.BpiFinanceAt{})
	// DB.AutoMigrate(&models.BpiAccreditation{}, &models.BpiAccreditationAt{})
	// DB.AutoMigrate(&models.BpiHistory{}, &models.BpiHistoryAt{})

	// // ACCOUNTING
	// //DB.AutoMigrate(&models.ChartOfAccounts{}, &models.ChartOfAccountsAt{})
	// DB.AutoMigrate(&accounting_models.SalesInvoice{}, &accounting_models.SalesInvoiceAt{})
	// DB.AutoMigrate(&accounting_models.Tax{}, &accounting_models.TaxAt{})
	// DB.AutoMigrate(&accounting_models.TaxDetails{}, &accounting_models.TaxDetailsAt{})
	// DB.AutoMigrate(&accounting_models.SalesInvoiceDetail{}, &accounting_models.SalesInvoiceDetailAt{})
	// DB.AutoMigrate(&accounting_models.JournalEntry{}, &accounting_models.JournalEntryAt{})
	// DB.AutoMigrate(&accounting_models.JournalEntryDetails{}, &accounting_models.JournalEntryDetailsAt{})

	// // JOB ORDER
	// DB.AutoMigrate(&models.JobOrder{}, &models.JobOrderAt{})

	// //ADMIN
	// DB.AutoMigrate(&models.User{}, &models.UserAt{}, &models.PositionModel{}, &models.PositionAt{}, &models.PositionAccessModel{}, &models.PositionAccessAt{}, &models.UserPermissionModel{}, &models.UserPermissionAt{})
	// DB.AutoMigrate(&models.VehicleModel{}, &models.VehicleAt{})
	// DB.AutoMigrate(&models.VehicleFileModel{}, &models.VehicleFileAt{})
	// DB.AutoMigrate(&models.CompanyModel{}, &models.CompanyAt{}, &models.CompanyAddressModel{}, &models.CompanyAddressAt{}, &models.CompanyContactModel{}, &models.CompanyContactAt{})
	// //JOB ORDER
	// DB.AutoMigrate(&models.JobOrder{}, &models.JobOrderAt{})

	// DB.AutoMigrate(&models.CalendarScheduleModel{}, &models.CalendarScheduleAt{},
	// 	&models.DeliveryReceiptModel{}, &models.DeliveryReceiptAt{}, &models.DeliveryReceiptAt{}, &models.TripCostModel{},
	// 	&models.TripCostContentAt{}, &models.ItemReleaseModel{}, &models.ItemReleaseAt{},
	// 	&models.ReceiptFileModel{}, &models.ReceiptFileAt{},
	// )

}
