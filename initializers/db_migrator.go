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
	DB.AutoMigrate(&models.ItemSpecs{}, &models.ItemSpecsAt{})
	DB.AutoMigrate(&models.Model{}, &models.ModelAt{})

	DB.AutoMigrate(&models.Social{}, &models.SocialAt{})
	DB.AutoMigrate(&models.Entity{}, &models.EntityAt{})

	// Sales
	DB.AutoMigrate(&models.SalesQuotation{}, &models.SalesQuotationAt{})

	DB.AutoMigrate(&models.User{}, &models.UserAt{})

	// Sample
	DB.AutoMigrate(&models.Parent{}, &models.ParentAt{})
	DB.AutoMigrate(&models.Childf{}, &models.ChildfAt{})
	DB.AutoMigrate(&models.Childs{}, &models.ChildsAt{})

	//BPI
	DB.AutoMigrate(&models.Bpi{}, &models.BpiAt{})
	DB.AutoMigrate(&models.BpiGeneral{}, &models.BpiGeneralAt{})
	DB.AutoMigrate(&models.BpiContacts{}, &models.BpiContactsAt{})
	DB.AutoMigrate(&models.BpiIndustries{}, &models.BpiIndustriesAt{})
}
