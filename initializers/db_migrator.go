package initializers

import "github.com/pierceperado/smpc/models"

func MigrateDb() {
	// Auth
	DB.AutoMigrate(&models.User{}, &models.UserAt{})

	// Setup
	// DB.AutoMigrate(&models.Brand{}, &models.BrandAt{})
	// DB.AutoMigrate(&models.UnitMeasurement{}, &models.UnitMeasurementAt{})
	// DB.AutoMigrate(&models.PaymentTerms{}, &models.PaymentTermsAt{})
	// DB.AutoMigrate(&models.Class{}, &models.ClassAt{})
	// DB.AutoMigrate(&models.Name{}, &models.NameAt{})
	// DB.AutoMigrate(&models.Type{}, &models.TypeAt{})
	// DB.AutoMigrate(&models.Item{}, &models.ItemAt{})
	// DB.AutoMigrate(&models.Model{}, &models.ModelAt{})
	// DB.AutoMigrate(&models.Application{}, &models.ApplicationAt{})

	// Sales
	DB.AutoMigrate(&models.Order{}, &models.OrderAt{})
	DB.AutoMigrate(&models.OrderDetails{}, &models.OrderDetailsAt{})
	// DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
	// DB.AutoMigrate(&models.SalesQuotationQuick{}, &models.SalesQuotationQuickAt{})
}
