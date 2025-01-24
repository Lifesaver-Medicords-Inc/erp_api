package initializers

import "github.com/pierceperado/smpc/models"

func MigrateDb() {
	// Auth
	DB.AutoMigrate(&models.User{}, &models.UserAt{})

	// Setup
	DB.AutoMigrate(&models.Brand{}, &models.BrandAt{})
	DB.AutoMigrate(&models.UnitMeasurement{}, &models.UnitMeasurementAt{})
	DB.AutoMigrate(&models.PaymentTerms{}, &models.PaymentTermsAt{})
	DB.AutoMigrate(&models.Class{}, &models.ClassAt{})
	DB.AutoMigrate(&models.Name{}, &models.NameAt{})
	DB.AutoMigrate(&models.Type{}, &models.TypeAt{})
	DB.AutoMigrate(&models.Item{}, &models.ItemAt{})
	DB.AutoMigrate(&models.Model{}, &models.ModelAt{})

	// Sales
	DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})
}
