package initializers

import "github.com/pierceperado/smpc/models"

func MigrateDb() {
	DB.AutoMigrate(&models.User{}, &models.UserAt{})

	// SETUP
	DB.AutoMigrate(&models.Brand{}, &models.BrandAt{})
	DB.AutoMigrate(&models.UnitMeasurement{}, &models.UnitMeasurementAt{})
	DB.AutoMigrate(&models.PaymentTerms{}, &models.PaymentTermsAt{})
	DB.AutoMigrate(&models.Class{}, &models.ClassAt{})
	DB.AutoMigrate(&models.Name{}, &models.NameAt{})
	DB.AutoMigrate(&models.Type{}, &models.TypeAt{})
}
