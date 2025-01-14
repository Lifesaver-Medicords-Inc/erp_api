package initializers

import "github.com/pierceperado/smpc/models"

func MigrateDb() {
	DB.AutoMigrate(&models.User{}, &models.UserAt{})

	// SETUP
	DB.AutoMigrate(&models.Brand{}, &models.BrandAt{})
}
