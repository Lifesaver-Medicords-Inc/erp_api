package initializers

import "github.com/pierceperado/smpc/models"

func MigrateDb() {
	// Auth

	//	DB.AutoMigrate(&models.User{}, &models.UserAt{})
	// Setup

	// DB.AutoMigrate(&models.Brand{}, &models.BrandAt{})
	// DB.AutoMigrate(&models.UnitMeasurement{}, &models.UnitMeasurementAt{})
	// DB.AutoMigrate(&models.PaymentTerms{}, &models.PaymentTermsAt{})

	// DB.AutoMigrate(&models.Class{}, &models.ClassAt{})
	// DB.AutoMigrate(&models.Name{}, &models.NameAt{})
	// DB.AutoMigrate(&models.Type{}, &models.TypeAt{})
	// DB.AutoMigrate(&models.Item{}, &models.ItemAt{})
	// DB.AutoMigrate(&models.ItemSpecs{}, &models.ItemSpecsAt{})
	// DB.AutoMigrate(&models.AdditionalSpecs{}, &models.AdditionalSpecsAt{})
	// DB.AutoMigrate(&models.TradeType{}, models.TradeTypeAt{})
	// DB.AutoMigrate(&models.Model{}, &models.ModelAt{})

	// DB.AutoMigrate(&models.Social{}, &models.SocialAt{})
	// DB.AutoMigrate(&models.Entity{}, &models.EntityAt{})

	// DB.AutoMigrate(&models.User{}, &models.UserAt{})

	// Setup

	// DB.AutoMigrate(&models.Brand{}, &models.BrandAt{})
	// DB.AutoMigrate(&models.UnitMeasurement{}, &models.UnitMeasurementAt{})
	// DB.AutoMigrate(&models.PaymentTerms{}, &models.PaymentTermsAt{})

	// DB.AutoMigrate(&models.Class{}, &models.ClassAt{})
	// DB.AutoMigrate(&models.Name{}, &models.NameAt{})
	// DB.AutoMigrate(&models.Type{}, &models.TypeAt{})
	// DB.AutoMigrate(&models.Item{}, &models.ItemAt{})
	// DB.AutoMigrate(&models.Model{}, &models.ModelAt{})

	//	DB.AutoMigrate(&models.Application{}, &models.ApplicationAt{})

	// // Sales
	// DB.AutoMigrate(&models.Order{}, &models.OrderAt{})
	// DB.AutoMigrate(&models.OrderDetails{}, &models.OrderDetailsAt{})
	//DB.AutoMigrate(&models.Opportunity{}, &models.OpportunityAt{})
	// // DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// // DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})
	// DB.AutoMigrate(&models.Order{}, &models.OrderAt{})

	// DB.AutoMigrate(&models.OrderDetails{}, &models.OrderDetailsAt{})
	// DB.AutoMigrate(&models.Opportunity{}, &models.OpportunityAt{})
	// DB.AutoMigrate(&models.Status{}, &models.StatusAt{})
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})

	// Purchasing
	//DB.AutoMigrate(&models.PurchaseRequisition{}, &models.PurchaseRequisitionAt{})
	//DB.AutoMigrate(&models.PROrders{}, &models.PROrdersAt{})

	// Sales
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	//DB.AutoMigrate(&models.Order{}, &models.OrderAt{})
	//DB.AutoMigrate(&models.OrderDetails{}, &models.OrderDetailsAt{})
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})

	//DB.AutoMigrate(&models.SalesProjectItems{}, &models.SalesProjectItemsAt{})

	//BPI
	// DB.AutoMigrate(&models.Bpi{}, &models.BpiAt{})
	// DB.AutoMigrate(&models.BpiGeneral{}, &models.BpiGeneralAt{})
	DB.AutoMigrate(&models.BpiContacts{}, &models.BpiContactsAt{})
	// DB.AutoMigrate(&models.BpiIndustries{}, &models.BpiIndustriesAt{})
	// DB.AutoMigrate(&models.BpiBranchIndustries{}, &models.BpiBranchIndustriesAt{})
	// DB.AutoMigrate(&models.BpiEntity{}, &models.BpiEntityAt{})
	// DB.AutoMigrate(&models.BpiAddress{}, &models.BpiAddressAt{})
	// DB.AutoMigrate(&models.BpiItems{}, &models.BpiItemsAt{})
	// DB.AutoMigrate(&models.BpiFinance{}, &models.BpiFinanceAt{})
	// DB.AutoMigrate(&models.BpiAccreditation{}, &models.BpiAccreditationAt{})

}
